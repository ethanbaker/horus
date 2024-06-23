package types

import (
	"time"

	"github.com/ethanbaker/horus/utils/config"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

/* ---- OUTREACH MESSAGE ---- */

// OutreachMessage type is used to encode a message that will be sent to the user through a preferred communication method
type OutreachMessage interface {
	// Returns a function that returns the content of the message
	GetContent(*config.Config) string

	// Return a list of channels to send the message
	GetChannels() []chan string

	// Start starts the message repeating again
	Start() error

	// Stop stops the message from being sent and cancels it permenantly
	Stop() error

	// Delete stops and deletes the message forever
	Delete() error
}

/* ---- OUTREACH METHODS ---- */

// OutreachMethods represent different methods a user can be connected to
type OutreachMethod string

const (
	DiscordMethod  OutreachMethod = "discord"
	TelegramMethod OutreachMethod = "telegram"
	SmsMethod      OutreachMethod = "sms"
)

/* ---- OUTREACH MODULES ---- */

// OutreachModule represent different types of outreach messages that can be used
type OutreachModule string

const (
	StaticModule  OutreachModule = "static"
	DynamicModule OutreachModule = "dynamic"
	TimerModule   OutreachModule = "timer"
)

/* ---- OUTREACH CONFIG ---- */
// OutreachConfig types are used to read in the 'outreach.yml' config file

type OutreachConfigMessage struct {
	Name     string           `yaml:"name"`
	Channels []OutreachMethod `yaml:"channels"`
	Data     map[string]any   `yaml:"data"`
}

type OutreachConfig struct {
	Static []OutreachConfigMessage `yaml:"static"`

	Dynamic []OutreachConfigMessage `yaml:"dynamic"`
}

/* ---- OUTREACH MANAGER ---- */

// OutreachManager
type OutreachManager struct {
	Config   *config.Config
	Services *OutreachServices
	Modules  map[OutreachModule]func(*OutreachManager, []chan string, map[string]any) (OutreachMessage, error)
	Channels map[OutreachMethod]chan string
	Messages map[string]OutreachMessage
}

// OutreachServices hold vital services for outreach messages to perform their operations
type OutreachServices struct {
	Config *config.Config
	DB     *gorm.DB
	Cron   *cron.Cron
	Clock  *time.Ticker
}
