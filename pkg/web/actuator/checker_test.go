package actuator

import (
	"context"
	"testing"
	"time"

	"github.com/synyx/tuwat/pkg/clock"
)

func TestNewHealthAccumulator(t *testing.T) {
	mockClock := clock.NewMockClock(time.Unix(0, 0))
	acc := NewHealthAccumulator(mockClock)

	acc.Register("test", testCheck)

	ctx, cancel := context.WithCancel(context.Background())

	go acc.Run(ctx)
	mockClock.Progress(checkInterval * 2)
	time.Sleep(20 * time.Millisecond)
	cancel()
}

func testCheck(ctx context.Context) (status Status, message string) {
	return Up, "test"
}
