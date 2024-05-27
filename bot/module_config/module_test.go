package module_config_test

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	horus "github.com/ethanbaker/horus/bot"
	"github.com/ethanbaker/horus/bot/module_config"
	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NOTE: any instance of 'sqlmock.AnyArg()' is used as a timestamp, since 'time.Now()' is unreliable, or as a tool call ID

/* ---- CONSTANTS ---- */

const ENV_FILEPATH = "../../testing/.env"

/* ---- SUITE ---- */

// Suite struct holds all globals and setup/teardown methods for tests
type Suite struct {
	suite.Suite
	DB   *gorm.DB        // The gorm DB to pass to the config
	mock sqlmock.Sqlmock // The SQL mock to pass to the config

	bot    *horus.Bot
	config *config.Config
}

// Setup the suite struct for testing
func (s *Suite) SetupSuite() {
	assert := assert.New(s.T())

	var (
		db  *sql.DB
		err error
	)

	// Create a new sqlmock instance
	db, s.mock, err = sqlmock.New()
	assert.Nil(err)

	// Create a new gorm database
	s.DB, err = gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.Nil(err)

	// Log all information about database entries
	s.DB.Logger.LogMode(logger.Info)

	// Create a new config with the fake mock service. The only error that should exist is an SQL one
	cfg, errs := config.NewConfigFromFile(ENV_FILEPATH)
	assert.Equal(1, len(errs))
	assert.Equal("cannot initialize mysql config; are all 'SQL' fields set?", errs[0].Error())

	cfg.Gorm = s.DB
	s.config = cfg

	// Expect table creation statements for SQL setup, then perform the SQL setup
	var tables = []string{"tool_calls", "messages", "memories", "conversations", "bots"}
	for _, table := range tables {
		s.mock.ExpectQuery("SELECT SCHEMA_NAME from Information_schema.SCHEMATA").
			WithArgs("%", "").
			WillReturnRows(sqlmock.NewRows([]string{"SCHEMA_NAME"}))

		s.mock.ExpectExec(fmt.Sprintf("CREATE TABLE `%v`", table)).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	horus.InitSQL(s.config)
}

// Create a fresh instance of a bot before each test
func (s *Suite) BeforeTest(_, _ string) {
	assert := assert.New(s.T())
	var err error

	var (
		name  string = "horus_testing"
		perms byte   = horus.PERMISSIONS_ALL
	)

	// Expect the bot to be inserted
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `bots` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, name, perms).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Expect the memory bank to be inserted
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `memories` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "", "", "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Create the bot
	s.bot, err = horus.NewBot(name, perms, s.config)
	assert.Nil(err)

	// Add the module
	module_config.NewModule(s.bot, true)
}

// Run checks after each test
func (s *Suite) AfterTest(_, _ string) {
	assert := assert.New(s.T())

	assert.Nil(s.mock.ExpectationsWereMet())
}

/* ---- SUITE TESTING ---- */

func (s *Suite) TestSetTimezone() {
	assert := assert.New(s.T())

	// Test constants
	var (
		name  = "test-001"
		input = types.Input{
			Message:     "Set the preferred timezone to America/New_York",
			Temperature: math.SmallestNonzeroFloat32,
		}
	)

	// TEST SQL OUTLINE
	// - Insert a conversation
	// - Insert the default first message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Save the bot
	//   - Save the bot's memory
	//   - Conversations are inserted or updated on duplicate key
	//     - Messages are inserted or udpated on duplicate key
	// - Insert a sent message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	//     - Tool calls in message are inserted or updated on duplicate key
	// - Insert the received message
	//   - Insert the received message's tool call
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Add function call
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Add function call response message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Save the bot's memory

	// Insert a conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `conversations` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Insert the default first message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Save the bot
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `bots`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, s.bot.Name, s.bot.Permissions, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Save the bot's memory
	s.mock.ExpectExec("^INSERT INTO `memories` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "", "", "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))

	// Conversations are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Insert a sent message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 1, "user", "", input.Message, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Insert the received message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 2, "assistant", "", "", "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Insert the received message's tool call
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `tool_calls` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "function", "set_timezone", sqlmock.AnyArg(), 1). // AnyArg() in the second to last argument is for the model returning various parameters
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Add function call
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 3, "tool", "set_timezone", sqlmock.AnyArg(), sqlmock.AnyArg()). // AnyArg()'s at the end are for content returned by function and call ID
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Add function call response message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 4, "assistant", "", sqlmock.AnyArg(), ""). // AnyArg() at second last argument is for bot's response to function call, which will change
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Save the bot's memory
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `memories`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "America/New_York", "", "", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Add a conversation to the bot
	err := s.bot.AddConversation(name)
	assert.Nil(err)

	// Send a message
	output, err := s.bot.SendMessage(name, &input)
	assert.Nil(err)

	assert.NotNil(output)
	log.Printf("[OUTPUT - set_timezone]: %v\n", output)
}

