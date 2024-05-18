// definitions.go contains all OpenAI function declarations
package module_ambient

import (
	"github.com/ethanbaker/horus/utils/schema"
	openai "github.com/sashabaranov/go-openai"
)

// A map of each OpenAI function declaration present in this module
var functionDefinitions = map[string]openai.FunctionDefinition{
	// Get the current time
	"get_current_time": {
		Name:        "get_current_time",
		Description: "Get the current time",
		Parameters: schema.Definition{
			Type:       schema.Object,
			Properties: map[string]schema.Definition{},
		},
	},

	// Get the current weather
	"get_current_weather": {
		Name:        "get_current_weather",
		Description: "Get the current weather in a given location",
		Parameters: schema.Definition{
			Type: schema.Object,
			Properties: map[string]schema.Definition{
				"location": {
					Type:        schema.String,
					Description: "The city (ex: 'Raleigh'). Do not include any state codes.",
				},
				"unit": {
					Type: schema.String,
					Enum: []string{"celsius", "fahrenheit"},
				},
			},
		},
	},
}
