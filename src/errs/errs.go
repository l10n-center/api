package errs

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const stacktraceName = "stacktrace"

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ZapStack try get stacktrace from error and convert it in zapcore.Field
func ZapStack(err error) zapcore.Field {
	if e, ok := err.(stackTracer); ok {
		return zap.String(
			stacktraceName,
			strings.TrimSpace(
				fmt.Sprintf("%+v", e.StackTrace()),
			),
		)
	}
	return zap.Stack(stacktraceName)
}
