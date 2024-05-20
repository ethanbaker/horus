package horus

import (
	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

var db *gorm.DB

// initSQL initializes the SQL database the structs are connected to
func initSQL(c *config.Config) error {
	var err error

	db = c.Gorm

	// Migrate all of the schemas
	if err = db.AutoMigrate(&ToolCall{}); err != nil {
		return err
	}
	if err = db.AutoMigrate(&Message{}); err != nil {
		return err
	}
	if err = db.AutoMigrate(&Memory{}); err != nil {
		return err
	}
	if err = db.AutoMigrate(&Conversation{}); err != nil {
		return err
	}
	if err = db.AutoMigrate(&Bot{}); err != nil {
		return err
	}

	return nil
}

// GetAllBots gets a list of all bots in the SQL database
func GetAllBots() ([]Bot, error) {
	bots := []Bot{}

	// Load associated memory objects
	if err := db.Model(&Bot{}).Preload("Memory").Preload("Conversations.Messages.ToolCalls").Find(&bots).Error; err != nil {
		return bots, err
	}

	// initialize bot fields
	for i := range bots {
		bot := &bots[i]
		bot.functionDefinitions = map[string]openai.FunctionDefinition{}
		bot.handlers = []func(function string, input *types.Input) any{}
		bot.functionQueue = []func(bot *Bot, input *types.Input) *types.Output{}
		bot.variables = map[string]any{}
	}

	return bots, nil
}

// GetBotByName gets a singular bot by name from the SQL database
func GetBotByName(name string) (*Bot, error) {
	// Get all of the bots
	bots, err := GetAllBots()
	if err != nil {
		return nil, err
	}

	// Try to find the bot
	for i := range bots {
		if bots[i].Name == name {
			return &bots[i], nil
		}
	}

	// Return a nil pointer if the bot is not found
	return nil, nil
}
