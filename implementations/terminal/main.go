// This file is a test-runner to directly test the horus 'Bot' object
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	horus "github.com/ethanbaker/horus/bot"
	module_ambient "github.com/ethanbaker/horus/bot/module_ambient"
	module_config "github.com/ethanbaker/horus/bot/module_config"
	module_keepass "github.com/ethanbaker/horus/bot/module_keepass"
	horus_config "github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
)

/* -------- CONSTANTS -------- */

const ASK_PROMPT = `Select an action:
1 - Add Conversation
2 - Remove Conversation
3 - Send Message
4 - Send Multiple Messages
5 - Quit
`

const INTERACTABLE = false

/* -------- GLOBALS -------- */

// The horus bot
var bot *horus.Bot

// The config for Horus to run off of
var config *horus_config.Config

// Scanner to read user input
var scanner *bufio.Scanner

/* -------- METHODS -------- */

func addConversation(name string) {
	// Create the new conversation
	if err := bot.AddConversation(name); err != nil {
		fmt.Printf("Error adding conversation: %v\n", err.Error())
	} else {
		fmt.Println("Successfully created conversation!")
	}
}

func removeConversation(name string) {
	// Create the new conversation
	if err := bot.DeleteConversation(name); err != nil {
		fmt.Printf("Error removing conversation: %v\n", err.Error())
	} else {
		fmt.Println("Successfully removed conversation!")
	}
}

func sendMessage(name string, content string) {
	// Send the message
	output, err := bot.SendMessage(name, &types.Input{
		Message: content,
	})

	if err != nil {
		fmt.Printf("Error sending message: %v\n", err.Error())
	} else {
		fmt.Println(output.Message)
	}
}

func sendMultiMessage(name string) {
	for content := ""; content != "stop"; content = scanner.Text() {
		// Send the message
		output, err := bot.SendMessage(name, &types.Input{
			Message: content,
		})

		if err != nil {
			fmt.Printf("Error sending message: %v\n", err.Error())
		} else {
			fmt.Println(output.Message)
		}

		scanner.Scan()
	}
}

/* -------- MAIN -------- */

func main() {
	var err error
	var errs []error

	// initialize the config
	config, errs = horus_config.NewConfigFromFile(".env")
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("[ERROR]: In terminal, error from config (%v)\n", err)
		}
		log.Fatalf("[ERROR]: In terminal, error reading config. Failing\n")
	}

	// Try to get a bot that we've already created
	bot, err = horus.GetBotByName("horus-testing", config)
	if err != nil {
		log.Fatal(err)
	}

	// If the bot is nil, we need to create one
	if bot == nil {
		bot, err = horus.NewBot("horus-testing", horus.PERMISSIONS_ALL, config)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup the bot
	module_ambient.NewModule(bot, true)
	module_config.NewModule(bot, true)
	module_keepass.NewModule(bot, true)

	// Read user input
	scanner = bufio.NewScanner(os.Stdin)
	for INTERACTABLE {
		fmt.Print(ASK_PROMPT)

		// Get the user's respose
		scanner.Scan()
		text := scanner.Text()
		fmt.Println()

		// Perform an action based on the user's choice
		switch text {
		case "1":
			// Get the conversation name
			fmt.Print("Enter the conversation's name: ")

			scanner.Scan()
			name := scanner.Text()

			addConversation(name)

		case "2":
			// Get the conversation name
			fmt.Print("Enter the conversation's name to delete: ")

			scanner.Scan()
			name := scanner.Text()

			removeConversation(name)

		case "3":
			// Get the conversation name to send a message to
			fmt.Print("Enter the conversation's name: ")

			scanner.Scan()
			name := scanner.Text()

			// Get the message to send to Horus
			fmt.Printf("Enter your message content: ")

			scanner.Scan()
			content := strings.TrimSpace(scanner.Text())

			fmt.Printf("%#v\n", content)

			sendMessage(name, content)

		case "4":
			// Get the conversation name to send a message to
			fmt.Print("Enter the conversation's name: ")

			scanner.Scan()
			name := scanner.Text()

			fmt.Printf("Now chatting under conversation '%v'. Type 'stop' to stop\n", name)
			sendMultiMessage(name)

		case "5":
			return
		}

		fmt.Println()
	}

	// Do debug testing here
	removeConversation("test-003")
	addConversation("test-003")
	sendMessage("test-003", "Create a new keepass password")
	sendMessage("test-003", "testing")
	sendMessage("test-003", "/path/")
	sendMessage("test-003", "username")
	sendMessage("test-003", "password")
	sendMessage("test-003", "url")
	sendMessage("test-003", "notes")
	sendMessage("test-003", "yes")
}
