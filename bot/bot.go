package horus

import (
	"encoding/json"
	"fmt"

	"github.com/ethanbaker/horus/utils/types"
	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/objx"
	"gorm.io/gorm"
)

// Bot represents an implementation of Horus. It contains multiple conversations with a given user as well as defining characteristics
type Bot struct {
	gorm.Model

	Name          string         // The name of the bot
	Permissions   byte           // A byte representation of allowed permissions
	Memory        Memory         // Static memory associated with a bot
	Conversations []Conversation // A list of conversations with the user

	// Initalized variables (don't change after creation)
	client              *openai.Client                                  `gorm:"-"` // The OpenAI client
	functionDefinitions map[string]openai.FunctionDefinition            `gorm:"-"` // Function definitions to plug into GPT prompts
	handlers            []func(function string, input *types.Input) any `gorm:"-"` // A list of handlers from associated modules

	// Dynamic variables (can change after creation)
	functionQueue []func(bot *Bot, input *types.Input) *types.Output `gorm:"-"` // Incoming functions to run instead of delegating to OpenAI
	variables     map[string]any                                     `gorm:"-"` // Any variables used by functions
}

// AddConversation adds a new conversation to the bot
func (b *Bot) AddConversation(key string) error {
	if key == "" {
		return fmt.Errorf("conversation key cannot be empty")
	}

	// Make sure the key is not a duplicate
	for _, c := range b.Conversations {
		if c.Name == key {
			return fmt.Errorf("cannot add conversation with duplicate key '%s'", key)
		}
	}

	// Create a new conversation to add
	c, err := newConversation(b.Model.ID, key)
	if err != nil {
		return err
	}
	c.setup(b.client, &b.functionDefinitions)

	// Add the conversation to the bot
	b.Conversations = append(b.Conversations, c)

	// Save the bot
	return db.Save(&b).Error
}

// DeleteConversation delets a conversation from the bot
func (b *Bot) DeleteConversation(key string) error {
	// Find the associated conversation and remove it from the library
	for i, c := range b.Conversations {
		if c.Name == key {
			// Remove the conversation and save the bot
			old := b.Conversations[i]
			b.Conversations = append(b.Conversations[:i], b.Conversations[i+1:]...)

			return old.Delete()
		}
	}

	return fmt.Errorf("conversation with given key does not exist")
}

// IsConversation returns true if the conversation exists
func (b *Bot) IsConversation(key string) bool {
	for _, c := range b.Conversations {
		if c.Name == key {
			return true
		}
	}

	return false
}

// SendMessage sends a message to the bot in a given conversation
func (b *Bot) SendMessage(key string, input *types.Input) (*types.Output, error) {
	output := types.Output{}

	input.Permissions = input.Permissions & b.Permissions

	// Continue only if GPT functionality is enabled
	if b.Permissions|PERMISSIONS_GPT == 0 {
		return nil, fmt.Errorf("gpt functionality is not enabled")
	}

	// Find the conversation
	var conversation *Conversation
	for i, c := range b.Conversations {
		if c.Name == key {
			conversation = &b.Conversations[i]
			break
		}
	}

	// If the conversation does not exist, return error
	if conversation.Name == "" {
		return nil, fmt.Errorf("conversation with key '%s' does not exist", key)
	}

	// If there is a queued function, run it
	if qf := b.nextQueuedFunction(); qf != nil {
		// A function is queued; get the response directly from the function
		output = *qf(b, input)
		return &output, output.Error
	}

	// Get the GPT response
	resp, err := conversation.SendMessage(openai.ChatMessageRoleUser, "user", input.Message)
	if err != nil {
		return nil, err
	}

	// Check for a function call
	if len(resp.Choices[0].Message.ToolCalls) != 0 {
		// Remove duplicate function calls
		allKeys := map[string]bool{}
		calls := []openai.ToolCall{}
		for _, call := range resp.Choices[0].Message.ToolCalls {
			if _, ok := allKeys[call.Function.Name]; !ok {
				allKeys[call.Function.Name] = true
				calls = append(calls, call)
			}
		}
		resp.Choices[0].Message.ToolCalls = calls

		// Go through each unique call
		for _, call := range calls {
			// Add the call to the conversation
			if err = conversation.AddFunctionCall(&resp.Choices[0].Message); err != nil {
				return nil, err
			}

			// A function call is present, parse the arguments
			input.Parameters, err = objx.FromJSON(call.Function.Arguments)
			if err != nil {
				return nil, err
			}

			// Call associated module handlers
			for _, f := range b.handlers {
				// Only continue for functions that return an output
				if output := f(call.Function.Name, input); output != nil {
					// Check if output matches the output type. If so, return
					val, ok := output.(*types.Output)
					if ok {
						return val, val.Error
					}

					// Marshal the output into a string
					message, err := json.Marshal(output)
					if err != nil {
						return nil, err
					}

					// Send the output of the function call to the OpenAI model and get a new response
					callMessage := openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						Name:       call.Function.Name,
						Content:    string(message),
						ToolCallID: call.ID,
					}
					if err = conversation.AddFunctionCall(&callMessage); err != nil {
						return nil, err
					}

					break
				}
			}
		}

		// Send the function calls for a new response
		resp, err = conversation.SendFunctionCalls()
		if err != nil {
			return nil, err
		}
	}

	output.Message = resp.Choices[0].Message.Content

	// Save any memory changes that may have taken place
	return &output, db.Save(&b.Memory).Error
}

