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
	config, errs := config.NewConfigFromFile("../../testing/.env")
	assert.Equal(1, len(errs))
	assert.Equal("cannot initialize mysql config; are all 'SQL' fields set?", errs[0].Error())

	err := static.Init(config)
	assert.Nil(err)

	// Run the function
	output := static.NotionDailyDigest(config)
	assert.NotNil(output)

	log.Println("Received output: ")
	log.Println(output)
}

func TestNotionNightAffirmations(t *testing.T) {
	assert := assert.New(t)

	// Initialize the environment
	config, errs := config.NewConfigFromFile("../../testing/.env")
	assert.Equal(1, len(errs))
	assert.Equal("cannot initialize mysql config; are all 'SQL' fields set?", errs[0].Error())

	err := static.Init(config)
	assert.Nil(err)

	// Run the function
	output := static.NotionNightAffirmations(config)
	assert.NotNil(output)

	log.Println("Received output: ")
	log.Println(output)
}

func TestNotionMorningAffirmations(t *testing.T) {
	assert := assert.New(t)

	// Initialize the environment
	config, errs := config.NewConfigFromFile("../../testing/.env")
	assert.Equal(1, len(errs))
	assert.Equal("cannot initialize mysql config; are all 'SQL' fields set?", errs[0].Error())

	err := static.Init(config)
	assert.Nil(err)

	// Run the function
	output := static.NotionMorningAffirmations(config)
	assert.NotNil(output)

	log.Println("Received output: ")
	log.Println(output)
}
