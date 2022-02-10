package policefeed

import (
	"time"
)

type Event struct {
	ID          string
	Title       string
	Region      string
	Description string
	PublishTime time.Time

	// Todo: add geometries
	// EventGeometryRetryTime time.Time // next time to try fetch event geometry
	// EventGeometry  geom.T
}
