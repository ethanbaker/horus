package horus_test

import (
	"database/sql"
	"fmt"
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	horus "github.com/ethanbaker/horus/bot"
	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NOTE: any instance of 'sqlmock.AnyArg()' is used as a timestamp, since 'time.Now()' is unreliable

/* ---- CONSTANTS ---- */

const ENV_FILEPATH = "./testing/.env"

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

	// Read in an environment file to get variables
	err = godotenv.Load(ENV_FILEPATH)
	assert.Nil(err)

	vars, err := godotenv.Read(ENV_FILEPATH)
	assert.Nil(err)

	// Create a new config with the fake mock service. The only error that should exist is an SQL one
	cfg, errs := config.NewConfigFromVariables(&vars)
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

	s.bot, err = horus.NewBot(name, perms, s.config)
	assert.Nil(err)

}

// Run checks after each test
func (s *Suite) AfterTest(_, _ string) {
	assert := assert.New(s.T())

	assert.Nil(s.mock.ExpectationsWereMet())
}

/* ---- SUITE TESTING ---- */

func (s *Suite) TestAddConversation() {
	assert := assert.New(s.T())

	// Test constants
	var (
		name = "test-001"
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

	// Add a conversation to the bot
	err := s.bot.AddConversation(name)
	assert.Nil(err)
}

func (s *Suite) TestDeleteConversation() {
	assert := assert.New(s.T())

	// Test constants
	var (
		name = "test-001"
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
	// - Delete the messages in the conversation by setting `deleted_at` to the current time
	// - Delete the conversation by setting `deleted_at` to the current time

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

	// Delete the messages in the conversation by setting `deleted_at` to the current time
	s.mock.ExpectBegin()
	s.mock.ExpectExec("UPDATE `messages`").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Delete the conversation by setting `deleted_at` to the current time
	s.mock.ExpectBegin()
	s.mock.ExpectExec("UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Add a conversation to the bot
	err := s.bot.AddConversation(name)
	assert.Nil(err)

	// Delete the conversation from the bot
	err = s.bot.DeleteConversation(name)
	assert.Nil(err)
}

func (s *Suite) TestIsConversation() {
	assert := assert.New(s.T())

	// Test constants
	var (
		name = "test-001"
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

	// Add a conversation to the bot
	err := s.bot.AddConversation(name)
	assert.Nil(err)

	// Make sure the conversation is a real conversation
	assert.True(s.bot.IsConversation(name))

	// Test invalid keys
	assert.False(s.bot.IsConversation(""))
	assert.False(s.bot.IsConversation("invalid"))
}

// Test EditVariable() and GetVariable()
func (s *Suite) TestVariables() {
	assert := assert.New(s.T())

	// Add variables to the bot
	s.bot.EditVariable("test_variable_1", 0)
	s.bot.EditVariable("test_variable_2", "abc")
	s.bot.EditVariable("test_variable_3", float64(0.123))

	// Get variables from the bot and assume equality
	var1, ok := s.bot.GetVariable("test_variable_1").(int)
	assert.True(ok)
	assert.Equal(0, var1)

	var2, ok := s.bot.GetVariable("test_variable_2").(string)
	assert.True(ok)
	assert.Equal("abc", var2)

	var3, ok := s.bot.GetVariable("test_variable_3").(float64)
	assert.True(ok)
	assert.Equal(0.123, var3)
}

func (s *Suite) TestSendMessage() {
	assert := assert.New(s.T())

	// Test constants
	var (
		name  = "test-001"
		input = types.Input{
			Message:     "Repeat exactly what I say: hello",
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
	// - Insert the received message
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
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 2, "assistant", "", "Hello", "").
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
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "", "", "", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Add a conversation to the bot
	err := s.bot.AddConversation(name)
	assert.Nil(err)

	// Send a message
	output, err := s.bot.SendMessage(name, &input)
	assert.Nil(err)

	assert.NotNil(output)
	assert.Equal("Hello", output.Message)
}

func (s *Suite) TestAddMessage() {
	assert := assert.New(s.T())

	// Test constants
	var (
		key     = "test-001"
		role    = "assistant"
		name    = ""
		content = "Hello! How may I help you today?"
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
	// - Add a message
	// - Save the conversation
	//   - Messages in conversation are inserted or updated on duplicate key
	// - Save the bot's memory

	// Insert a conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `conversations` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, key).
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
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, key, 1).
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
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, key, 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Add a message
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 1, role, name, content, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Save the conversation
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `conversations`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, key, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Messages in conversation are inserted or updated on duplicate key
	s.mock.ExpectExec("^INSERT INTO `messages` (.+)").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, 0, "system", "system", horus.OPENAI_SYSPROMPT, "", 1).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrNoRows))
	s.mock.ExpectCommit()

	// Save the bot's memory
	s.mock.ExpectBegin()
	s.mock.ExpectExec("^UPDATE `memories`").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, 1, "", "", "", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// Add a conversation to the bot
	err := s.bot.AddConversation(key)
	assert.Nil(err)

	// Add a message
	err = s.bot.AddMessage(key, role, name, content)
	assert.Nil(err)
}

/* ---- INIT ---- */

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}
