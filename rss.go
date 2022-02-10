package policefeed

import (
	"encoding/xml"
	"fmt"
	"io"
	"time"
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

func EventsFromRSS(r io.ReadCloser) ([]Event, error) {
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
		events[i] = Event{
			ID:          item.Guid,
			Title:       item.Title,
			Region:      feed.Channel.Title,
			Description: item.Description,
			PublishTime: publishTime,
		}
	}
	return events, nil
}
