package horus

import openai "github.com/sashabaranov/go-openai"

/* ---- PERMISSION CONSTANTS ---- */

const (
	PERMISSIONS_NONE       = 0b00000000 // Null value - nothing is enabled
	PERMISSIONS_GPT        = 0b00000001 // GPT functionality is enabled
	PERMISSIONS_PRVMODULES = 0b00000010 // Only "private" custom modules are enabled
	PERMISSIONS_PUBMODULES = 0b00000100 // Only "public" custom modules are enabled

	PERMISSIONS_ALLMODULES = 0b00000110 // All custom modules are enabled
	PERMISSIONS_ALL        = 0b11111111 // All custom modules and GPT functionality is enabled
)

/* ---- OPENAI CONSTANTS ---- */

const (
	OPENAI_MODEL     = openai.GPT3Dot5Turbo // What OpenAI model is being used
	OPENAI_ROLE      = openai.ChatMessageRoleSystem
	OPENAI_MAXTOKENS = 500                                                 // What is the maximum amount of tokens that can be sent to the model? (Tokens = Words / 0.75)
	OPENAI_SYSPROMPT = `You are a helpful personal assistant named Horus.` // System Prompt for the OpenAI model
)
