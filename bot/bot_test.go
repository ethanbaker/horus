package horus_test

import (
	"database/sql"
	"log"
	"testing"

	horus "github.com/ethanbaker/horus/bot"
	"github.com/ethanbaker/horus/utils/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

/* ---- GLOBALS ---- */

var bot *horus.Bot

var db *sql.DB

/* ---- TESTS ---- */

func TestAddConversation(t *testing.T) {
	assert := assert.New(t)

	conversationName := "test-001"

	// Add a conversation to the bot
	err := bot.AddConversation(conversationName)
	assert.Nil(err)

	// Verify the conversation is in the SQL
	rows, err := db.Query("SELECT id, bot_id, name FROM conversations WHERE name = ?", conversationName)
	assert.Nil(err)

	// Start parsing the rows
	defer rows.Close()
	assert.True(rows.Next())
	assert.Nil(rows.Err())

	// Read in row values
	id, botID, name := -1, -1, ""
	assert.Nil(rows.Scan(&id, &botID, &name))

	// Check for expected values
	assert.Equal(int(bot.Model.ID), botID)
	assert.Equal(conversationName, name)

	// There should only be one row
	assert.False(rows.Next())
}

/*
func TestDeleteConversation(t *testing.T) {
	assert := assert.New(t)

}

func TestIsConversation(t *testing.T) {
	assert := assert.New(t)

}

func TestSendMessage(t *testing.T) {
	assert := assert.New(t)

}

func TestAddMessage(t *testing.T) {
	assert := assert.New(t)

}

func TestEditVariable(t *testing.T) {
	assert := assert.New(t)

}

func TestGetVariable(t *testing.T) {
	assert := assert.New(t)

}
*/

/* ---- MAIN ---- */

func TestMain(m *testing.M) {
	var err error

	// Read the config
	config, errs := config.NewConfigFromFile("./testing/.env")
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("[ERROR]: error from config (%v)\n", err)
		}
		log.Fatalf("[ERROR]: error reading config. Failing\n")
	}

	// Get the MySQL Database
	db, err = config.Gorm.DB()
	if err != nil {
		log.Fatalf("[ERROR]: error fetching MySQL Database (%v)", err)
	}

	// Create a new bot
	bot, err = horus.NewBot("horus-test", horus.PERMISSIONS_ALL)
	if err != nil {
		log.Fatalf("[ERROR]: error initalizing bot (%v)\n", err)
	}

	// Setup the bot
	if err = bot.Setup(config); err != nil {
		log.Fatalf("[ERROR]: error setting up bot (%v)\n", err)
	}

	// Begin a transaction
	config.Gorm = config.Gorm.Begin()
	config.Gorm.SavePoint("pretest")
	defer func() {
		if r := recover(); r != nil {
			config.Gorm.RollbackTo("pretest")
		}
	}()

	// Run the tests
	m.Run()

	// Rollback the transaction
	config.Gorm.RollbackTo("pretest")
}
