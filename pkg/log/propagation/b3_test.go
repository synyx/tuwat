package propagation

import (
	"context"
	"strings"
	"testing"

	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel/trace"
)

func TestExtractTracestate_B3(t *testing.T) {
	table := make(amqp.Table)
	table["b3"] = "2de8dcbd7b05b4bb-91a804b4e6d8af01-0"

	ctx := B3TraceContext{}.Extract(context.Background(), HeaderCarrier(table))

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().TraceID().IsValid() || !strings.HasSuffix(span.SpanContext().TraceID().String(), "2de8dcbd7b05b4bb") {
		t.Error()
	}
	if !span.SpanContext().SpanID().IsValid() || span.SpanContext().SpanID().String() != "91a804b4e6d8af01" {
		t.Error()
	}
}
