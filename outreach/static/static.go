package static

import (
	"fmt"

	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/robfig/cron/v3"
)

/* ---- INIT ---- */

// Run Init functions
func Init(config *config.Config) error {
	if err := NotionInit(config); err != nil {
		return err
	}

	return nil
}

/* ---- CONSTANTS ---- */

// A map of enabled functions in the submodule. New messages can be
// called using a key in this map
var enabledFunctions = map[string]func(*config.Config) string{
	"ping":                        Ping,
	"notion-daily-digest":         NotionDailyDigest,
	"notion-night-affirmations":   NotionNightAffirmations,
	"notion-morning-affirmations": NotionMorningAffirmations,
}

/* ---- MESSAGE ---- */

// StaticOutreachMessage represents a statically sent message to the user. Static messages
// repeat on a given interval supplied by a CRON string by calling a custom-built
// function
type StaticOutreachMessage struct {
	Function func(*config.Config) string // The function that will be called repeatedly
	Channels []chan string               // Channels that Horus will send the response to

	Repeat string // The cron string used to repeat calls

	cronjob *cron.Cron `gorm:"-"` // The cronjob representing the repeating function (in case we want to end the cronjob)
	stopped bool       `gorm:"-"` // Whether or not the task is stopped
}

func (m *StaticOutreachMessage) GetContent(c *config.Config) string {
	return m.Function(c)
}

func (m *StaticOutreachMessage) GetChannels() []chan string {
	return m.Channels
}

func (m *StaticOutreachMessage) Start() error {
	// Make sure the message is stopped
	if !m.stopped {
		return fmt.Errorf("message is still active")
	}

	// Make sure the cronjob is not nil
	if m.cronjob == nil {
		return fmt.Errorf("cron is nil")
	}

	m.stopped = false
	m.cronjob.Start()
	return nil
}

func (m *StaticOutreachMessage) Stop() error {
	// Make sure the message has not already been stopped
	if m.stopped {
		return fmt.Errorf("message is already stopped")
	}
	// Make sure the cronjob is not nil
	if m.cronjob == nil {
		return fmt.Errorf("cron is nil")
	}

	// Stop the message
	m.stopped = true
	return m.cronjob.Stop().Err()
}

func (m *StaticOutreachMessage) Delete() error {
	// Make sure the cronjob is not nil
	if m.cronjob == nil {
		return fmt.Errorf("cron is nil")
	}

	// Stop the message
	if err := m.Stop(); err != nil {
		return err
	}

	m.cronjob = nil
	return nil
}

// New creates a new static message (we don't save these to a DB because they're statically called in implementations)
func New(services *types.OutreachServices, channels []chan string, raw any) (types.OutreachMessage, error) {
	// Run the init functions
	if err := Init(services.Config); err != nil {
		return nil, err
	}

	// Cast the raw input data
	data, ok := raw.(types.StaticOutreach)
	if !ok {
		return nil, fmt.Errorf("cannot properly cast input data to StaticOutreach type")
	}

	// Find the function key we want to run
	f, ok := enabledFunctions[data.Function]
	if !ok {
		return nil, fmt.Errorf("function with key %v is not present", data.Function)
	}

	// Create the message
	m := StaticOutreachMessage{
		Function: f,
		Repeat:   data.Repeat,
		Channels: channels,
		stopped:  false,
	}

	// Create a new cron job to send the message every interval
	m.cronjob = services.Cron
	_, err := m.cronjob.AddFunc(m.Repeat, func() {
		// Get the content of the message once
		content := m.GetContent(services.Config)

		// Send the content through to each channel
		for _, c := range m.GetChannels() {
			c <- content
		}
	})

	return &m, err
}
