package config_test

import (
	"fmt"
	"testing"

	"github.com/ethanbaker/horus/utils/config"
	"github.com/stretchr/testify/assert"
)

// Expected errors our config should encounter
var expectedErrors = []error{
	fmt.Errorf("cannot initialize openai client; is 'OPENAI_TOKEN' set?"),
	fmt.Errorf("cannot initialize mysql config; are all 'SQL' fields set?"),
	fmt.Errorf("cannot initialize notion client. Is 'NOTION_API_TOKEN' set?"),
}

func TestNewConfigFromFile(t *testing.T) {
	assert := assert.New(t)

	// Create a new config
	config, errs := config.NewConfigFromFile("./testing/.env.test")
	assert.NotNil(config)

	// Make sure the variables got loaded correctly from the .env file
	assert.Equal("abc123", config.Getenv("VARIABLE_1"))
	assert.Equal("def456", config.Getenv("VARIABLE_2"))
	assert.Equal("ghi789", config.Getenv("VARIABLE_3"))

	// Make sure the discord channel IDs loaded correctly
	assert.Equal(3, len(config.DiscordOpenChannels))
	assert.Equal("123", config.DiscordOpenChannels[0])
	assert.Equal("456", config.DiscordOpenChannels[1])
	assert.Equal("789", config.DiscordOpenChannels[2])

	assert.Equal(3, len(config.DiscordThreadChannels))
	assert.Equal("111", config.DiscordThreadChannels[0])
	assert.Equal("222", config.DiscordThreadChannels[1])
	assert.Equal("333", config.DiscordThreadChannels[2])

	// Expect different errors from services not being set up
	//  - OpenAI Client
	//  - MySQL Client
	//  - Notion Client
	assert.Equal(len(expectedErrors), len(errs))

	for _, err := range errs {
		// Match up against expected errors
		matched := false
		for i := 0; i < len(expectedErrors) && !matched; i++ {
			actual := err.Error()
			expected := expectedErrors[i].Error()

			matched = matched || actual == expected
		}

		// If no matching error, fail
		if !matched {
			assert.Failf("unexpected error (%v)\n", err.Error())
		}
	}
}

func TestNewConfigFromVariables(t *testing.T) {
	assert := assert.New(t)

	vars := map[string]string{
		"VARIABLE_1":                  "abc123",
		"VARIABLE_2":                  "def456",
		"VARIABLE_3":                  "ghi789",
		"DISCORD_BOT_OPEN_CHANNELS":   "123,456,789",
		"DISCORD_BOT_THREAD_CHANNELS": "111,222,333",
	}

	// Create a new config
	config, errs := config.NewConfigFromVariables(&vars)
	assert.NotNil(config)

	// Make sure the variables got loaded correctly from the .env file
	assert.Equal("abc123", config.Getenv("VARIABLE_1"))
	assert.Equal("def456", config.Getenv("VARIABLE_2"))
	assert.Equal("ghi789", config.Getenv("VARIABLE_3"))

	// Make sure the discord channel IDs loaded correctly
	assert.Equal(3, len(config.DiscordOpenChannels))
	assert.Equal("123", config.DiscordOpenChannels[0])
	assert.Equal("456", config.DiscordOpenChannels[1])
	assert.Equal("789", config.DiscordOpenChannels[2])

	assert.Equal(3, len(config.DiscordThreadChannels))
	assert.Equal("111", config.DiscordThreadChannels[0])
	assert.Equal("222", config.DiscordThreadChannels[1])
	assert.Equal("333", config.DiscordThreadChannels[2])

	// Expect different errors from services not being set up
	//  - OpenAI Client
	//  - MySQL Client
	//  - Notion Client
	assert.Equal(len(expectedErrors), len(errs))

	for _, err := range errs {
		// Match up against expected errors
		matched := false
		for i := 0; i < len(expectedErrors) && !matched; i++ {
			actual := err.Error()
			expected := expectedErrors[i].Error()

			matched = matched || actual == expected
		}

		// If no matching error, fail
		if !matched {
			assert.Failf("unexpected error (%v)\n", err.Error())
		}
	}
}
