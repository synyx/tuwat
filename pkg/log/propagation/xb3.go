package propagation

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	xb3TraceHeader   = "X-B3-TraceId"
	xb3SpanHeader    = "X-B3-SpanId"
	xb3SampledHeader = "X-B3-Sampled"
)

type XB3TraceContext struct{}

var _ propagation.TextMapPropagator = XB3TraceContext{}

func (tc XB3TraceContext) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}

	// Clear all flags other than the trace-context supported sampling bit.
	flags := sc.TraceFlags() & trace.FlagsSampled

	h := strings.Join([]string{sc.TraceID().String(), sc.SpanID().String(), flags.String()}, "-")
	carrier.Set(b3Header, h)
}

func (tc XB3TraceContext) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sc := tc.extractXB3(carrier)
	if !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func (tc XB3TraceContext) Fields() []string {
	return []string{xb3TraceHeader, xb3SpanHeader, xb3SampledHeader}
}

func (tc XB3TraceContext) extractXB3(carrier propagation.TextMapCarrier) (spanContext trace.SpanContext) {
	var err error
	var spanConfig trace.SpanContextConfig

	if field := carrier.Get(xb3TraceHeader); field == "" {
		return
	} else if spanConfig.TraceID, err = trace.TraceIDFromHex(fixB3TID(field)); err != nil {
		return
	}

	if field := carrier.Get(xb3SpanHeader); field == "" {
		return
	} else if spanConfig.SpanID, err = trace.SpanIDFromHex(field); err != nil {
		return
	}

	if field := carrier.Get(xb3SampledHeader); field == "1" {
		spanConfig.TraceFlags |= trace.FlagsSampled
	}

	return trace.NewSpanContext(spanConfig)
}
