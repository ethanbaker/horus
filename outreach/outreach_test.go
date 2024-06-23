package outreach_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

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
	cfg, errs := config.New()
	assert.Equal(0, len(errs))

	cfg.Gorm = s.DB
	s.config = cfg

	// Setup the outreach configuration
	err = outreach.Setup(s.config)
	assert.Nil(err)

	// Read in the outreach config
	yamlFile, err := os.ReadFile(filepath.Join(s.config.Getenv("BASE_PATH"), s.config.Getenv("OUTREACH_CONFIG")))
	assert.Nil(err)

	err = yaml.Unmarshal(yamlFile, &s.outreachConfig)
	assert.Nil(err)

	// Create a new channel
	ch, err := outreach.AddChannel(types.DiscordMethod)
	assert.NotNil(ch)
	assert.Nil(err)

	// Make sure you can't add another channel
	_, err = outreach.AddChannel(types.DiscordMethod)
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
	err := outreach.New("invalid_module", "name", []types.OutreachMethod{types.DiscordMethod}, map[string]any{})
	assert.NotNil(err)
	assert.Equal("requested module 'invalid_module' does not exist", err.Error())

	// Test invalid channel setup
	err = outreach.New(types.DynamicModule, "name", []types.OutreachMethod{}, map[string]any{})
	assert.NotNil(err)
	assert.Equal("no channels selected", err.Error())

	// Test invalid name
	err = outreach.New(types.DynamicModule, "", []types.OutreachMethod{types.DiscordMethod}, map[string]any{})
	assert.NotNil(err)
	assert.Equal("invalid name ''", err.Error())
}

func (s *Suite) TestDynamic() {
	assert := assert.New(s.T())

	// Add the dynamic outreaches if they match this implementation key
	for _, msg := range s.outreachConfig.Dynamic {
		// Add the outreach
		err := outreach.New(types.DynamicModule, msg.Name, msg.Channels, msg.Data)
		assert.Nil(err)

		// Stop the outreach
		err = outreach.Stop(msg.Name)
		assert.Nil(err)
		err = outreach.Stop(msg.Name)
		assert.NotNil(err)
		assert.Equal("message is already stopped", err.Error())

		// Start the outreach
		err = outreach.Start(msg.Name)
		assert.Nil(err)
		err = outreach.Start(msg.Name)
		assert.NotNil(err)
		assert.Equal("message is still active", err.Error())

		// Delete the outreach
		err = outreach.Delete(msg.Name)
		assert.Nil(err)
		err = outreach.Delete(msg.Name)
		assert.NotNil(err)
		assert.Equal("message has been deleted", err.Error())

		// We cannot stop or start the message after it has been deleted
		err = outreach.Start(msg.Name)
		assert.NotNil(err)
		assert.Equal("message has been deleted", err.Error())

		err = outreach.Stop(msg.Name)
		assert.NotNil(err)
		assert.Equal("message has been deleted", err.Error())
	}
}

func (s *Suite) TestStatic() {
	assert := assert.New(s.T())

	// Add the static outreaches if they match this implementation key
	for _, msg := range s.outreachConfig.Static {
		// Add the outreach
		err := outreach.New(types.StaticModule, msg.Name, msg.Channels, msg.Data)
		assert.Nil(err)

		// Stop the outreach
		err = outreach.Stop(msg.Name)
		assert.Nil(err)
		err = outreach.Stop(msg.Name)
		assert.NotNil(err)
		assert.Equal("message is already stopped", err.Error())

		// Start the outreach
		err = outreach.Start(msg.Name)
		assert.Nil(err)
		err = outreach.Start(msg.Name)
		assert.NotNil(err)
		assert.Equal("message is still active", err.Error())

		// Delete the outreach
		err = outreach.Delete(msg.Name)
		assert.Nil(err)
		err = outreach.Delete(msg.Name)
		assert.NotNil(err)
		assert.Equal("message has been deleted", err.Error())

		// We cannot stop or start the message after it has been deleted
		err = outreach.Start(msg.Name)
		assert.NotNil(err)
		assert.Equal("message has been deleted", err.Error())

		err = outreach.Stop(msg.Name)
		assert.NotNil(err)
		assert.Equal("message has been deleted", err.Error())
	}

}

/* ---- INIT ---- */

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}
