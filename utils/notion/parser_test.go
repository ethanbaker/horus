package notion_test

import (
	"testing"

	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/notion"
	"github.com/stretchr/testify/assert"
)

func TestParsePage(t *testing.T) {
	assert := assert.New(t)

	// Read in the testing config to get a notion page to parse
	config, errs := config.New()
	assert.Equal(0, len(errs))
	assert.NotNil(config.Notion)

	// Get the notion page ID
	id := config.Getenv("NOTION_TEST_PAGE")
	assert.NotEmpty(id)

	// Parse a page
	output, err := notion.ParsePage(config.Notion, id)
	assert.Nil(err)
	assert.NotNil(output)
}
