// Package format is used to format Horus' output to different implementations
package format

import "strings"

// FormatDiscord turns a library response from Horus into one for Discord
func FormatDiscord(res string) string {
	output := res

	output = strings.ReplaceAll(output, "<STRONG>", "**")
	output = strings.ReplaceAll(output, "<EM>", "*")
	output = strings.ReplaceAll(output, "<INS>", "__")
	output = strings.ReplaceAll(output, "<DEL>", "~~")
	output = strings.ReplaceAll(output, "<BLOCKQUOTE_IN>", "> ")
	output = strings.ReplaceAll(output, "<BLOCKQUOTE>", "\n>>> ")
	output = strings.ReplaceAll(output, "<CODE_IN>", "`")
	output = strings.ReplaceAll(output, "<CODE>", "```")
	output = strings.ReplaceAll(output, "<SPOILER>", "||")

	return output
}
