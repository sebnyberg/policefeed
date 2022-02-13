package feed

import (
	"context"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel RSSChannel
}

type RSSChannel struct {
	XMLName     xml.Name      `xml:"channel"`
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Description string        `xml:"description"`
	Items       []RSSFeedItem `xml:"item"`
}

type RSSFeedItem struct {
	XMLName     xml.Name `xml:"item"`
	Guid        string   `xml:"guid"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	PubDateStr  string   `xml:"pubDate"`
	Link        string   `xml:"link"`
}

const rssBaseURL = "https://polisen.se/aktuellt/rss/%v/handelser-rss---%v/"

func EventsFromRSS(ctx context.Context, regionIDs []string) ([]Event, error) {
	events := make(chan []Event)
	if len(regionIDs) == 1 && regionIDs[0] == "" {
		regionIDs = keys(rssRegions)
	}

	doRegion := func(regionCtx context.Context, regionID string) func() error {
		return func() error {
			// Validate region ID
			if _, exists := rssRegions[regionID]; !exists {
				return fmt.Errorf(
					"unknown region %v, choose one or more of %v",
					regionID, strings.Join(keys(rssRegions), ","),
				)
			}

			// Create RSS URL
			url := fmt.Sprintf(rssBaseURL, regionID, regionID)
			if regionID == "jonkoping" {
				url = fmt.Sprintf(rssBaseURL, "jonkopings-lan", "jonkoping")
			}

			// Make request
			req, err := http.NewRequestWithContext(regionCtx, http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("create request, %w", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("send request, %w", err)
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf("unexpected response code %v", resp.StatusCode)
			}

			// Parse response
			parsedEvents, err := eventsFromRSSBody(resp.Body)
			if err != nil {
				return fmt.Errorf("parse events err, %w", err)
			}
			select {
			case <-regionCtx.Done():
			case events <- parsedEvents:
			}
			return nil
		}
	}

	// For each region, collect events from the RSS feed and put into events chan
	g, gCtx := errgroup.WithContext(ctx)
	for _, regionID := range regionIDs {
		g.Go(doRegion(gCtx, regionID))
	}

	// Collect events from events chan
	result := make([]Event, 0, 2000)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for eventBatch := range events {
			result = append(result, eventBatch...)
		}
	}()

	// Wait for producers to finish
	if err := g.Wait(); err != nil {
		return nil, err
	}
	close(events)
	<-done

	return result, nil
}

func eventsFromRSSBody(r io.ReadCloser) ([]Event, error) {
	var feed RSS
	if err := xml.NewDecoder(r).Decode(&feed); err != nil {
		return nil, err
	}
	events := make([]Event, len(feed.Channel.Items))
	for i, item := range feed.Channel.Items {
		publishTime, err := time.Parse(time.RFC1123Z, item.PubDateStr)
		if err != nil {
			return nil, fmt.Errorf("parse publish time, %w", err)
		}
		h := sha256.New()
		if _, err := h.Write([]byte(item.Title)); err != nil {
			return nil, fmt.Errorf("content hash, %w", err)
		}
		if _, err := h.Write([]byte(item.Description)); err != nil {
			return nil, fmt.Errorf("content hash, %w", err)
		}
		contentHash := h.Sum(nil)
		events[i] = Event{
			ID:          NewEventID(item.Guid),
			URL:         item.Guid,
			Title:       item.Title,
			Region:      feed.Channel.Title,
			Description: item.Description,
			CreateTime:  time.Now(),
			PublishTime: publishTime,
			ContentHash: contentHash,
		}
	}
	return events, nil
}
