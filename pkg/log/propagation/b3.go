package propagation

import (
	"context"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const b3Header = "b3"

type B3TraceContext struct{}

var _ propagation.TextMapPropagator = B3TraceContext{}

func (tc B3TraceContext) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}

	// Clear all flags other than the trace-context supported sampling bit.
	flags := sc.TraceFlags() & trace.FlagsSampled

	h := strings.Join([]string{sc.TraceID().String(), sc.SpanID().String(), flags.String()}, "-")
	carrier.Set(b3Header, h)
}

func (tc B3TraceContext) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sc := tc.extractB3(carrier)
	if !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func (tc B3TraceContext) Fields() []string {
	return []string{b3Header}
}

func (tc B3TraceContext) extractB3(carrier propagation.TextMapCarrier) (spanContext trace.SpanContext) {
	var err error
	var spanConfig trace.SpanContextConfig

	line := carrier.Get(b3Header)

	chunks := strings.Split(line, "-")
	if len(chunks) < 2 {
		return
	}

	if spanConfig.TraceID, err = trace.TraceIDFromHex(fixB3TID(chunks[0])); err != nil {
		return
	}
	chunks = chunks[1:]

	spanConfig.SpanID, err = trace.SpanIDFromHex(chunks[0])
	if err != nil {
		return
	}
	chunks = chunks[1:]

	if len(chunks) > 0 && len(chunks[0]) == 1 {
		if i, err := strconv.ParseInt(chunks[0], 10, 64); err == nil {
			spanConfig.TraceFlags = trace.TraceFlags(i)
		}
	}

	return trace.NewSpanContext(spanConfig)
}
