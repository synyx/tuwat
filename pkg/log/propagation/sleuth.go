package propagation

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	sleuthTraceHeader   = "spanTraceId"
	sleuthSpanHeader    = "spanId"
	sleuthSampledHeader = "spanSampled"
)

type SleuthTraceContext struct{}

var _ propagation.TextMapPropagator = SleuthTraceContext{}

func (tc SleuthTraceContext) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}

	// Clear all flags other than the trace-context supported sampling bit.
	flags := sc.TraceFlags() & trace.FlagsSampled

	h := strings.Join([]string{sc.TraceID().String(), sc.SpanID().String(), flags.String()}, "-")
	carrier.Set(b3Header, h)
}

func (tc SleuthTraceContext) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sc := tc.extractSleuth(carrier)
	if !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func (tc SleuthTraceContext) Fields() []string {
	return []string{sleuthTraceHeader, sleuthSpanHeader, sleuthSampledHeader}
}

func (tc SleuthTraceContext) extractSleuth(carrier propagation.TextMapCarrier) (spanContext trace.SpanContext) {
	var err error
	var spanConfig trace.SpanContextConfig

	if field := carrier.Get(sleuthTraceHeader); field == "" {
		return
	} else if spanConfig.TraceID, err = trace.TraceIDFromHex(fixB3TID(field)); err != nil {
		return
	}

	if field := carrier.Get(sleuthSpanHeader); field == "" {
		return
	} else if spanConfig.SpanID, err = trace.SpanIDFromHex(field); err != nil {
		return
	}

	if field := carrier.Get(sleuthSampledHeader); field == "1" {
		spanConfig.TraceFlags |= trace.FlagsSampled
	}

	return trace.NewSpanContext(spanConfig)
}
