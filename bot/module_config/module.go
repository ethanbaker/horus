package module_config

import (
	horus "github.com/ethanbaker/horus/bot"
	"github.com/ethanbaker/horus/utils/types"
	openai "github.com/sashabaranov/go-openai"
)

// Module stores this module's functions and capabilities in an easily exportable struct
type Module struct {
	Enabled     bool // Whether or not the module is enabled
	Permissions byte // The permissions this module needs to be activated

	FunctionDefinitions map[string]openai.FunctionDefinition                    // Function definitions for an OpenAI model
	Functions           map[string]func(bot *horus.Bot, input *types.Input) any // Functions that can be called

	bot *horus.Bot // The Horus bot this module is attached to
}

// Return the name of the module
func (m *Module) Name() string {
	return "config"
}

// Handle a function call
func (m *Module) Handler(function string, input *types.Input) any {
	// Check for permissions
	if input.Permissions|m.Permissions == 0 {
		return nil
	}

	// Check all functions
	for label, f := range m.Functions {
		if label == function {
			return f(m.bot, input)
		}
	}

	return nil
}

// Create a new Module
func NewModule(bot *horus.Bot, enabled bool) {
	// Create the module and add static information
	var m Module
	m.Enabled = enabled
	m.Permissions = horus.PERMISSIONS_PRVMODULES
	m.Functions = functions
	m.bot = bot

	// Add the module's handler and function definitions to the bot
	bot.AddHandlers(m.Handler)
	bot.AddDefinitions(m.Name(), &functionDefinitions)
}
