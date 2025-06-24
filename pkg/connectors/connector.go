package connectors

import (
	"context"
)

type Connector interface {
	Collect(ctx context.Context) ([]Alert, error)
	String() string
	Tag() string
}
