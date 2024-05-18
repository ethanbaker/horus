package outreach

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Timed is a Message that only occurs once
type Timed struct {
	gorm.Model

	Content   string    `json:"content"` // The content of the message
	Methods   []Method  `json:"methods"` // The methods used to contact the user
	Timestamp time.Time `json:"time"`    // The time to send the outreach

	timer     *time.Timer `json:"-" gorm:"-"` // A timer to allow the message to be cancelled
	isExpired bool        `json:"-" gorm:"-"` // Whether or not the message has expired
}

func (m *Timed) GetContent() string {
	return m.Content
}

func (m *Timed) GetMethods() []Method {
	return m.Methods
}

func (m *Timed) Stop() error {
	// Make sure the timer is not expired
	if m.isExpired {
		return fmt.Errorf("cannot stop exired message")
	}
	// Make sure the timer is not nil
	if m.timer == nil {
		return fmt.Errorf("timer is null")
	}

	// Stop the timer
	if !m.timer.Stop() {
		return fmt.Errorf("unable to stop timer")
	}

	return nil
}

// Create a new message
func NewTimer(content string, methods []Method, timestamp time.Time) Message {
	// Create the message struct
	m := Timed{
		Content:   content,
		Methods:   methods,
		Timestamp: timestamp,
		isExpired: false,
	}

	m.timer = time.AfterFunc(m.Timestamp.UTC().Sub(time.Now().UTC()), func() {
		// Keep trying to send the message on error
		err := fmt.Errorf("")
		for err != nil {
			err = Send(&m)
		}
		m.isExpired = true
	})

	return &m
}
