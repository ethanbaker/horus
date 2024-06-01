// parser.go is used to parse notion pages and convert them to HTML
package notion

import (
	"context"
	"fmt"
	"log"
	"reflect"

	notionapi "github.com/dstotijn/go-notion"
)

// Parse a page from notion and return the corresponding HTML
func ParsePage(client *notionapi.Client, id string) (string, error) {
	// Get the page from notion
	page, err := client.FindBlockChildrenByID(context.Background(), id, &notionapi.PaginationQuery{
		StartCursor: "",
		PageSize:    100,
	})
	if err != nil {
		return "", err
	}

	content := ""
	for _, raw := range page.Results {
		switch block := raw.(type) {
		// Heading 1
		case *notionapi.Heading1Block:
			if len(block.RichText) > 0 {
				content += fmt.Sprintf("<H1>%v</H1>\n", block.RichText[0].Text.Content)
			}

		// Heading 2
		case *notionapi.Heading2Block:
			if len(block.RichText) > 0 {
				content += fmt.Sprintf("<H2>%v</H2>\n", block.RichText[0].Text.Content)
			}

		// Heading 3
		case *notionapi.Heading3Block:
			if len(block.RichText) > 0 {
				content += fmt.Sprintf("<H3>%v</H3>\n", block.RichText[0].Text.Content)
			}

		// Paragraph
		case *notionapi.ParagraphBlock:
			if len(block.RichText) > 0 {
				content += fmt.Sprintf("<P>%v</P>\n", block.RichText[0].Text.Content)
			}

		// For unknown blocks, print out a log message
		default:
			log.Printf("[WARN]: unhandled block type %v\n", reflect.TypeOf(block))
		}
	}

	return content, nil
}
