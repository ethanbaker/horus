package static_test

import (
	"log"
	"testing"

	"github.com/ethanbaker/horus/outreach/static"
	"github.com/ethanbaker/horus/utils/config"
	"github.com/stretchr/testify/assert"
)

/* ---- MESSAGE TESTS ---- */

func TestNotionDailyDigest(t *testing.T) {
	assert := assert.New(t)

	// Initialize the environment
	config, errs := config.New()
	assert.Equal(0, len(errs))

	err := static.Init(config)
	assert.Nil(err)

	// Run the function
	output := static.NotionDailyDigest(config, map[string]any{})
	assert.NotNil(output)

	// NOTE: the received output won't sort the all day events correctly because the test database lies and says past events are today
	log.Printf("[RESULT]: Received output `%v`\n", output)
}

func TestNotionReadPage(t *testing.T) {
	assert := assert.New(t)

	// Initialize the environment
	config, errs := config.New()
	assert.Equal(0, len(errs))

	err := static.Init(config)
	assert.Nil(err)

	// Run the function
	output := static.NotionReadPage(config, map[string]any{
		"page-id": config.Getenv("NOTION_TEST_PAGE"),
	})
	assert.NotNil(output)

	log.Printf("[RESULT]: Received output \n`%v`\n", output)
}
