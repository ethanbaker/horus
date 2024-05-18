package dynamic

import (
	"fmt"
	"log"
	"time"

	"github.com/ethanbaker/horus/utils/types"
)

/* ---- INIT ---- */

// Run init functions
func Init() error {
	return nil
}

/* ---- CONSTANTS ---- */

// A map of enabled functions in the submodule. New messages can be
// called using a key in this map
var enabledFunctions = map[string]func(*DynamicOutreachMessage, time.Time) string{
	"notion-schedule-reminders": NotionScheduleReminders,
}

// A map of associated update functions for each function key
var updateFunctions = map[string]func(*DynamicOutreachMessage) error{
	"notion-schedule-reminders": NotionScheduleRemindersUpdate,
}

/* ---- MESSAGE ---- */

// DynamicOutreachMessage represents a message that gets sent to the user at
// any randomly determined time based on dynamic content. The message pulls
// dynamic content from a source and then sends messages to the user at any
// given moment
type DynamicOutreachMessage struct {
	Function func(*DynamicOutreachMessage, time.Time) string // The function that will be called repeatedly
	Channels []chan string                                   // Channels that Horus will send the response to

	Update   func(*DynamicOutreachMessage) error // The function to update dynamic content
	Interval time.Duration                       // How long until successive updates

	stopChan chan bool `gorm:"-"` // Used to stop the message
	stopped  bool      `gorm:"-"` // Whether or not the task is stopped

	intervalTicker *time.Ticker `gorm:"-"` // Used to repeatedly call Update
	clock          *time.Ticker `gorm:"-"` // Used to send messages to channels
	data           any          `gorm:"-"` // The dynamic content this message is getting
}

func (m *DynamicOutreachMessage) GetContent() string {
	return m.Function(m, time.Now())
}

func (m *DynamicOutreachMessage) GetChannels() []chan string {
	return m.Channels
}

func (m *DynamicOutreachMessage) Start() error {
	// Make sure the message is stopped
	if !m.stopped {
		return fmt.Errorf("message is still active")
	}

	// Just don't do anything here until I need to
	return nil
}

func (m *DynamicOutreachMessage) Stop() error {
	// Make sure the message has not already been stopped
	if m.stopped {
		return fmt.Errorf("message is already stopped")
	}

	m.stopChan <- true
	return nil
}

func (m *DynamicOutreachMessage) Delete() error {
	// Stop the message
	if err := m.Stop(); err != nil {
		return err
	}

	// Delete the message's data
	m.data = nil
	return nil
}

// New creates a new dynamic message (we don't save these to a DB because they're statically called in implementations)
func New(services *types.OutreachServices, channels []chan string, raw any) (types.OutreachMessage, error) {
	// Cast the raw input data
	data, ok := raw.(types.DynamicOutreach)
	if !ok {
		return nil, fmt.Errorf("cannot properly cast input data to DynamicOutreach type")
	}

	// Find the function key we want to run
	f, ok := enabledFunctions[data.Function]
	if !ok {
		return nil, fmt.Errorf("function with key %v is not present", data.Function)
	}

	// Find the associated update function
	u, ok := updateFunctions[data.Function]
	if !ok {
		return nil, fmt.Errorf("update function with key %v is not present", data.Function)
	}

	// Create the message
	m := DynamicOutreachMessage{
		Function: f,
		Update:   u,
		Channels: channels,
		Interval: data.IntervalMinutes,
		stopChan: make(chan bool),
		stopped:  false,
		clock:    services.Clock,
	}

	// Update the message right now
	if err := m.Update(&m); err != nil {
		return nil, err
	}

	// Every interval minutes fetch dynamic content
	m.intervalTicker = time.NewTicker(m.Interval)
	go func() {
		for {
			select {
			// When the interval ticker goes off, call the update function
			case <-m.intervalTicker.C:
				// Update the dynamic content using the update function
				if err := m.Update(&m); err != nil {
					log.Printf("[ERROR]: in dynamic/%v, error updating content (err: %v)\n", data.Function, err)
				}

			case <-m.clock.C:
				// Check if the user should be sent content
				if content := m.GetContent(); content != "" {
					for _, c := range m.GetChannels() {
						c <- content
					}
				}

			case <-m.stopChan:
				return
			}
		}
	}()
	return &m, nil
}
