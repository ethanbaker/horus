// Handle commands for Horus
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/horus/utils/config"
)

// Add commands to the discord bot
func addCommands(s *discordgo.Session, cfg *config.Config) error {
	// Add custom commands
	_, err := s.ApplicationCommandBulkOverwrite(cfg.Getenv("DISCORD_APP_ID"), cfg.Getenv("DISCORD_GUILD_ID"), []*discordgo.ApplicationCommand{
		{
			Name:        "conversation",
			Description: "Create a new conversation with Horus",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Name of the conversation",
					Required:    false,
				},
			},
		},
		{
			Name:        "purge",
			Description: "Purge messages in a given channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Number of messages to delete",
					Required:    true,
				},
			},
		},
	})

	return err
}

// onCommand handles a command being called for Horus
func onCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Find the name of the command
	data := i.ApplicationCommandData()
	log.Printf("[STATUS]: New thread command (%v)\n", data.Name)

	// Delegate to submethods based on the command name
	switch data.Name {
	case "conversation":
		onCommandConversation(s, i)

	case "purge":
		onCommandPurge(s, i)
	}
}

// Handle the 'Conversation' command
func onCommandConversation(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Determine whether or not this command was called from an allowed thread-making channel
	valid := false
	for _, id := range cfg.DiscordThreadChannels {
		if i.ChannelID == id {
			valid = true
			break
		}
	}

	// Ignore commands where:
	//  - Cannot get the associated channel
	//  - Channel is not equal to nil and channel is a thread
	//  - The message isn't in an allowed thread channel
	if ch, err := s.State.Channel(i.ChannelID); err != nil ||
		(ch != nil && ch.IsThread()) ||
		!valid {
		return
	}

	// Access options from the command by putting them all into a map
	optionsData := i.ApplicationCommandData().Options
	options := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(optionsData))
	for _, opt := range optionsData {
		options[opt.Name] = opt
	}

	// Create a new message to host the thread
	if err := s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hello! What do you need help with today?",
			},
		},
	); err != nil {
		s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred: >>> %v\n", err.Error()))
		return
	}

	// Get the response ID
	m, err := s.InteractionResponse(i.Interaction)
	if err != nil {
		s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred: >>> %v\n", err.Error()))
		return
	}

	// Find the name of the thread
	name := "New Conversation"
	if provided, ok := options["name"]; ok {
		name = provided.StringValue()
	}

	// Create a new thread
	thread, err := s.MessageThreadStartComplex(i.ChannelID, m.ID, &discordgo.ThreadStart{
		Name:      name,
		Invitable: false,
	})
	if err != nil {
		s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred:\n >>> %v\n", err.Error()))
		return
	}

	// Register the thread to the bot
	if err := bot.AddConversation("discord-" + thread.ID); err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, an error occurred: >>> %v\n", err.Error()))
	}
}

// Handle the 'purge' command
func onCommandPurge(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Access options from the command by putting them all into a map
	optionsData := i.ApplicationCommandData().Options
	options := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(optionsData))
	for _, opt := range optionsData {
		options[opt.Name] = opt
	}

	errMsg := "Purging messages"

	// Get the count of messages to delete
	rawCount, ok := options["count"]
	if !ok {
		errMsg = "Invalid count provided, please try again"
	}

	count := int(rawCount.IntValue())
	if count < 0 {
		errMsg = "Count cannot be negative, please try again"
	}
	count++

	// Send the interaction response
	if err := s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: errMsg,
			},
		},
	); err != nil {
		s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred creating the interaction response: >>> %v\n", err.Error()))
		return
	}

	// Purge messages
	var lastMessageID string
	twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)
	for count > 0 {
		// Fetch messages
		limit := count
		if limit > 100 {
			limit = 100
		}

		messages, err := s.ChannelMessages(i.ChannelID, limit, lastMessageID, "", "")
		if err != nil {
			s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred while fetching messages: >>> %v\n", err.Error()))
			return
		}

		// Separate messages into recent and old
		var recentMessageIDs []string
		for _, message := range messages {
			if message.Timestamp.Before(twoWeeksAgo) {
				// Delete old messages one by one
				if err := s.ChannelMessageDelete(i.ChannelID, message.ID); err != nil {
					s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred while fetching messages: >>> %v\n", err.Error()))
					return
				}
			} else {
				// Collect IDs of recent messages
				recentMessageIDs = append(recentMessageIDs, message.ID)
			}
		}

		// Bulk delete recent messages if there are any
		if len(recentMessageIDs) > 1 {
			err = s.ChannelMessagesBulkDelete(i.ChannelID, recentMessageIDs)
			if err != nil {
				s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred while bulk deleting messages: >>> %v\n", err.Error()))
				return
			}
		} else if len(recentMessageIDs) == 1 {
			// If there's only one recent message, use the single delete method
			err = s.ChannelMessageDelete(i.ChannelID, recentMessageIDs[0])
			if err != nil {
				s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred while deleting a single message: >>> %v\n", err.Error()))
				return
			}
		}

		// Update the last message ID for the next batch
		if len(messages) > 0 {
			lastMessageID = messages[len(messages)-1].ID
		}

		// Decrement the count
		count -= len(messages)
	}
}
