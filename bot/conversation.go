package horus

import (
	"context"

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
	return db.Delete(c).Error
}

// Append a message to the conversation's request
func (c *Conversation) appendMessage(m Message) error {
	// Add the message to the conversation
	c.Messages = append(c.Messages, m)

	c.request.Messages = append(c.request.Messages, *m.ChatCompletionMessage)

	// Save the conversation
	return db.Save(&c).Error
}

// Add function call to the conversation
func (c *Conversation) AddFunctionCall(message *openai.ChatCompletionMessage) error {
	m, err := newMessage(c.Model.ID, uint(len(c.Messages)), message)
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
	m, err := newMessage(c.Model.ID, uint(len(c.Messages)), &resp.Choices[0].Message)
	if err != nil {
		return nil, err
	}

	return &resp, c.appendMessage(m)
}

// SendMessage sends a message to OpenAI
func (c *Conversation) SendMessage(role string, name string, content string) (*openai.ChatCompletionResponse, error) {
	// Add the message to the chat completion request
	chatCompletionMessage := openai.ChatCompletionMessage{
		Role:    role,
		Name:    name,
		Content: content,
	}

	// Create a new message from the user
	m, err := newMessage(c.Model.ID, uint(len(c.Messages)), &chatCompletionMessage)
	if err != nil {
		return nil, err
	}

	// Add the message to the conversation and save it
	if err := c.appendMessage(m); err != nil {
		return nil, err
	}

	// Get the chat completion
	resp, err := c.client.CreateChatCompletion(context.Background(), c.request)
	if err != nil {
		return nil, err
	}

	// Return on function calls
	if resp.Choices[0].Message.Content == "" {
		return &resp, nil
	}

	// Create a new message from the bot and add it to the conversation
	m, err = newMessage(c.Model.ID, uint(len(c.Messages)), &resp.Choices[0].Message)
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
	m, err := newMessage(c.Model.ID, uint(len(c.Messages)), &chatCompletionMessage)
	if err != nil {
		return err
	}

	// Add the message to the conversation and save it
	if err := c.appendMessage(m); err != nil {
		return err
	}

	return nil
}

// Sets up a conversation with the OpenAI client
func (c *Conversation) setup(client *openai.Client, functions *map[string]openai.FunctionDefinition) {
	// Setup the client and request
	c.client = client
	c.request = openai.ChatCompletionRequest{
		Model:    OPENAI_MODEL,
		Messages: []openai.ChatCompletionMessage{},
		Stream:   false,
	}

	// Add existing messages to the request
	for _, m := range c.Messages {
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
func newConversation(botID uint, key string) (Conversation, error) {
	// Create the new conversation
	c := Conversation{
		BotID: botID,
		Name:  key,
	}

	// Save the conversation
	if res := db.Create(&c); res.Error != nil {
		return c, res.Error
	}

	// Add an initial setup message to the conversation
	message := openai.ChatCompletionMessage{
		Role:    OPENAI_ROLE,
		Name:    "system",
		Content: OPENAI_SYSPROMPT,
	}

	m, err := newMessage(c.Model.ID, 0, &message)
	if err != nil {
		return c, err
	}
	if err := c.appendMessage(m); err != nil {
		return c, err
	}

	return c, nil
}
