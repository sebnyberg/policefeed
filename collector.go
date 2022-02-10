package policefeed

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type EventCollectorConfig struct {
}

type EventCollector struct {
	running bool

	refreshInterval time.Duration
	sendTimeout     time.Duration
	events          chan Event

	err       error
	errCtx    context.Context
	errCancel context.CancelFunc
	errOnce   sync.Once

	subMtx sync.RWMutex
	subs   []chan Event

	regions       Regions
	activeRegions []string
	regionSeen    []map[string]struct{}
}

func NewCollector(
	regionIDs []string,
	rssBaseURLTemplate string,
	sendTimeout, refreshInterval time.Duration,
) (*EventCollector, error) {
	var c EventCollector
	c.errCtx, c.errCancel = context.WithCancel(context.Background())
	c.events = make(chan Event)
	c.sendTimeout = sendTimeout
	c.refreshInterval = refreshInterval
	c.regions = NewRegions(rssBaseURLTemplate)
	if len(regionIDs) == 1 && regionIDs[0] == "all" {
		regionIDs = c.regions.ListIDs()
	}
	for _, regionID := range regionIDs {
		// Validate
		if !c.regions.Exists(regionID) {
			return nil, fmt.Errorf(
				"unknown region %v, choose one or more of %v",
				regionID, strings.Join(c.regions.ListIDs(), ","),
			)
		}
		// Add to list of regions
		c.activeRegions = append(c.activeRegions, regionID)
		c.regionSeen = append(c.regionSeen, make(map[string]struct{}, 1000))
	}

	return &c, nil
}

func (c *EventCollector) run() {
	if c.running {
		return
	}
	c.running = true

	// Start a worker that reads RSS feeds
	for i, region := range c.activeRegions {
		go func(region string, idx int) {
			if err := c.readRSSFeed(region, idx); err != nil {
				c.errOnce.Do(func() {
					c.err = err
					c.errCancel()
					close(c.events)
				})
			}
		}(region, i)
	}

	// Start a worker that consumes events on the event channel
	go func() {
		sendToSubs := func(event Event) error {
			c.subMtx.RLock()
			for _, sub := range c.subs {
				c.subMtx.RUnlock()
				select {
				case sub <- event:
				case <-c.errCtx.Done():
					return c.err
				}
				c.subMtx.RLock()
			}
			c.subMtx.RUnlock()
			return nil
		}

		for event := range c.events {
			if err := sendToSubs(event); err != nil {
				return
			}
		}
	}()

	// Done!
}

// readRSSFeed reads the RSS feed, putting any unseen events into the events
// channel.
func (c *EventCollector) readRSSFeed(region string, idx int) error {
	url := c.regions.GetRSSURL(region)
	for {
		select {
		case <-c.errCtx.Done():
			return c.err
		default:
		}
		// Read from RSS feed
		res, err := http.Get(url)
		if err != nil {
			fmt.Println(url)
			return fmt.Errorf("get rss feed err, %w", err)
		}
		if res.StatusCode != http.StatusOK {
			fmt.Println(url)
			return fmt.Errorf("read rss feed fetch got non-200 status code, was %v", res.StatusCode)
		}
		events, err := EventsFromRSS(res.Body)
		if err != nil {
			return fmt.Errorf("parse rss feed err, %v", res.StatusCode)
		}
		for _, event := range events {
			if _, exists := c.regionSeen[idx][event.ID]; !exists {
				select {
				case <-c.errCtx.Done():
					return c.err
				case c.events <- event:
				case <-time.After(c.sendTimeout):
					return fmt.Errorf("event channel receive timeout %v", c.sendTimeout)
				}
				c.regionSeen[idx][event.ID] = struct{}{}
			}
		}
		select {
		case <-time.After(c.refreshInterval):
		case <-c.errCtx.Done():
		}
	}
}

func (c *EventCollector) Subscribe(
	ctx context.Context,
) (func() (Event, error), error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.errCtx.Done():
		return nil, c.err
	default:
	}

	// Add subscriber
	c.subMtx.Lock()
	defer c.subMtx.Unlock()
	sub := make(chan Event)
	c.subs = append(c.subs, sub)
	defer c.run()

	go func() {
		<-ctx.Done()
		close(sub)
	}()

	return func() (Event, error) {
		select {
		case <-ctx.Done():
			return Event{}, ctx.Err()
		case <-c.errCtx.Done():
			return Event{}, c.err
		case evt := <-sub:
			return evt, nil
		}
	}, nil
}
