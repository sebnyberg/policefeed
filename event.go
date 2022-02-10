package policefeed

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID
	URL         string
	Title       string
	Region      string
	Description string
	PublishTime time.Time

	// Todo: add geometries
	// EventGeometryRetryTime time.Time // next time to try fetch event geometry
	// EventGeometry  geom.T
}

var eventIDNamespace = uuid.NewSHA1(uuid.NameSpaceDNS, []byte("policefeed.v1.PoliceEvent.ID"))

func NewEventID(URL string) uuid.UUID {
	return uuid.NewSHA1(eventIDNamespace, []byte(URL))
}
