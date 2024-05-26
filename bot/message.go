package horus

import (
	"github.com/ethanbaker/horus/utils/config"
	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// TooCall is a helper struct to encode tool call information into messages
type ToolCall struct {
	gorm.Model

	ID            string
	Type          string
	CallName      string
	CallArguments string
	MessageID     uint
}

// Message represents a sent message in a conversation
type Message struct {
	gorm.Model

	ConversationID uint   // The conversation this message is related to
	Idx            uint   // The index of the message in the conversation
	Role           string // The role of the entity speaking the message
	Name           string // The message's type
	Content        string // The content of the message

	// Tools related to the message call
	ToolCallID string
	ToolCalls  []ToolCall

	// The openAI message this message is representing
	ChatCompletionMessage *openai.ChatCompletionMessage `gorm:"-"`

	// The config used for saving this message to the database
	config *config.Config `gorm:"-"`
}

// Delete deletes the message from the SQL database
func (m *Message) Delete() error {
	// Delete tool calls
	for _, c := range m.ToolCalls {
		if err := m.config.Gorm.Delete(c).Error; err != nil {
			return err
		}
	}

	m.ToolCalls = []ToolCall{}
	return m.config.Gorm.Delete(m).Error
}

// newMessage creates a new message
func newMessage(config *config.Config, conversationID uint, index uint, message *openai.ChatCompletionMessage) (Message, error) {
	// Create the new message
	m := Message{
		ConversationID:        conversationID,
		Idx:                   index,
		Role:                  message.Role,
		Name:                  message.Name,
		Content:               message.Content,
		ToolCallID:            message.ToolCallID,
		ToolCalls:             []ToolCall{},
		ChatCompletionMessage: message,
		config:                config,
	}

	// Save the message
	if err := m.config.Gorm.Create(&m).Error; err != nil {
		return m, err
	}

	// Add function calls to the message
	for _, call := range message.ToolCalls {
		c := ToolCall{
			ID:            call.ID,
			Type:          string(call.Type),
			CallName:      call.Function.Name,
			CallArguments: call.Function.Arguments,
			MessageID:     m.Model.ID,
		}

		if err := m.config.Gorm.Create(&c).Error; err != nil {
			return m, err
		}

		m.ToolCalls = append(m.ToolCalls, c)
	}

	// Save the message to the SQL database and return
	return m, nil
}
