package errs_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/l10n-center/api/src/errs"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestZapStack(t *testing.T) {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	err := errors.New("test")
	sterr := err.(stackTracer)

	st := strings.TrimSpace(
		fmt.Sprintf("%+v", sterr.StackTrace()),
	)

	stf := errs.ZapStack(err)
	assert.Equal(t, "stacktrace", stf.Key)
	assert.Equal(t, st, stf.String)
}
