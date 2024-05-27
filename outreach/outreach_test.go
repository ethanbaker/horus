package outreach_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethanbaker/horus/outreach"
	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NOTE: any instance of 'sqlmock.AnyArg()' is used as a timestamp, since 'time.Now()' is unreliable

/* ---- CONSTANTS ---- */

const ENV_FILEPATH = "../testing/.env"

/* ---- SUITE ---- */

// Suite struct holds all globals and setup/teardown methods for tests
type Suite struct {
	suite.Suite
	DB   *gorm.DB        // The gorm DB to pass to the config
	mock sqlmock.Sqlmock // The SQL mock to pass to the config

	//bot            *horus.Bot
	config         *config.Config
	outreachConfig *types.OutreachConfig
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

	// Setup the outreach configuration
	err = outreach.Setup(s.config)
	assert.Nil(err)

	// Read in the outreach config
	yamlFile, err := os.ReadFile(s.config.Getenv("BASE_PATH") + s.config.Getenv("OUTREACH_CONFIG"))
	assert.Nil(err)

	err = yaml.Unmarshal(yamlFile, &s.outreachConfig)
	assert.Nil(err)

	// Create a new channel
	ch, err := outreach.AddChannel(types.Discord)
	assert.NotNil(ch)
	assert.Nil(err)

	// Make sure you can't add another channel
	_, err = outreach.AddChannel(types.Discord)
	assert.NotNil(err)
	assert.Equal("channel with key discord already exists", err.Error())

	// NOTE: would do base table assertions here if outreach uses SQL in future
}

// Run checks after each test
func (s *Suite) AfterTest(_, _ string) {
	assert := assert.New(s.T())

	assert.Nil(s.mock.ExpectationsWereMet())
}

/* ---- SUITE TESTING ---- */

// Test exceptions/errors in the New function
func (s *Suite) TestNew() {
	assert := assert.New(s.T())

	// Test invalid module name
	m, err := outreach.New("invalid_module", []types.OutreachMethod{types.Discord}, nil)
	assert.Nil(m)
	assert.NotNil(err)
	assert.Equal("requested message key invalid_module does not exist", err.Error())

	// Test invalid channel setup
	m, err = outreach.New("dynamic", []types.OutreachMethod{}, nil)
	assert.Nil(m)
	assert.NotNil(err)
	assert.Equal("no channels selected", err.Error())
}

func (s *Suite) TestDynamic() {
	assert := assert.New(s.T())

	// Add the dynamic outreaches if they match this implementation key
	for _, msg := range s.outreachConfig.Dynamic {
		// Add the outreach
		m, err := outreach.New("dynamic", []types.OutreachMethod{types.Discord}, types.DynamicOutreach{
			Function:        msg.Name,
			IntervalMinutes: time.Minute * time.Duration(msg.IntervalMinutes),
		})
		assert.Nil(err)

		// Test the GetChannels function
		assert.Equal(1, len(m.GetChannels()))

		// Test the Start/Stop/Delete functions

		// Try to start the message when it's already started
		err = m.Start()
		assert.NotNil(err)
		assert.Equal("message is still active", err.Error())

		// Stop the message
		err = m.Stop()
		assert.Nil(err)

		// Start the message
		err = m.Start()
		assert.Nil(err)

		// Delete the message
		err = m.Delete()
		assert.NotNil(err)
		assert.Equal("message is already stopped", err.Error())
	}
}

func (s *Suite) TestStatic() {
	assert := assert.New(s.T())

	// Add the static outreaches if they match this implementation key
	for _, msg := range s.outreachConfig.Static {
		// Add the outreach
		m, err := outreach.New("static", []types.OutreachMethod{types.Discord}, types.StaticOutreach{
			Function: msg.Name,
			Repeat:   msg.Repeat,
		})
		assert.Nil(err)

		// Test the GetChannels function
		assert.Equal(1, len(m.GetChannels()))

		// Test the Start/Stop/Delete functions

		// Try to start the message when it's already started
		err = m.Start()
		assert.NotNil(err)
		assert.Equal("message is still active", err.Error())

		// Stop the message
		err = m.Stop()
		assert.Nil(err)

		// Try to stop the message when it's already stopped
		err = m.Stop()
		assert.NotNil(err)
		assert.Equal("message is already stopped", err.Error())

		// Start the message
		err = m.Start()
		assert.Nil(err)

		// Delete the message
		err = m.Delete()
		assert.Nil(err)

		// Try to start the message
		err = m.Start()
		assert.NotNil(err)
		assert.Equal("cron is nil", err.Error())

		// Try to stop the message
		err = m.Stop()
		assert.NotNil(err)
		assert.Equal("message is already stopped", err.Error())
	}

}

/* ---- INIT ---- */

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}
