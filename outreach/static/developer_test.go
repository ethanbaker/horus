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
	config, errs := config.NewConfigFromFile("../../testing/.env")
	assert.Equal(1, len(errs))
	assert.Equal("cannot initialize mysql config; are all 'SQL' fields set?", errs[0].Error())

	// Run the test
	output := static.Ping(config)
	assert.Equal("Ping!", output)
}
