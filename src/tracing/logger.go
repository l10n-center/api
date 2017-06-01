package tracing

import (
	"context"

	"go.uber.org/zap"
)

// Logger return logger with trace id
func Logger(ctx context.Context) *zap.Logger {
	l := zap.L()
	l = l.With(zap.String("traceId", TraceIDFromContext(ctx)))

	return l
}
