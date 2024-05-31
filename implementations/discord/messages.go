// Handle normal messages being sent to open channels
package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/horus/utils/format"
	"github.com/ethanbaker/horus/utils/types"
)

// onMessageCreate function handles any message sent in a bot-specific channel
func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Determine if this message was sent in an open bot channel
	valid := false
	for _, id := range cfg.DiscordOpenChannels {
		if m.ChannelID == id {
			valid = true
			break
		}
	}

	// Ignore messages where:
	//  - Cannot get the associated channel
	//  - Channel is not equal to nil and channel is a thread
	//  - The bot sent this message
	//  - The length of the content is zero
	//  - The message type isn't a default message
	//  - The message isn't in an open bot channel
	if ch, err := s.State.Channel(m.ChannelID); err != nil ||
		(ch != nil && ch.IsThread()) ||
		m.Author.ID == s.State.User.ID ||
		len(m.Content) == 0 ||
		m.Type != discordgo.MessageTypeDefault ||
		!valid {
		return
	}
	log.Printf("[STATUS]: New message (%v)\n", m.Content)

	// Determine what conversation this message should belong to
	name := ""
	obj := currentConversation[m.ChannelID]
	if obj.LastMessageTime.Add(BOT_CHANNEL_OFFSET).Compare(time.Now().UTC()) >= 0 {
		// This message is in the same conversation
		name = obj.Name
	} else {
		// This message should be in a new conversation
		name = fmt.Sprintf("discord-%v-%v", m.ChannelID, time.Now().UTC().Unix())
		obj.Name = name
	}

	// Make sure the conversation exists
	if !bot.IsConversation(name) {
		if err := bot.AddConversation(name); err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, an error occurred: >>> %v\n", err.Error()))
			return
		}
	}

	// Send the message to the horus bot
	resp, err := bot.SendMessage(name, &types.Input{
		Message: m.Content,
	})

	// Print any errors if they occur
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, an error occurred:\n >>> %v\n", err.Error()))
		return
	} else if resp.Error != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, an error occurred: >>> %v\n", resp.Error.Error()))
		return
	}

	// Send files if present
	file, ok := resp.Data.(types.FileOutput)
	if ok {
		r := bytes.NewReader(file.Content)
		s.ChannelFileSend(m.ChannelID, file.Filename, r)
	}

	// Format the output
	output := format.FormatDiscord(resp.Message)

	// Send the output
	s.ChannelMessageSend(m.ChannelID, output)
}
