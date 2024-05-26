package horus

import (
	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	openai "github.com/sashabaranov/go-openai"
)

// InitSQL initializes the SQL database the structs are connected to
func InitSQL(c *config.Config) error {
	var err error

	// Migrate all of the schemas
	if err = c.Gorm.AutoMigrate(&ToolCall{}); err != nil {
		return err
	}
	if err = c.Gorm.AutoMigrate(&Message{}); err != nil {
		return err
	}
	if err = c.Gorm.AutoMigrate(&Memory{}); err != nil {
		return err
	}
	if err = c.Gorm.AutoMigrate(&Conversation{}); err != nil {
		return err
	}
	if err = c.Gorm.AutoMigrate(&Bot{}); err != nil {
		return err
	}

	return nil
}

// GetAllBots gets a list of all bots in the SQL database
func GetAllBots(c *config.Config) ([]Bot, error) {
	bots := []Bot{}

	// Load associated memory objects
	if err := c.Gorm.Model(&Bot{}).Preload("Memory").Preload("Conversations.Messages.ToolCalls").Find(&bots).Error; err != nil {
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
func GetBotByName(name string, cfg *config.Config) (*Bot, error) {
	// Get all of the bots
	bots, err := GetAllBots(cfg)
	if err != nil {
		return nil, err
	}

	// Try to find the bot
	for i := range bots {
		if bots[i].Name == name {
			return &bots[i], bots[i].setup(cfg, false)
		}
	}

	// Return a nil pointer if the bot is not found
	return nil, nil
}
