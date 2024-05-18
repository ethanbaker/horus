package types

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/objx"
)

/* ---- I/O TYPES ---- */

// Input content to a Horus library
type Input struct {
	Message     string // The user's message in plaintext
	Permissions byte   // The permissions this input send has
	Data        any    // Any external program data from implementations

	Parameters objx.Map // Function parameters given in a function call by the model
}

// Get a string parameter from the model
func (i *Input) GetString(label string, def string) (string, bool) {
	output := i.Parameters.Get(label).Str(def)
	return output, output != def
}

// Get an integer parameter from the model
func (i *Input) GetInteger(label string, def int) (int, bool) {
	output := i.Parameters.Get(label).Int(def)
	return output, output != def
}

// Get an boolean parameter from the model
func (i *Input) GetBool(label string) bool {
	return i.Parameters.Get(label).Bool()
}

// Output response from a Horus library
type Output struct {
	Message string `json:"message"` // The library's message in plaintext
	Data    any    `json:"data"`    // Any external program data returned by the library
	Error   error  `json:"error"`   // Any error present in finding the output
}

/* ---- I/O INTERFACE TYPES ---- */

// Input data coming from discord
type DiscordInput struct {
	Message *discordgo.MessageCreate
}

// Output data going to a local file
type FileOutput struct {
	Filename string
	Content  []byte
}
