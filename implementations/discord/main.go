package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	horus "github.com/ethanbaker/horus/bot"
	module_ambient "github.com/ethanbaker/horus/bot/module_ambient"
	module_config "github.com/ethanbaker/horus/bot/module_config"
	module_keepass "github.com/ethanbaker/horus/bot/module_keepass"
	"github.com/ethanbaker/horus/outreach"
	"github.com/ethanbaker/horus/utils/types"
	mysql_driver "github.com/go-sql-driver/mysql"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

// TODO: instead of calling stuff through a bot, call it through an API

/* -------- TYPES -------- */

type ChannelInfo struct {
	Name            string
	LastMessageTime time.Time
}

/* -------- CONSTANTS -------- */

// Discord credentials
var (
	APP_ID        string = os.Getenv("DISCORD_APP_ID")
	PUBKEY        string = os.Getenv("DISCORD_PUBKEY")
	CLIENT_ID     string = os.Getenv("DISCORD_CLIENT_ID")
	CLIENT_SECRET string = os.Getenv("DISCORD_CLIENT_SECRET")
	TOKEN         string = os.Getenv("DISCORD_TOKEN")

	GUILD_ID            string   = os.Getenv("DISCORD_GUILD_ID")
	BOT_OPEN_CHANNELS   []string = strings.Split(os.Getenv("DISCORD_BOT_OPEN_CHANNELS"), ",")
	BOT_THREAD_CHANNELS []string = strings.Split(os.Getenv("DISCORD_BOT_THREAD_CHANNELS"), ",")
)

// SQL config
var config = mysql_driver.Config{
	User:      os.Getenv("SQL_USER"),
	Passwd:    os.Getenv("SQL_PASSWD"),
	Net:       os.Getenv("SQL_NET"),
	Addr:      os.Getenv("SQL_ADDR"),
	DBName:    os.Getenv("SQL_DBNAME"),
	ParseTime: true,
	Loc:       time.Local,
}

// How long until a new conversation begins in bot channels
const BOT_CHANNEL_OFFSET = 6 * time.Hour

/* -------- GLOBALS -------- */

// The OpenAI client
var client *openai.Client

// The Horus bot
var bot *horus.Bot

// The current conversation in each bot channel
var currentConversation = make(map[string]*ChannelInfo)

/* ------------------ FUNCTIONS ------------------ */

// main starts the discord bot
func main() {
	var err error

	// Update the current conversation list
	for _, channel := range BOT_OPEN_CHANNELS {
		currentConversation[channel] = &ChannelInfo{
			Name:            fmt.Sprintf("discord-%v-%v", channel, time.Now().UTC().Unix()),
			LastMessageTime: time.Now().UTC(),
		}
	}

	// Initialize the SQl
	if err = horus.InitSQL(config.FormatDSN()); err != nil {
		log.Fatal(err)
	}

	// Create the OpenAI client
	client = openai.NewClient(os.Getenv("OPENAI_TOKEN"))

	// Try to get a bot that we've already created
	bot, err = horus.GetBotByName("horus-main")
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error getting horus bot (err: %v)\n", err)
	}

	// If the bot is nil, we need to create one
	if bot == nil {
		bot, err = horus.NewBot("horus-main", horus.PERMISSIONS_ALL)
		if err != nil {
			log.Fatalf("[ERROR]: In discord, error making horus bot (err: %v)\n", err)
		}
	}

	// Setup the bot
	module_ambient.NewModule(bot, true)
	module_config.NewModule(bot, true)
	module_keepass.NewModule(bot, true)
	bot.Setup(client)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error creating Discord session (err: %v)\n", err)
	}

	// Add handlers
	dg.AddHandler(onMessageCreate)
	dg.AddHandler(onThreadMessageCreate)
	dg.AddHandler(onCommand)

	// Add custom commands
	_, err = dg.ApplicationCommandBulkOverwrite(APP_ID, GUILD_ID, []*discordgo.ApplicationCommand{
		{
			Name:        "conversation",
			Description: "Create a new conversation with Horus",
		},
	})
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error creating Discord commands (err: %v)\n", err)
	}

	// Set the intents for what the bot will do
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	dg.Identify.Intents = discordgo.IntentsDirectMessages

	// Setup outreach
	if err = outreach.Setup(config.FormatDSN()); err != nil {
		log.Fatalf("[ERROR]: In discord, error initalizing db for outreach (err: %v)\n", err)
	}

	var ch chan string
	if ch, err = outreach.AddChannel("discord"); err != nil {
		log.Fatalf("[ERROR]: In discord, error adding channel to outreach (err: %v)\n", err)
	}

	// On new content, send it to the user
	go onOutreach(dg, ch)

	// Read in outreach config
	yamlFile, err := os.ReadFile(os.Getenv("BASE_PATH") + os.Getenv("OUTREACH_CONFIG"))
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error opening outreach config (err: %v)\n", err)
	}

	var outreachConfig types.OutreachConfig
	if err = yaml.Unmarshal(yamlFile, &outreachConfig); err != nil {
		log.Fatalf("[ERROR]: In discord, error opening outreach config (err: %v)\n", err)
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
			log.Fatalf("[ERROR]: In discord, error setting up outreach (err: %v)\n", err)
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
			log.Fatalf("[ERROR]: In discord, error setting up outreach (err: %v)\n", err)
		}
	}

	// Open a websocket connection to discord
	err = dg.Open()
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error opening connection (err: %v)\n", err)
	}

	// Update the status of the bot
	//dg.UpdateGameStatus(0, "Type '!help' for help")

	// Make a channel to wait for an interrupt signal (keep the bot running)
	log.Println("[STATUS]: Discord is now running!  (Press CTRL-C to exit)")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Close the discord session
	dg.Close()
}

