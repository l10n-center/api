package config_test

import (
	"testing"

	"github.com/l10n-center/api/src/config"
	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	assert.False(t, cfg.Debug)
	assert.Len(t, cfg.Secret, 20)
}
