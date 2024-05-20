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

	// initialize the environment
	config, errs := config.NewConfigFromFile("./testing/.env")
	assert.Equal(0, len(errs))

	// Run the test
	output := static.Ping(config)
	assert.Equal("Ping!", output)
}