// onMessageCreate function handles any message sent in a bot-specific channel
func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages not in one of the specified bot channels
	valid := false
	for _, id := range BOT_OPEN_CHANNELS {
		if m.ChannelID == id {
			valid = true
			break
		}
	}

	// Ignore all messages created by the bot itself, messages in threads and messages with 0 length
	if ch, err := s.State.Channel(m.ChannelID); err == nil || (ch != nil && ch.IsThread()) || m.Author.ID == s.State.User.ID || len(m.Content) == 0 || !valid {
		return
	}

	// Determine what conversation this message should belong to
	var name string
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
	output := format(resp.Message)

	// Send the output
	s.ChannelMessageSend(m.ChannelID, output)
}

// onThreadMessageCreate function handles any message sent in threads
func onThreadMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself, messages not in threads and messages with 0 length
	if ch, err := s.State.Channel(m.ChannelID); err == nil || (ch != nil && !ch.IsThread()) || m.Author.ID == s.State.User.ID || len(m.Content) == 0 {
		return
	}

	// Determine what conversation this message should belong to
	name := fmt.Sprintf("discord-" + m.ChannelID)

	// Make sure the conversation exists
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

func onCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Find the name of the command
	data := i.ApplicationCommandData()
	switch data.Name {
	case "conversation":
		// Ignore commands not in one of the specified bot channels
		valid := false
		for _, id := range BOT_THREAD_CHANNELS {
			if i.ChannelID == id {
				valid = true
				break
			}
		}
		if !valid {
			return
		}

		// Make sure the channel is not a thread
		if ch, err := s.State.Channel(i.ChannelID); err == nil || (ch != nil && ch.IsThread()) {
			return
		}

		// Create a new message to host the thread
		err := s.InteractionRespond(
			i.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hello! What do you need help with today?",
				},
			},
		)
		if err != nil {
			s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred: >>> %v\n", err.Error()))
			return
		}

		// Get the response ID
		m, err := s.InteractionResponse(i.Interaction)
		if err != nil {
			s.ChannelMessageSend(i.ChannelID, fmt.Sprintf("Sorry, an error occurred: >>> %v\n", err.Error()))
			return
		}

		// Create a new thread
		thread, err := s.MessageThreadStartComplex(i.ChannelID, m.ID, &discordgo.ThreadStart{
			Name:      "New Conversation",
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

// Format an output string to match discord messages
func format(output string) string {
	output = strings.ReplaceAll(output, "<STRONG>", "**")
	output = strings.ReplaceAll(output, "<EM>", "*")
	output = strings.ReplaceAll(output, "<EM>", "*")
	output = strings.ReplaceAll(output, "<INS>", "__")
	output = strings.ReplaceAll(output, "<DEL>", "~~")
	output = strings.ReplaceAll(output, "<DEL>", "~~")
	output = strings.ReplaceAll(output, "<BLOCKQUOTE_IN>", "> ")
	output = strings.ReplaceAll(output, "<BLOCKQUOTE>", ">>> ")
	output = strings.ReplaceAll(output, "<CODE_IN>", "`")
	output = strings.ReplaceAll(output, "<CODE>", "```")
	output = strings.ReplaceAll(output, "<SPOILER>", "||")

	return output
}
