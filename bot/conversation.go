package horus

import (
	"context"

	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// Conversation represents a conversation between Horus and the user
type Conversation struct {
	gorm.Model

	BotID    uint      // The foreign key to relate the conversation to a bot
	Name     string    // A unique identifying key for the converesation
	Messages []Message // A list of messages in the conversation

	client  *openai.Client               `gorm:"-"` // The OpenAI client the conversation is attached to
	request openai.ChatCompletionRequest `gorm:"-"` // The OpenAI request this conversation is emulating
	config  *config.Config               `gorm:"-"`
}

// Delete a conversation and all associated messages
func (c *Conversation) Delete() error {
	// Delete all messages
	var err error
	for i := range c.Messages {
		if err = c.Messages[i].Delete(); err != nil {
			return err
		}
	}

	// Delete the conversation
	c.Messages = []Message{}
	return c.config.Gorm.Delete(c).Error
}

// Append a message to the conversation's request
func (c *Conversation) appendMessage(m Message) error {
	// Add the message to the conversation
	c.Messages = append(c.Messages, m)

	c.request.Messages = append(c.request.Messages, *m.ChatCompletionMessage)

	// Save the conversation
	return c.config.Gorm.Save(&c).Error
}

// Add function call to the conversation
func (c *Conversation) AddFunctionCall(message *openai.ChatCompletionMessage) error {
	m, err := newMessage(c.config, c.Model.ID, uint(len(c.Messages)), message)
	if err != nil {
		return err
	}

	return c.appendMessage(m)
}

// SendFunctionCalls gets a new response with added function calls
func (c *Conversation) SendFunctionCalls() (*openai.ChatCompletionResponse, error) {
	// Get the chat completion
	resp, err := c.client.CreateChatCompletion(context.Background(), c.request)
	if err != nil {
		return nil, err
	}

	// Create a new message from the bot and add it to the conversation
	m, err := newMessage(c.config, c.Model.ID, uint(len(c.Messages)), &resp.Choices[0].Message)
	if err != nil {
		return nil, err
	}

	return &resp, c.appendMessage(m)
}

// SendMessage sends a message to OpenAI
func (c *Conversation) SendMessage(role string, name string, input *types.Input) (*openai.ChatCompletionResponse, error) {
	// Add the message to the chat completion request
	chatCompletionMessage := openai.ChatCompletionMessage{
		Role:    role,
		Name:    name,
		Content: input.Message,
	}

	// Create a new message from the user
	m, err := newMessage(c.config, c.Model.ID, uint(len(c.Messages)), &chatCompletionMessage)
	if err != nil {
		return nil, err
	}

	// Add the message to the conversation and save it
	if err := c.appendMessage(m); err != nil {
		return nil, err
	}

	// Get the chat completion
	c.request.Temperature = input.Temperature
	resp, err := c.client.CreateChatCompletion(context.Background(), c.request)
	if err != nil {
		return nil, err
	}

	// Return on function calls
	if resp.Choices[0].Message.Content == "" {
		return &resp, nil
	}

	// Create a new message from the bot and add it to the conversation
	m, err = newMessage(c.config, c.Model.ID, uint(len(c.Messages)), &resp.Choices[0].Message)
	if err != nil {
		return nil, err
	}

	return &resp, c.appendMessage(m)
}

// Add a message to the conversation without sending it
func (c *Conversation) AddMessage(role string, name string, content string) error {
	// Add the message to the chat completion request
	chatCompletionMessage := openai.ChatCompletionMessage{
		Role:    role,
		Name:    name,
		Content: content,
	}

	// Create a new message from the user
	m, err := newMessage(c.config, c.Model.ID, uint(len(c.Messages)), &chatCompletionMessage)
	if err != nil {
		return err
	}

	// Add the message to the conversation and save it
	if err := c.appendMessage(m); err != nil {
		return err
	}

	return nil
}

// Add a tool response to the conversation without sending it
func (c *Conversation) AddToolResponse(role string, name string, content string, id string) error {
	// Add the message to the chat completion request
	chatCompletionMessage := openai.ChatCompletionMessage{
		Role:       role,
		Name:       name,
		Content:    content,
		ToolCallID: id,
	}

	// Create a new message from the user
	m, err := newMessage(c.config, c.Model.ID, uint(len(c.Messages)), &chatCompletionMessage)
	if err != nil {
		return err
	}

	// Add the message to the conversation and save it
	if err := c.appendMessage(m); err != nil {
		return err
	}

	return nil
}

// Remove function calls not matching a given ID. We won't be running the associated function so we need to pretend they didn't exist
func (c *Conversation) TruncateCalls(id string) {
	// The bot makes the function call in the last message
	lastMessage := &c.Messages[len(c.Messages)-1]

	for _, call := range lastMessage.ToolCalls {
		// If ids match, empty the rest of the tool calls
		if call.ID == id {
			lastMessage.ToolCalls = []ToolCall{call}
			break
		}
	}

	// Do the same for the request
	lastRequest := &c.request.Messages[len(c.request.Messages)-1]

	for _, call := range lastRequest.ToolCalls {
		// If ids match, empty the rest of the tool calls
		if call.ID == id {
			lastRequest.ToolCalls = []openai.ToolCall{call}
			break
		}
	}
}

// Sets up a conversation with the OpenAI client
func (c *Conversation) setup(config *config.Config) {
	c.config = config

	// Setup the client and request
	c.client = config.Openai
	c.request = openai.ChatCompletionRequest{
		Model:    OPENAI_MODEL,
		Messages: []openai.ChatCompletionMessage{},
		Stream:   false,
	}

	// Add existing messages to the request
	for i := 0; i < len(c.Messages); i++ {
		m := &c.Messages[i]
		m.setup(config)

		// Create the chat completion message
		ccm := openai.ChatCompletionMessage{
			Role:       m.Role,
			Name:       m.Name,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}

		// Add tool calls
		for _, call := range m.ToolCalls {
			ccm.ToolCalls = append(ccm.ToolCalls, openai.ToolCall{
				ID:   call.ID,
				Type: openai.ToolType(call.Type),
				Function: openai.FunctionCall{
					Name:      call.CallName,
					Arguments: call.CallArguments,
				},
			})
		}

		c.request.Messages = append(c.request.Messages, ccm)
	}
}

// addDefinitions adds tool definitions to the conversation
func (c *Conversation) addDefinitions(functions *map[string]openai.FunctionDefinition) {
	// Setup the function calls/tools
	var tools []openai.Tool
	for k := range *functions {
		def := (*functions)[k]
		tools = append(tools, openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: &def,
		})
	}

	// Add the functions to the request
	c.request.Tools = tools
}

// newConversation creates a new conversation
func newConversation(config *config.Config, botID uint, key string) (Conversation, error) {
	// Create the new conversation
	c := Conversation{
		BotID:  botID,
		Name:   key,
		config: config,
	}

	// Save the conversation
	if res := config.Gorm.Create(&c); res.Error != nil {
		return c, res.Error
	}

	// Add an initial setup message to the conversation
	message := openai.ChatCompletionMessage{
		Role:    OPENAI_ROLE,
		Name:    "system",
		Content: OPENAI_SYSPROMPT,
	}

	m, err := newMessage(c.config, c.Model.ID, 0, &message)
	if err != nil {
		return c, err
	}
	if err := c.appendMessage(m); err != nil {
		return c, err
	}

	return c, nil
}
