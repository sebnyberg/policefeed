package feed_test

import (
	"context"
	"testing"
	"time"

	"github.com/sebnyberg/policefeed/feed"
	"github.com/sebnyberg/policefeed/feed/feedfakes"
	"github.com/stretchr/testify/require"
)

func TestDuplicateRSSEvent(t *testing.T) {
	// Test whether duplicate RSS ids are handled correctly by the updater.
	// Only the most recent publishTime should be used.
	target := new(feedfakes.FakeEventListerCreator)
	inputs := []feed.Event{
		{
			ID:          [16]byte{},
			PublishTime: time.Time{}.Add(time.Second * 1),
			Title:       "new",
		},
		{
			ID:          [16]byte{},
			PublishTime: time.Time{},
			Title:       "old",
		},
	}
	source := feed.NewRSSAdapter([]string{},
		func(ctx context.Context, regionIDs []string) ([]feed.Event, error) {
			return inputs, nil
		},
	)
	target.ListUniqueEventsReturns(nil, nil)
	up := feed.NewUpdater()
	up.Update(context.Background(), source, target)

	require.NotEmpty(t, target.Invocations()["CreateEvents"])
	args := target.Invocations()["CreateEvents"]
	require.NotEmpty(t, args[0][1])
	createInputEvents := args[0][1].([]feed.Event)
	require.Len(t, createInputEvents, 1)
	require.Equal(t, "new", createInputEvents[0].Title)
}
