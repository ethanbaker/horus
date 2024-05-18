package module_config

import (
	"github.com/ethanbaker/horus/utils/schema"
	openai "github.com/sashabaranov/go-openai"
)

// A map of each OpenAI function declaration present in this module
var functionDefinitions = map[string]openai.FunctionDefinition{
	// Configure the timezone
	"set_timezone": {
		Name:        "set_timezone",
		Description: "Set the user's preferred timezone to a value supplied by the user",
		Parameters: schema.Definition{
			Type: schema.Object,
			Properties: map[string]schema.Definition{
				"timezone": {
					Type:        schema.String,
					Description: "The timezone corresponding to the IANA Time Zone Database, ex: America/New_York",
				},
			},
			Required: []string{"timezone"},
		},
	},

	// Configure the city
	"set_city": {
		Name:        "set_city",
		Description: "Set the user's preferred city location to a value supplied by the user",
		Parameters: schema.Definition{
			Type: schema.Object,
			Properties: map[string]schema.Definition{
				"city": {
					Type:        schema.String,
					Description: "The city, ex: Raleigh",
				},
			},
			Required: []string{"city"},
		},
	},

	// Configure the preferred temperature unit
	"set_temperature_unit": {
		Name:        "set_temperature_unit",
		Description: "Record the user's preferred temperature unit",
		Parameters: schema.Definition{
			Type: schema.Object,
			Properties: map[string]schema.Definition{
				"unit": {
					Type: schema.String,
					Enum: []string{"celsius", "fahrenheit"},
				},
			},
			Required: []string{"unit"},
		},
	},
}
