package feed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
)

// EventLister lists events using an optional filter.
type EventLister interface {
	// ListEvents lists events. If ids is non-empty, it is used to filter the
	// result.
	ListEvents(ctx context.Context, ids []uuid.UUID) ([]Event, error)
}

// RSSAdapter makes it possible to use the RSS Events functions as an
// EventLister.
type RSSAdapter struct {
	// List of region IDs
	RegionIDs []string
}

// ListEvents lists events found in the RSS feeds for the given regions
func (a *RSSAdapter) ListEvents(ctx context.Context, ids []uuid.UUID) ([]Event, error) {
	if len(ids) != 0 {
		return nil, errors.New("RSS feed cannot filter by ID")
	}
	return EventsFromRSS(ctx, a.RegionIDs)
}

// EventCreator creates events.
type EventCreator interface {
	// CreateEvents creates the provided events.
	CreateEvents(context.Context, []Event) error
}

// EventListerCreator supprots both creating and listing events. This is
// required by the target of the Update functionality.
type EventListerCreator interface {
	EventCreator
	EventLister
}

// Update updates the provided database with records from the Police RSS feed
// on an interval given by refreshTime.
func Update(
	ctx context.Context,
	source EventLister,
	target EventListerCreator,
	refreshTime time.Duration,
) error {
	toCreate := make(map[uuid.UUID]Event, 1000)
	toCreateList := make([]Event, 1000)
	ids := make([]uuid.UUID, 0, 1000)
	for {
		// Cleanup
		ids = ids[:0]
		toCreateList = toCreateList[:0]
		for k := range toCreate {
			delete(toCreate, k)
		}

		// Fetch RSS events
		sourceEvents, err := source.ListEvents(ctx, nil)
		if err != nil {
			return fmt.Errorf("update events err, %w", err)
		}
		// De-duplicate - prefer newer
		sort.Slice(sourceEvents, func(i, j int) bool {
			if sourceEvents[i].URL == sourceEvents[j].URL {
				return sourceEvents[i].PublishTime.After(sourceEvents[j].PublishTime)
			}
			return sourceEvents[i].URL < sourceEvents[j].URL
		})
		for i, j := 0, 0; i < len(sourceEvents); i++ {
			if i > 0 && sourceEvents[i].URL == sourceEvents[i-1].URL {
				continue
			}
			sourceEvents[j] = sourceEvents[i]
			j++
		}

		// Gather ids for Events
		for _, evt := range sourceEvents {
			ids = append(ids, evt.ID)
		}

		// Fetch existing events
		targetEvents, err := target.ListEvents(ctx, ids)
		if err != nil {
			return fmt.Errorf("fetch existing events err, %w", err)
		}

		// Add existing events to map
		for _, evt := range targetEvents {
			toCreate[evt.ID] = evt
		}
		for _, evt := range sourceEvents {
			if v, exists := toCreate[evt.ID]; exists {
				if bytes.Equal(v.ContentHash, evt.ContentHash) {
					delete(toCreate, evt.ID) // no update needed
					continue
				}
				evt.Revision = v.Revision + 1
				toCreate[evt.ID] = evt
			} else { // Create new event
				evt.Revision = 1
				toCreate[evt.ID] = evt
			}
		}

		// Create events
		log.Printf("Creating %d records...\n", len(toCreate))
		for _, evt := range toCreate {
			toCreateList = append(toCreateList, evt)
		}
		if err := target.CreateEvents(ctx, toCreateList); err != nil {
			return fmt.Errorf("failed to create new events, %w", err)
		}

		// Pace updates
		time.Sleep(refreshTime)
	}
}
