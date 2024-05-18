package module_keepass

import (
	"github.com/ethanbaker/horus/utils/schema"
	openai "github.com/sashabaranov/go-openai"
)

// A map of each OpenAI function declaration present in this module
var functionDefinitions = map[string]openai.FunctionDefinition{
	// Get the keepass information
	"keepass_get": {
		Name:        "keepass_get",
		Description: "Get the keepass password database",
		Parameters: schema.Definition{
			Type:       schema.Object,
			Properties: map[string]schema.Definition{},
		},
	},

	// Create a keepass profile
	"keepass_create": {
		Name:        "keepass_create",
		Description: "Create a password profile in a keepass database",
		Parameters: schema.Definition{
			Type:       schema.Object,
			Properties: map[string]schema.Definition{},
		},
	},

	// Update a keepass profile
	"keepass_update": {
		Name:        "keepass_update",
		Description: "Update a password profile in a keepass database",
		Parameters: schema.Definition{
			Type:       schema.Object,
			Properties: map[string]schema.Definition{},
		},
	},

	// Delete a keepass profile
	"keepass_delete": {
		Name:        "keepass_delete",
		Description: "Delete a password profile in a keepass database",
		Parameters: schema.Definition{
			Type:       schema.Object,
			Properties: map[string]schema.Definition{},
		},
	},
}
