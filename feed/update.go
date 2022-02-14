package feed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// EventLister lists events using an optional filter.
type EventLister interface {
	// ListUniqueEvents lists events. If ids is non-empty, it is used to filter the
	// result.
	ListUniqueEvents(ctx context.Context, ids []uuid.UUID) ([]Event, error)
}

// RSSAdapter makes it possible to use the RSS Events functions as an
// EventLister.
type RSSAdapter struct {
	// List of region IDs
	regionIDs []string
	read      func(ctx context.Context, regionIDs []string) ([]Event, error)
}

func NewRSSAdapter(
	regionIDs []string,
	readFn func(ctx context.Context, regionIDs []string) ([]Event, error),
) *RSSAdapter {
	if readFn == nil {
		readFn = EventsFromRSS
	}
	return &RSSAdapter{
		regionIDs: regionIDs,
		read:      readFn,
	}
}

// ListUniqueEvents lists events found in the RSS feeds for the given regions
func (a *RSSAdapter) ListUniqueEvents(ctx context.Context, ids []uuid.UUID) ([]Event, error) {
	if len(ids) != 0 {
		return nil, errors.New("RSS feed cannot filter by ID")
	}
	events, err := a.read(ctx, a.regionIDs)
	if err != nil {
		return nil, err
	}

	// De-duplicate - prefer more recent
	sort.Slice(events, func(i, j int) bool {
		if events[i].ID == events[j].ID {
			return events[i].PublishTime.After(events[j].PublishTime)
		}
		return bytes.Compare(events[i].ID[:], events[j].ID[:]) < 0
	})
	var j int
	for i := 0; i < len(events); i++ {
		if i > 0 && events[i].ID == events[i-1].ID {
			continue
		}
		events[j] = events[i]
		j++
	}
	events = events[:j]

	return events, nil
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

type Updater struct {
	toCreate     map[uuid.UUID]Event
	toCreateList []Event
	ids          []uuid.UUID
	mtx          sync.Mutex
}

func NewUpdater() *Updater {
	return &Updater{
		toCreate:     make(map[uuid.UUID]Event, 1000),
		toCreateList: make([]Event, 1000),
		ids:          make([]uuid.UUID, 1000),
	}
}

// Update updates the provided database with records from the Police RSS feed
// on an interval given by refreshTime.
func (u *Updater) Update(
	ctx context.Context,
	rss EventLister,
	target EventListerCreator,
) error {
	u.mtx.Lock()
	defer u.mtx.Unlock()
	// Cleanup
	u.ids = u.ids[:0]
	u.toCreateList = u.toCreateList[:0]
	for k := range u.toCreate {
		delete(u.toCreate, k)
	}

	// Fetch RSS events
	rssEvents, err := rss.ListUniqueEvents(ctx, nil)
	if err != nil {
		return fmt.Errorf("update events err, %w", err)
	}

	// Gather ids for Events
	for _, evt := range rssEvents {
		u.ids = append(u.ids, evt.ID)
	}

	// Fetch existing events
	targetEvents, err := target.ListUniqueEvents(ctx, u.ids)
	if err != nil {
		return fmt.Errorf("fetch existing events err, %w", err)
	}

	// Add existing events to map
	for _, evt := range targetEvents {
		u.toCreate[evt.ID] = evt
	}
	for _, evt := range rssEvents {
		if v, exists := u.toCreate[evt.ID]; exists {
			if bytes.Equal(v.ContentHash, evt.ContentHash) {
				delete(u.toCreate, evt.ID) // no update needed
				continue
			}
			evt.Revision = v.Revision + 1
			u.toCreate[evt.ID] = evt
		} else { // Create new event
			evt.Revision = 1
			u.toCreate[evt.ID] = evt
		}
	}

	// Create events
	log.Printf("Creating %d records...\n", len(u.toCreate))
	for _, evt := range u.toCreate {
		u.toCreateList = append(u.toCreateList, evt)
	}
	if err := target.CreateEvents(ctx, u.toCreateList); err != nil {
		return fmt.Errorf("failed to create new events, %w", err)
	}

	return nil
}
