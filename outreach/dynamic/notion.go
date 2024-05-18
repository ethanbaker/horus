package dynamic

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	notionapi "github.com/dstotijn/go-notion"
)

// notion.go contains all dynamic notion reminders

/* ---- TYPES ---- */

// NotionDatabase type holds a database ID and query to get the database with
type NotionDatabase struct {
	ID    string
	Query notionapi.DatabaseQuery
}

// Event type used to hold calendar events for nice formatting
type Event struct {
	Start    time.Time
	Name     string
	Timespan string
}

type httpTransport struct {
	w io.Writer
}

// RoundTrip implements http.RoundTripper. It multiplexes the read HTTP response
// data to an io.Writer for debugging.
func (t *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	res.Body = io.NopCloser(io.TeeReader(res.Body, t.w))

	return res, nil
}

/* ---- CONSTANTS ---- */

const SCHEDULE_REMINDERS_ERROR_LIMIT = 10

var SCHEDULE_ITEMS = NotionDatabase{
	ID: os.Getenv("NOTION_DATABASE_SCHEDULE_ID"),
	Query: notionapi.DatabaseQuery{
		Filter: &notionapi.DatabaseQueryFilter{
			// 'Day' is checked
			Property: "Day",
			DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
				Formula: &notionapi.FormulaDatabaseQueryFilter{
					Checkbox: &notionapi.CheckboxDatabaseQueryFilter{
						Equals: boolPointer(true),
					},
				},
			},
		},
		// Sort by ascending date
		Sorts: []notionapi.DatabaseQuerySort{
			{
				Property:  "Date",
				Direction: notionapi.SortDirAsc,
			},
		},
	},
}

/* ---- GLOBALS ---- */

// The notion client functions will be using
var notion *notionapi.Client = notionapi.NewClient(os.Getenv("NOTION_API_TOKEN"), notionapi.WithHTTPClient(&http.Client{
	Timeout:   20 * time.Second,
	Transport: &httpTransport{w: &bytes.Buffer{}},
}))

/* ---- METHODS ---- */

// Get a list of Notion schedule items and add them to the message struct
func NotionScheduleRemindersUpdate(m *DynamicOutreachMessage) error {
	var events []Event
	var mainErr error

	// Run for a given number of times
	for i := 0; i < SCHEDULE_REMINDERS_ERROR_LIMIT; i++ {
		// Query the schedule items
		schedule, err := notion.QueryDatabase(context.Background(), SCHEDULE_ITEMS.ID, &SCHEDULE_ITEMS.Query)
		if err != nil {
			mainErr = err
			continue
		}

		// Loop for each task page
		for _, p := range schedule.Results {
			// Get the page property IDs from Notion
			page, err := notion.FindPageByID(context.Background(), p.ID)
			if err != nil {
				mainErr = err
				break
			}

			// Get the page property values from their IDs
			properties, ok := page.Properties.(notionapi.DatabasePageProperties)
			if !ok {
				mainErr = fmt.Errorf("cannot cast properties")
				break
			}

			// Get the name of the task
			nameField := properties["Name"]
			if len(nameField.Title) == 0 {
				continue
			}
			name := nameField.Title[0].Text.Content

			// Get the start of the task
			startField := properties["Date"]
			start := startField.Date.Start.Time

			// Get the date of the task
			timespanField := properties["Timespan"]
			timespan := *timespanField.Formula.String

			// Add the event to the calendar events
			events = append(events, Event{
				Start:    start,
				Name:     name,
				Timespan: timespan,
			})
		}

		// On no error eeturn
		if mainErr == nil {
			break
		}
		log.Printf("[ERROR]: Error in notion/NotionScheduleReminders, retrying (err: %v)\n", err)
	}

	// If there's an error, return it
	if mainErr != nil {
		return mainErr
	}

	// Add the events to the struct
	m.data = events
	return nil
}

// Send a message to the user if the
func NotionScheduleReminders(m *DynamicOutreachMessage, now time.Time) string {
	// Get the list of events
	events, ok := m.data.([]Event)
	if !ok {
		return ""
	}

	// If an event starts on this second, return it's format string and remove it from the list
	current := []Event{}
	for i := 0; i < len(events); i++ {
		e := events[i]

		// If the event is this minute, add it
		if now.Truncate(time.Minute).Equal(e.Start) {
			current = append(current, e)
		}
	}

	// Do formatting and return output
	if len(current) == 0 {
		return ""
	} else if len(current) == 1 {
		return fmt.Sprintf("<STRONG>Schedule Event:<STRONG> %v (%v)\n", current[0].Name, current[0].Timespan)
	}

	// Handle the multi event case
	output := "<STRONG>Schedule Events:<STRONG>\n"
	for _, e := range current {
		output += fmt.Sprintf("- %v (%v)\n", e.Name, e.Timespan)
	}
	return output
}

/* ---- HELPER FUNCTIONS ---- */

func boolPointer(val bool) *bool {
	return &val
}
