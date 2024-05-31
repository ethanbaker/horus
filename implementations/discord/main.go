package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	horus "github.com/ethanbaker/horus/bot"
	module_ambient "github.com/ethanbaker/horus/bot/module_ambient"
	module_config "github.com/ethanbaker/horus/bot/module_config"
	module_keepass "github.com/ethanbaker/horus/bot/module_keepass"
	"github.com/ethanbaker/horus/utils/config"
)

/* -------- GLOBALS -------- */

// The Horus bot
// TODO: instead of calling stuff through a bot, call it through an API
var bot *horus.Bot

// The current conversation in each bot channel
var currentConversation = make(map[string]*ChannelInfo)

// The config the bot is running off of
var cfg *config.Config

/* ------------------ MAIN ------------------ */

// main starts the discord bot
func main() {
	var err error

	// Get the path to read (default to 'config/.env' on no environment variable)
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/.env"
	}

	// Initialize the config
	conf, errs := config.NewConfigFromFile(path)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("[ERROR]: In discord, error from config (%v)\n", err)
		}
		log.Fatalf("[ERROR]: In discord, error reading config. Failing\n")
	}
	cfg = conf

	// Initalize the SQL databases
	if err = horus.InitSQL(cfg); err != nil {
		log.Fatalf("[ERROR]: In discord, error initalizing sql (%v)\n", err)
	}

	// Initialize the current conversation list to begin conversations now
	for _, channel := range cfg.DiscordOpenChannels {
		currentConversation[channel] = &ChannelInfo{
			Name:            fmt.Sprintf("discord-%v-%v", channel, time.Now().UTC().Unix()),
			LastMessageTime: time.Now().UTC(),
		}
	}

	// Try to get a bot that we've already created
	bot, err = horus.GetBotByName("horus-main", cfg)
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error getting horus bot (err: %v)\n", err)
	}

	// If the bot is nil, we need to create one
	if bot == nil {
		bot, err = horus.NewBot("horus-main", horus.PERMISSIONS_ALL, cfg)
		if err != nil {
			log.Fatalf("[ERROR]: In discord, error making horus bot (err: %v)\n", err)
		}
	}
	log.Println("[STATUS]: Successfully retrieved bot")

	// Setup the bot
	module_ambient.NewModule(bot, true)
	module_config.NewModule(bot, true)
	module_keepass.NewModule(bot, true)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + cfg.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error creating Discord session (err: %v)\n", err)
	}

	// Add handlers
	dg.AddHandler(onMessageCreate)
	dg.AddHandler(onThreadMessageCreate)
	dg.AddHandler(onCommand)

	// Open a websocket connection to discord
	err = dg.Open()
	if err != nil {
		log.Fatalf("[ERROR]: In discord, error opening connection (err: %v)\n", err)
	}
	log.Println("[STATUS]: Successfully connected to discord")

	// Add commands
	if err := addCommands(dg, cfg); err != nil {
		log.Fatalf("[ERROR]: In discord, error creating Discord commands (err: %v)\n", err)
	}

	// Set the intents for what the bot will do
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	// Setup outreach
	if err := setupOutreach(dg, cfg); err != nil {
		log.Fatalf("[ERROR]: In discord, error setting up outreach (err: %v)\n", err)
	}
	log.Println("[STATUS]: Successfully set up outreach")

	// Update the status of the bot
	//dg.UpdateGameStatus(0, "Type '!help' for help")

	// Make a channel to wait for an interrupt signal (keep the bot running)
	log.Println("[STATUS]: Discord is now running!  (Press CTRL-C to exit)")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Close the discord session
	dg.Close()
}
