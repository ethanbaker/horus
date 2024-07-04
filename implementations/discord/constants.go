// Constants and types used throughout the implementation
package main

import (
	"time"
)

/* ---- TYPES ---- */

// ChannelInfo represents a discord channel. It holds the ID of the channel and the last message
// sent in the channel for determining if Horus should keep certain messages in its context
type ChannelInfo struct {
	Name            string
	LastMessageTime time.Time
}

/* -------- CONSTANTS -------- */

// How long until a new conversation begins in bot channels
const BOT_CHANNEL_OFFSET = 6 * time.Hour
