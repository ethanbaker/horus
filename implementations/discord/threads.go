// Handle messages being sent to threads
package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/horus/utils/types"
)

// onThreadMessageCreate function handles any message sent in threads
func onThreadMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages where:
	//  - Cannot get the associated channel
	//  - Channel is not equal to nil and channel is not a thread
	//  - The bot sent this message
	//  - The length of the content is zero
	//  - The message type isn't a default message
	if ch, err := s.State.Channel(m.ChannelID); err != nil ||
		(ch != nil && !ch.IsThread()) ||
		m.Author.ID == s.State.User.ID ||
		len(m.Content) == 0 ||
		m.Type != discordgo.MessageTypeDefault {
		return
	}
	log.Printf("[STATUS]: New thread message (%v)\n", m.Content)

	// Determine what conversation this message should belong to using the thread ID
	name := fmt.Sprintf("discord-" + m.ChannelID)

	// Make sure the conversation exists. If it doesn't then Horus didn't create the thread and we shouldn't speak in it
	if !bot.IsConversation(name) {
		return
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
	output := format(resp.Message)

	// Send the output
	s.ChannelMessageSend(m.ChannelID, output)
}
