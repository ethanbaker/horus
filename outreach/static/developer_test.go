package static_test

import (
	"testing"

	"github.com/ethanbaker/horus/outreach/static"
	"github.com/ethanbaker/horus/utils/config"
	"github.com/stretchr/testify/assert"
)

/* ---- MESSAGE TESTS ---- */

func TestPing(t *testing.T) {
	assert := assert.New(t)

	// Initialize the environment (there should be one error for SQL, as we don't need it)
	config, errs := config.New()
	assert.Equal(0, len(errs))

	// Run the test
	output := static.Ping(config, map[string]any{})
	assert.Equal("Ping!", output)
}
