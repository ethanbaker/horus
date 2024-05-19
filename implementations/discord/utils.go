// Helper/utility functions
package main

import "strings"

// Format an output string to match discord messages
func format(output string) string {
	output = strings.ReplaceAll(output, "<STRONG>", "**")
	output = strings.ReplaceAll(output, "<EM>", "*")
	output = strings.ReplaceAll(output, "<EM>", "*")
	output = strings.ReplaceAll(output, "<INS>", "__")
	output = strings.ReplaceAll(output, "<DEL>", "~~")
	output = strings.ReplaceAll(output, "<DEL>", "~~")
	output = strings.ReplaceAll(output, "<BLOCKQUOTE_IN>", "> ")
	output = strings.ReplaceAll(output, "<BLOCKQUOTE>", ">>> ")
	output = strings.ReplaceAll(output, "<CODE_IN>", "`")
	output = strings.ReplaceAll(output, "<CODE>", "```")
	output = strings.ReplaceAll(output, "<SPOILER>", "||")

	return output
}
