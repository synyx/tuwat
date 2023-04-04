package propagation

import (
	"context"
	"strings"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

func TestExtractTracestate_XB3(t *testing.T) {
	table := make(amqp.Table)
	table["X-B3-TraceId"] = "2de8dcbd7b05b4bb"
	table["X-B3-SpanId"] = "91a804b4e6d8af01"
	table["X-B3-Sampled"] = true
	table["X-B3-ParentSpanId"] = "-"

	ctx := XB3TraceContext{}.Extract(context.Background(), HeaderCarrier(table))

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().TraceID().IsValid() || !strings.HasSuffix(span.SpanContext().TraceID().String(), "2de8dcbd7b05b4bb") {
		t.Error()
	}
	if !span.SpanContext().SpanID().IsValid() || span.SpanContext().SpanID().String() != "91a804b4e6d8af01" {
		t.Error()
	}
}
