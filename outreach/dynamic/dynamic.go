package dynamic

import (
	"fmt"
	"log"
	"time"

	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
)

/* ---- INIT ---- */

// Run init functions
func Init(config *config.Config) error {
	if err := NotionInit(config); err != nil {
		return err
	}

	return nil
}

/* ---- CONSTANTS ---- */

// A map of enabled functions in the submodule. New messages can be
// called using a key in this map
var enabledFunctions = map[string]func(*config.Config, *DynamicOutreachMessage, time.Time) string{
	"notion-schedule-reminders": NotionScheduleReminders,
}

// A map of associated update functions for each function key
var updateFunctions = map[string]func(*config.Config, *DynamicOutreachMessage) error{
	"notion-schedule-reminders": NotionScheduleRemindersUpdate,
}

/* ---- MESSAGE ---- */

// DynamicOutreachMessage represents a message that gets sent to the user at
// any randomly determined time based on dynamic content. The message pulls
// dynamic content from a source and then sends messages to the user at any
// given moment
type DynamicOutreachMessage struct {
	function func(*config.Config, *DynamicOutreachMessage, time.Time) string // The function that will be called repeatedly
	channels []chan string                                                   // Channels that Horus will send the response to
	update   func(*config.Config, *DynamicOutreachMessage) error             // The function to update dynamic content
	interval time.Duration                                                   // How long until successive updates

	stopped bool      // Whether or not the message is stopped
	deleted chan bool // Used to delete the message

	intervalTicker *time.Ticker // Used to repeatedly call Update
	clock          *time.Ticker // Used to send messages to channels
	data           any          // The dynamic content this message is getting
}

func (m *DynamicOutreachMessage) GetContent(c *config.Config) string {
	if m.stopped {
		return ""
	}

	return m.function(c, m, time.Now())
}

func (m *DynamicOutreachMessage) GetChannels() []chan string {
	return m.channels
}

func (m *DynamicOutreachMessage) Start() error {
	// Make sure the message hasn't been deleted
	if m.clock == nil {
		return fmt.Errorf("message has been deleted")
	}

	// Make sure the message is stopped
	if !m.stopped {
		return fmt.Errorf("message is still active")
	}

	m.stopped = false
	return nil
}

func (m *DynamicOutreachMessage) Stop() error {
	// Make sure the message hasn't been deleted
	if m.clock == nil {
		return fmt.Errorf("message has been deleted")
	}

	// Make sure the message has not already been stopped
	if m.stopped {
		return fmt.Errorf("message is already stopped")
	}

	m.stopped = true
	return nil
}

func (m *DynamicOutreachMessage) Delete() error {
	// Stop the message
	if err := m.Stop(); err != nil {
		return err
	}

	// Delete the message
	m.deleted <- true
	m.clock = nil
	return nil
}

// New creates a new dynamic message (we don't save these to a DB because they're statically called in implementations)
func New(manager *types.OutreachManager, chans []chan string, data map[string]any) (types.OutreachMessage, error) {
	// Data should have a function name for us to get
	label, ok := data["function"].(string)
	if !ok {
		return nil, fmt.Errorf("data['function'] field is not present")
	}

	// Get the function pointer for getting content and updates from the provided name
	f, ok := enabledFunctions[label]
	if !ok {
		return nil, fmt.Errorf("function with label '%v' is not present", label)
	}

	u, ok := updateFunctions[label]
	if !ok {
		return nil, fmt.Errorf("update function with label '%v' is not present", label)
	}

	// Data should have an interval count for us to get
	interval, ok := data["interval"].(int)
	if !ok {
		return nil, fmt.Errorf("data['interval'] field is not present")
	}

	// Create the message
	m := DynamicOutreachMessage{
		function: f,
		update:   u,
		channels: chans,
		interval: time.Minute * time.Duration(interval),
		deleted:  make(chan bool),
		stopped:  false,
		clock:    manager.Services.Clock,
	}

	// Update the message right now
	if err := m.update(manager.Services.Config, &m); err != nil {
		return nil, err
	}

	// Every interval minutes fetch dynamic content
	m.intervalTicker = time.NewTicker(m.interval)
	go func() {
		for {
			select {
			// When the interval ticker goes off, call the update function
			case <-m.intervalTicker.C:
				// Update the dynamic content using the update function
				if err := m.update(manager.Services.Config, &m); err != nil {
					log.Printf("[ERROR]: in dynamic/%v, error updating content (err: %v)\n", label, err)
				}

			case <-m.clock.C:
				// Check if the user should be sent content
				if content := m.GetContent(manager.Services.Config); content != "" {
					for _, c := range m.GetChannels() {
						c <- content
					}
				}

			case <-m.deleted:
				return
			}
		}
	}()
	return &m, nil
}
