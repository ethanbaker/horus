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
	// Delegate to specific init functions
	if err := NotionInit(config); err != nil {
		return err
	}

	return nil
}

/* ---- CONSTANTS ---- */

// A map of enabled functions in the submodule. New messages can be
// called using a key in this map
var enabledFunctions = map[string]func(*config.Config, map[string]any) string{
	"ping":                Ping,
	"notion-daily-digest": NotionDailyDigest,
	"notion-read-page":    NotionReadPage,
}

/* ---- MESSAGE ---- */

// StaticOutreachMessage represents a statically sent message to the user. Static messages
// repeat on a given interval supplied by a CRON string by calling a custom-built
// function
type StaticOutreachMessage struct {
	function func(*config.Config, map[string]any) string // The function that will be called repeatedly
	channels []chan string                               // Channels that Horus will send the response to
	repeat   string                                      // The cron string used to repeat calls

	cronjob *cron.Cron     // The cronjob representing the repeating function (in case we want to end the cronjob)
	cronID  cron.EntryID   // The ID of the cronjob representing this message
	stopped bool           // Whether or not the task is stopped
	data    map[string]any // Data passed in from the creation of the message
}

func (m *StaticOutreachMessage) GetContent(c *config.Config) string {
	if m.stopped {
		return ""
	}

	return m.function(c, m.data)
}

func (m *StaticOutreachMessage) GetChannels() []chan string {
	return m.channels
}

func (m *StaticOutreachMessage) Start() error {
	// Make sure the cronjob is not nil
	if m.cronjob == nil {
		return fmt.Errorf("message has been deleted")
	}

	// Make sure the message is stopped
	if !m.stopped {
		return fmt.Errorf("message is still active")
	}

	m.stopped = false
	return nil
}

func (m *StaticOutreachMessage) Stop() error {
	// Make sure the cronjob is not nil
	if m.cronjob == nil {
		return fmt.Errorf("message has been deleted")
	}

	// Make sure the message has not already been stopped
	if m.stopped {
		return fmt.Errorf("message is already stopped")
	}

	// Stop the message
	m.stopped = true
	return nil
}

func (m *StaticOutreachMessage) Delete() error {
	// Make sure the cronjob is not nil
	if m.cronjob == nil {
		return fmt.Errorf("message has been deleted")
	}

	// Stop the message
	if err := m.Stop(); err != nil {
		return err
	}

	// Remove the cronjob
	m.cronjob.Remove(m.cronID)
	m.cronjob = nil
	return nil
}

// New creates a new static message (we don't save these to a DB because they're statically called in implementations)
func New(manager *types.OutreachManager, chans []chan string, data map[string]any) (types.OutreachMessage, error) {
	// Data should have a function name for us to get
	label, ok := data["function"].(string)
	if !ok {
		return nil, fmt.Errorf("data['function'] field is not present")
	}

	// Get the function pointer from the provided name
	f, ok := enabledFunctions[label]
	if !ok {
		return nil, fmt.Errorf("function with label '%v' is not present", label)
	}

	// Data should have a repeat string for us to get
	repeat, ok := data["repeat"].(string)
	if !ok {
		return nil, fmt.Errorf("data['repeat'] field is not present")
	}

	// Create the message
	m := StaticOutreachMessage{
		function: f,
		repeat:   repeat,
		channels: chans,
		stopped:  false,
		data:     data,
	}

	// Create a new cron job to send the message every interval
	m.cronjob = manager.Services.Cron
	cronID, err := m.cronjob.AddFunc(m.repeat, func() {
		// Get the content of the message once
		content := m.GetContent(manager.Services.Config)

		// Send the content through to each channel
		for _, c := range m.GetChannels() {
			c <- content
		}
	})
	m.cronID = cronID

	return &m, err
}