// Add a message to a conversation
func (b *Bot) AddMessage(key string, role string, name string, content string) error {
	// Find the conversation
	var conversation *Conversation
	for i, c := range b.Conversations {
		if c.Name == key {
			conversation = &b.Conversations[i]
			break
		}
	}

	// If the conversation does not exist, return error
	if conversation.Name == "" {
		return fmt.Errorf("conversation with key '%s' does not exist", key)
	}

	// Add the message to the conversation
	if err := conversation.AddMessage(role, name, content); err != nil {
		return err
	}

	return db.Save(&b.Memory).Error
}

// Get the next queued function
func (b *Bot) nextQueuedFunction() func(bot *Bot, input *types.Input) *types.Output {
	if len(b.functionQueue) == 0 {
		return nil
	}

	f := b.functionQueue[0]
	b.functionQueue = b.functionQueue[1:]

	return f
}

// Add a list of functions to the function queue
func (b *Bot) AddQueuedFunctions(funcs ...func(bot *Bot, input *types.Input) *types.Output) {
	b.functionQueue = append(b.functionQueue, funcs...)
}

// Adds handlers to the bot's handlers
func (b *Bot) AddHandlers(handlers ...func(function string, input *types.Input) any) {
	b.handlers = append(b.handlers, handlers...)
}

// Adds definitions to the bot's function definitions
func (b *Bot) AddDefinitions(name string, definitions *map[string]openai.FunctionDefinition) {
	for key, f := range *definitions {
		b.functionDefinitions[fmt.Sprintf("%v-%v", name, key)] = f
	}
}

// Writes a value to the variables map
func (b *Bot) EditVariable(key string, value any) {
	b.variables[key] = value
}

// Returns a value from the variables map
func (b *Bot) GetVariable(key string) any {
	return b.variables[key]
}

// Setup sets up the bot with an openai Client
func (b *Bot) Setup(client *openai.Client) {
	b.client = client

	// Set up each associated conversations
	for i := range b.Conversations {
		b.Conversations[i].setup(client, &b.functionDefinitions)
	}
}

// NewBot creates a new Bot object
func NewBot(name string, permissions byte) (*Bot, error) {
	// Create the library
	b := Bot{
		Name:                name,
		Permissions:         permissions,
		Memory:              Memory{},
		Conversations:       []Conversation{},
		functionDefinitions: map[string]openai.FunctionDefinition{},
		handlers:            []func(function string, input *types.Input) any{},
		functionQueue:       []func(bot *Bot, input *types.Input) *types.Output{},
		variables:           map[string]any{},
	}

	return &b, db.Create(&b).Error
}
