package module_template

import (
	"github.com/ethanbaker/horus/utils/schema"
	openai "github.com/sashabaranov/go-openai"
)

// A map of each OpenAI function declaration present in this module
var functionDefinitions = map[string]openai.FunctionDefinition{
	// Template
	"function_1": { // Name of the function in the map
		Name:        "function_1",          // The name of the function for the model to know
		Description: "Desc for function_1", // The description of the function for the model to understand
		Parameters:  schema.Definition{},   // The parameters of the function for the model to understand usage (it must be json serialized if a struct)
	},

	// OpenAI Documentation Demo
	"get_current_weather_demo": {
		Name:        "get_current_weather_demo",
		Description: "Get the current weather in a given location",
		Parameters: schema.Definition{
			Type: schema.Object,
			Properties: map[string]schema.Definition{
				"location": {
					Type:        schema.String,
					Description: "The city and state, e.g. San Francisco, CA",
				},
				"unit": {
					Type: schema.String,
					Enum: []string{"celsius", "fahrenheit"},
				},
			},
		},
	},
}
