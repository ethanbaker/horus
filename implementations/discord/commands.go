// Handle commands for Horus
package main

import (
	"fmt"
	"log"

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
