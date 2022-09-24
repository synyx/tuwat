package propagation

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	sleuthCurrentTraceHeader   = "currentSpan.Trace"
	sleuthCurrentSpanHeader    = "currentSpan.Span"
	sleuthCurrentSampledHeader = "currentSpan.spanSampled"
)

type SleuthCurrentTraceContext struct{}

var _ propagation.TextMapPropagator = SleuthCurrentTraceContext{}

func (tc SleuthCurrentTraceContext) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return
	}

	// Clear all flags other than the trace-context supported sampling bit.
	flags := sc.TraceFlags() & trace.FlagsSampled

	h := strings.Join([]string{sc.TraceID().String(), sc.SpanID().String(), flags.String()}, "-")
	carrier.Set(b3Header, h)
}

func (tc SleuthCurrentTraceContext) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	sc := tc.extractSleuthCurrent(carrier)
	if !sc.IsValid() {
		return ctx
	}
	return trace.ContextWithRemoteSpanContext(ctx, sc)
}

func (tc SleuthCurrentTraceContext) Fields() []string {
	return []string{sleuthCurrentTraceHeader, sleuthCurrentSpanHeader, sleuthCurrentSampledHeader}
}

func (tc SleuthCurrentTraceContext) extractSleuthCurrent(carrier propagation.TextMapCarrier) (spanContext trace.SpanContext) {
	var err error
	var spanConfig trace.SpanContextConfig

	if field := carrier.Get(sleuthCurrentTraceHeader); field == "" {
		return
	} else if spanConfig.TraceID, err = trace.TraceIDFromHex(fixB3TID(field)); err != nil {
		return
	}

	if field := carrier.Get(sleuthCurrentSpanHeader); field == "" {
		return
	} else if spanConfig.SpanID, err = trace.SpanIDFromHex(field); err != nil {
		return
	}

	if field := carrier.Get(sleuthCurrentSampledHeader); field == "1" {
		spanConfig.TraceFlags |= trace.FlagsSampled
	}

	return trace.NewSpanContext(spanConfig)
}
