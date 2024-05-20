package format_test

import (
	"testing"

	"github.com/ethanbaker/horus/utils/format"
	"github.com/stretchr/testify/assert"
)

func TestFormatDiscord(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("**", format.FormatDiscord("<STRONG>"))
	assert.Equal("*", format.FormatDiscord("<EM>"))
	assert.Equal("__", format.FormatDiscord("<INS>"))
	assert.Equal("~~", format.FormatDiscord("<DEL>"))
	assert.Equal("> ", format.FormatDiscord("<BLOCKQUOTE_IN>"))
	assert.Equal("\n>>> ", format.FormatDiscord("<BLOCKQUOTE>"))
	assert.Equal("`", format.FormatDiscord("<CODE_IN>"))
	assert.Equal("```", format.FormatDiscord("<CODE>"))
	assert.Equal("||", format.FormatDiscord("<SPOILER>"))
}
