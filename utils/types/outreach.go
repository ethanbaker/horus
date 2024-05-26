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

type OutreachMethod string

const (
	Discord  OutreachMethod = "discord"
	Telegram OutreachMethod = "telegram"
)

/* ---- OUTREACH INPUT ---- */

type StaticOutreach struct {
	Function string
	Repeat   string
}

type DynamicOutreach struct {
	Function        string
	IntervalMinutes time.Duration
}

/* ---- OUTREACH CONFIG ---- */

type OutreachConfig struct {
	Static []struct {
		Name   string `yaml:"name"`
		Key    string `yaml:"key"`
		Repeat string `yaml:"repeat"`
	} `yaml:"static"`

	Dynamic []struct {
		Name            string `yaml:"name"`
		Key             string `yaml:"key"`
		IntervalMinutes int    `yaml:"interval"`
	} `yaml:"dynamic"`
}

/* ---- OUTREACH SERVICES ---- */

type OutreachServices struct {
	Config *config.Config
	DB     *gorm.DB
	Cron   *cron.Cron
	Clock  *time.Ticker
}