func (s *Suite) TestSetCity() {
	assert := assert.New(s.T())

	// Test constants
	var (
		name  = "test-001"
		input = types.Input{
			Message:     "Set the preferred city to Raleigh",
			Temperature: math.SmallestNonzeroFloat32,
		}
	)

	// TEST SQL OUTLINE
	// - Insert a conversation
	// - Insert the default first message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Save the bot
	//   - Save the bot's memory
	//   - Conversations are inserted or updated on duplicate key
	//     - Messages are inserted or udpated on duplicate key
	// - Insert a sent message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	//     - Tool calls in message are inserted or updated on duplicate key
	// - Insert the received message
	//   - Insert the received message's tool call
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Add function call
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Add function call response message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Save the bot's memory

	// Insert a conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `conversations` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Insert the default first message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Save the bot
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `bots`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, s.bot.Name, s.bot.Permissions, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Save the bot's memory
	s.mock.ExpectExec("^INSERT INTO `memories` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "", "", "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))

	// Conversations are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Insert a sent message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 1, "user", "", input.Message, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Insert the received message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 2, "assistant", "", "", "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Insert the received message's tool call
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `tool_calls` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "function", "set_city", sqlmock.AnyArg(), 1). // AnyArg() in the second to last argument is for the model returning various parameters
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Add function call
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 3, "tool", "set_city", sqlmock.AnyArg(), sqlmock.AnyArg()). // AnyArg()'s at the end are for content returned by function and call ID
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Add function call response message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 4, "assistant", "", sqlmock.AnyArg(), ""). // AnyArg() at second last argument is for bot's response to function call, which will change
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Save the bot's memory
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `memories`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "", "Raleigh", "", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Add a conversation to the bot
	err := s.bot.AddConversation(name)
	assert.Nil(err)

	// Send a message
	output, err := s.bot.SendMessage(name, &input)
	assert.Nil(err)

	assert.NotNil(output)
	log.Printf("[OUTPUT - set_city]: %v\n", output)
}

func (s *Suite) TestSetTemperatureUnit() {
	assert := assert.New(s.T())

	// Test constants
	var (
		name  = "test-001"
		input = types.Input{
			Message:     "Set the preferred temperature unit to fahrenheit",
			Temperature: math.SmallestNonzeroFloat32,
		}
	)

	// TEST SQL OUTLINE
	// - Insert a conversation
	// - Insert the default first message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Save the bot
	//   - Save the bot's memory
	//   - Conversations are inserted or updated on duplicate key
	//     - Messages are inserted or udpated on duplicate key
	// - Insert a sent message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	//     - Tool calls in message are inserted or updated on duplicate key
	// - Insert the received message
	//   - Insert the received message's tool call
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Add function call
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Add function call response message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Save the bot's memory

	// Insert a conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `conversations` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Insert the default first message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Save the bot
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `bots`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, s.bot.Name, s.bot.Permissions, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Save the bot's memory
	s.mock.ExpectExec("^INSERT INTO `memories` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "", "", "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))

	// Conversations are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Insert a sent message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 1, "user", "", input.Message, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Insert the received message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 2, "assistant", "", "", "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Insert the received message's tool call
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `tool_calls` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "function", "set_temperature_unit", sqlmock.AnyArg(), 1). // AnyArg() in the second to last argument is for the model returning various parameters
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Add function call
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 3, "tool", "set_temperature_unit", sqlmock.AnyArg(), sqlmock.AnyArg()). // AnyArg()'s at the end are for content returned by function and call ID
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Add function call response message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 4, "assistant", "", sqlmock.AnyArg(), ""). // AnyArg() at second last argument is for bot's response to function call, which will change
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, name, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Save the bot's memory
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `memories`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "", "", "fahrenheit", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Add a conversation to the bot
	err := s.bot.AddConversation(name)
	assert.Nil(err)

	// Send a message
	output, err := s.bot.SendMessage(name, &input)
	assert.Nil(err)

	assert.NotNil(output)
	log.Printf("[OUTPUT - set_temperature_unit]: %v\n", output)
}

/* ---- INIT ---- */

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}
