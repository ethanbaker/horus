// Package config is used as a config service throughout the rest of Horus
package config

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	notionapi "github.com/dstotijn/go-notion"
	mysql_driver "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config type is used to hold configuration and set on initialization
type Config struct {
	variables *map[string]string // Variables stored in the environment

	// Config globals
	Openai                *openai.Client      // The OpenAI client
	DSN                   mysql_driver.Config // An SQL config to attach to the MySQL database
	Gorm                  *gorm.DB            // The Gorm Database to connect to
	Notion                *notionapi.Client   // The Notion client
	DiscordOpenChannels   []string            // Open channels in discord
	DiscordThreadChannels []string            // Thread channels in discord
}

// Set a single environment variable
func (c *Config) Setenv(key string, value string) error {
	if key == "" {
		return fmt.Errorf("environment key cannot be empty")
	}

	(*c.variables)[key] = value
	return nil
}

// Get a single environment variable
func (c *Config) Getenv(key string) string {
	str, ok := (*c.variables)[key]
	if !ok {
		return ""
	}

	return str
}

// Load the config from a given '.env' file
func (c *Config) loadFromFile(filename string) []error {
	// Load the .env file
	if err := godotenv.Load(filename); err != nil {
		return []error{err}
	}

	// Read the environment into the config
	vars, err := godotenv.Read(filename)
	if err != nil {
		return []error{err}
	}

	c.variables = &vars
	return c.setup()
}

// Set the config variables to a map
func (c *Config) loadFromVariables(vars *map[string]string) []error {
	c.variables = vars

	return c.setup()
}

// Setup globals in the config. We return a list of errors in case a testing function doesn't
// require different parts of the setup process. In production we check for any errors, while in
// testing we check for unexpected ones
func (c *Config) setup() []error {
	var errs []error

	// Setup the OpenAI Client
	openaiToken := c.Getenv("OPENAI_TOKEN")
	if openaiToken == "" {
		errs = append(errs, fmt.Errorf("cannot initialize openai client; is 'OPENAI_TOKEN' set?"))
	}
	c.Openai = openai.NewClient(c.Getenv("OPENAI_TOKEN"))

	// Setup the MySQL config
	user := c.Getenv("SQL_USER")
	passwd := c.Getenv("SQL_PASSWD")
	net := c.Getenv("SQL_NET")
	addr := c.Getenv("SQL_ADDR")
	dbname := c.Getenv("SQL_DBNAME")

	if user == "" || passwd == "" || net == "" || addr == "" || dbname == "" {
		errs = append(errs, fmt.Errorf("cannot initialize mysql config; are all 'SQL' fields set?"))
	}

	c.DSN = mysql_driver.Config{
		User:      user,
		Passwd:    passwd,
		Net:       net,
		Addr:      addr,
		DBName:    dbname,
		ParseTime: true,
		Loc:       time.Local,
	}

	// Open a database with gorm
	gorm, err := gorm.Open(mysql.Open(c.DSN.FormatDSN()), &gorm.Config{})
	if err != nil {
		errs = append(errs, fmt.Errorf("cannot initalize gorm client (%v)", err))
	}
	c.Gorm = gorm

	// Setup open discord channels
	openChannels := c.Getenv("DISCORD_BOT_OPEN_CHANNELS")
	if openChannels == "" {
		errs = append(errs, fmt.Errorf("cannot set up allowed discord channels. Is 'DISCORD_BOT_OPEN_CHANNELS' set?"))
	}
	c.DiscordOpenChannels = strings.Split(openChannels, ",")

	// Setup thread discord channels
	threadChannels := c.Getenv("DISCORD_BOT_THREAD_CHANNELS")
	if threadChannels == "" {
		errs = append(errs, fmt.Errorf("cannot set up allowed thread discord channels. Is 'DISCORD_BOT_THREAD_CHANNELS' set?"))
	}
	c.DiscordThreadChannels = strings.Split(threadChannels, ",")

	// Setup the notion client
	notionToken := c.Getenv("NOTION_API_TOKEN")
	if notionToken == "" {
		errs = append(errs, fmt.Errorf("cannot initialize notion client. Is 'NOTION_API_TOKEN' set?"))
	}
	c.Notion = notionapi.NewClient(notionToken, notionapi.WithHTTPClient(&http.Client{
		Timeout:   20 * time.Second,
		Transport: &httpTransport{w: &bytes.Buffer{}},
	}))

	return errs
}

// Create a new config from an environment file
func NewConfigFromFile(filename string) (*Config, []error) {
	c := Config{}
	return &c, c.loadFromFile(filename)
}

// Create a new config with no file
func NewConfigFromVariables(vars *map[string]string) (*Config, []error) {
	c := Config{}
	return &c, c.loadFromVariables(vars)
}
