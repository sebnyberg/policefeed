package feed_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/sebnyberg/policefeed/feed"
	"github.com/sebnyberg/policefeed/feed/feedfakes"
	"github.com/stretchr/testify/require"
)

func cast[T1, T2 any](items []T1, mapper func(T1) T2) []T2 {
	res := make([]T2, len(items))
	for i, item := range items {
		res[i] = mapper(item)
	}
	return res
}

func TestUpdate(t *testing.T) {
	type simpleEvent struct {
		id    byte
		t     time.Time
		title string
	}
	asEvent := func(e simpleEvent) feed.Event {
		var id [16]byte
		id[0] = e.id
		return feed.Event{
			ID:          id,
			PublishTime: e.t,
			Title:       e.title,
			ContentHash: []byte(e.title),
		}
	}
	asSimpleEvent := func(e feed.Event) simpleEvent {
		return simpleEvent{
			id:    e.ID[0],
			t:     e.PublishTime,
			title: e.Title,
		}
	}
	oldT := time.Time{}
	newT := oldT.Add(time.Second)
	for _, tc := range []struct {
		name            string
		newEvents       []simpleEvent
		sourceErr       error
		oldEvents       []simpleEvent
		targetListErr   error
		targetCreateErr error
		want            []simpleEvent
		wantErr         error
	}{
		{
			name: "happy path",
			newEvents: []simpleEvent{
				{1, newT, "new1"}, // duplicate entry - newest wins
				{1, oldT, "old1"}, // duplicate entry - old is discarded
				{2, newT, "new2"}, // unique entry
				{3, newT, "old3"}, // old entry with same hash does not get updated
				{4, newT, "new4"}, // entry already exists, new hash
			},
			sourceErr: nil,
			oldEvents: []simpleEvent{
				{3, oldT, "old3"}, // old entry with same hash does not get created
				{4, oldT, "old4"}, // old entry with old hash
			},
			targetListErr:   nil,
			targetCreateErr: nil,
			want: []simpleEvent{
				{1, newT, "new1"}, // duplicate entry - newest
				{2, newT, "new2"}, // unique entry
				{4, newT, "new4"}, // new hash replaces old
			},
			wantErr: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := feed.NewRSSAdapter([]string{},
				func(ctx context.Context, regionIDs []string) ([]feed.Event, error) {
					return cast(tc.newEvents, asEvent), tc.sourceErr
				},
			)
			target := new(feedfakes.FakeEventListerCreator)
			target.ListUniqueEventsReturns(cast(tc.oldEvents, asEvent), tc.targetListErr)
			target.CreateEventsReturns(tc.targetCreateErr)
			up := feed.NewUpdater()
			gotErr := up.Update(context.Background(), source, target)

			// Assert
			if tc.wantErr != nil {
				require.ErrorIs(t, gotErr, tc.wantErr)
				return
			}

			require.NotEmpty(t, target.Invocations()["CreateEvents"])
			got := cast(target.Invocations()["CreateEvents"][0][1].([]feed.Event), asSimpleEvent)
			if len(got) != len(tc.want) {
				t.Fatalf("wanted %v events, got %v", len(tc.want), len(got))
			}
			sort.Slice(tc.want, func(i, j int) bool {
				return tc.want[i].id < tc.want[j].id
			})
			sort.Slice(got, func(i, j int) bool {
				return got[i].id < got[j].id
			})
			for i := range tc.want {
				require.Equal(t, tc.want[i], got[i])
			}
		})
	}
}
