// Handle any outreach messages
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/horus/outreach"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

// Setup Outreach functionality
func setupOutreach(s *discordgo.Session) error {
	var err error

	// Setup outreach
	if err = outreach.Setup(config.FormatDSN()); err != nil {
		return err
	}

	var ch chan string
	if ch, err = outreach.AddChannel("discord"); err != nil {
		return err
	}

	// On new content, send it to the user
	go onOutreach(s, ch)

	// Read in outreach config
	yamlFile, err := os.ReadFile(os.Getenv("BASE_PATH") + os.Getenv("OUTREACH_CONFIG"))
	if err != nil {
		return err
	}

	var outreachConfig types.OutreachConfig
	if err = yaml.Unmarshal(yamlFile, &outreachConfig); err != nil {
		return err
	}

	// Add the static outreaches if they match this implementation key
	for _, msg := range outreachConfig.Static {
		// Only add keys that contain discord
		if msg.Key != "discord" {
			continue
		}

		// Add the outreach
		_, err = outreach.New("static", []types.OutreachMethod{types.Discord}, types.StaticOutreach{
			Function: msg.Name,
			Repeat:   msg.Repeat,
		})

		if err != nil {
			return err
		}
	}

	// Add the dynamic outreaches if they match this implementation key
	for _, msg := range outreachConfig.Dynamic {
		// Only add keys that contain discord
		if msg.Key != "discord" {
			continue
		}

		// Add the outreach
		_, err = outreach.New("dynamic", []types.OutreachMethod{types.Discord}, types.DynamicOutreach{
			Function:        msg.Name,
			IntervalMinutes: time.Minute * time.Duration(msg.IntervalMinutes),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// Handle messages that should be sent to a user
func onOutreach(s *discordgo.Session, ch chan string) {
	// Open a channel to the user
	channel, err := s.UserChannelCreate(os.Getenv("DISCORD_USER_ID"))
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error opening up user channel (err: %v)\n", err)
	}

	BOT_OPEN_CHANNELS = append(BOT_OPEN_CHANNELS, channel.ID)

	for {
		content := <-ch
		log.Printf("[STATUS]: New outreach message (%v)\n", content)

		// Send the user a message
		_, err = s.ChannelMessageSend(channel.ID, format(content))
		if err != nil {
			log.Printf("[ERROR]: In discord, error sending user message (err: %v)\n", err)
			continue
		}

		// Always create a new conversation when an outreach message appears
		currentConversation[channel.ID] = &ChannelInfo{
			Name:            fmt.Sprintf("discord-%v-%v", channel.ID, time.Now().UTC().Unix()),
			LastMessageTime: time.Now().UTC(),
		}
		name := currentConversation[channel.ID].Name

		// Make sure the conversation exists
		if !bot.IsConversation(name) {
			if err := bot.AddConversation(name); err != nil {
				log.Printf("[ERROR]: In discord, error sending creating conversation (err: %v)\n", err)
				continue
			}
		}

		// Add the outreach message to the conversation
		if err := bot.AddMessage(name, openai.ChatMessageRoleAssistant, "", content); err != nil {
			log.Printf("[ERROR]: In discord, error adding message to conversation (err: %v)\n", err)
			continue
		}
	}
}
