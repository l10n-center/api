package tracing

import (
	"context"
)

type traceIDCtxKey struct{}

// ContextWithTraceID save trace id in context
func ContextWithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDCtxKey{}, id)
}

// TraceIDFromContext extract trace id if span started or empty string otherwise
func TraceIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDCtxKey{}).(string); ok {
		return id
	}

	return ""
}
