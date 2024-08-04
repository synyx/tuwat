package connectors

import (
	"context"
	"time"
)

type SilencerFunc func(ctx context.Context, duration time.Duration, user string) error
