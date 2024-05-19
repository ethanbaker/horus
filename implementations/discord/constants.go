// Constants and types used throughout the implementation
package main

import (
	"os"
	"strings"
	"time"

	mysql_driver "github.com/go-sql-driver/mysql"
)

/* ---- TYPES ---- */

// ChannelInfo represents a discord channel. It holds the ID of the channel and the last message
// sent in the channel for determining if Horus should keep certain messages in its context
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
