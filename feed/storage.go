package feed

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/sebnyberg/policefeed/feed/feedpg"
)

var _ EventLister = new(EventStorage)

var _ EventCreator = new(EventStorage)

type EventStorage struct {
	db      *sql.DB
	queries *feedpg.Queries
}

func NewEventStorage(db *sql.DB) *EventStorage {
	return &EventStorage{
		db:      db,
		queries: feedpg.New(db),
	}
}

func (s *EventStorage) ListEvents(
	ctx context.Context,
	ids []uuid.UUID,
) ([]Event, error) {
	dbEvents, err := s.queries.ListRecentEvents(ctx, ids)
	if err != nil {
		return nil, err
	}
	events := make([]Event, len(dbEvents))
	for i, dbEvent := range dbEvents {
		events[i] = Event{
			ID:          dbEvent.ID,
			URL:         dbEvent.Url,
			Title:       dbEvent.Title,
			Region:      dbEvent.Region,
			Description: dbEvent.Description,
			Revision:    dbEvent.Revision,
			PublishTime: dbEvent.PublishTime,
			ContentHash: dbEvent.ContentHash,
		}
	}
	return events, nil
}

func (s *EventStorage) CreateEvents(
	ctx context.Context, events []Event,
) (retErr error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("unexpected error opening conn, %w", err)
	}
	defer func() {
		err := conn.Close()
		if retErr == nil {
			retErr = err
		}
	}()

	return conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()
		rows := make([][]interface{}, len(events))
		for i, evt := range events {
			rows[i] = []interface{}{
				evt.ID,
				evt.URL,
				evt.Title,
				evt.Region,
				evt.Description,
				evt.PublishTime,
				evt.CreateTime,
				evt.ContentHash,
				evt.Revision,
			}
		}
		_, err = conn.CopyFrom(ctx,
			pgx.Identifier{"police_event"},
			[]string{
				"id",
				"url",
				"title",
				"region",
				"description",
				"publish_time",
				"create_time",
				"content_hash",
				"revision",
			},
			pgx.CopyFromRows(rows),
		)
		return err
	})
}
