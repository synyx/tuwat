package actuator

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestNewHealthAccumulator(t *testing.T) {
	mockClock := clock.NewMock()
	acc := NewHealthAccumulator(mockClock)

	acc.Register("test", testCheck)

	ctx, cancel := context.WithCancel(context.Background())

	go acc.Run(ctx)
	mockClock.Add(checkInterval * 2)
	time.Sleep(20 * time.Millisecond)
	cancel()
}

func testCheck(ctx context.Context) (status Status, message string) {
	return Up, "test"
}
