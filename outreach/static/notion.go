package static

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	ics "github.com/arran4/golang-ical"
	notionapi "github.com/dstotijn/go-notion"
	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/teambition/rrule-go"
	"gopkg.in/yaml.v3"
)

// notion.go contains all static messages relating to Notion digests

/* ---- TYPES ---- */

// NotionDatabase type holds a database ID and query to get the database with
type NotionDatabase struct {
	ID    string
	Query notionapi.DatabaseQuery
}

// Event type used to hold calendar events for nice formatting
type Event struct {
	Start  time.Time
	Format string
	AllDay bool
}

/* ---- CONSTANTS ---- */

const DIGEST_ERROR_LIMIT = 10

var NORMAL_TASKS = NotionDatabase{
	ID: "",
	Query: notionapi.DatabaseQuery{
		Filter: &notionapi.DatabaseQueryFilter{
			And: []notionapi.DatabaseQueryFilter{
				// 'Complete' is unchecked
				{
					Property: "Complete",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Checkbox: &notionapi.CheckboxDatabaseQueryFilter{
							Equals: boolPointer(false),
						},
					},
				},
				// And 'Canceled' is unchecked
				{
					Property: "Canceled",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Formula: &notionapi.FormulaDatabaseQueryFilter{
							Checkbox: &notionapi.CheckboxDatabaseQueryFilter{
								Equals: boolPointer(false),
							},
						},
					},
				},
				// And 'Priority' is not critical
				{
					Property: "Priority",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Select: &notionapi.SelectDatabaseQueryFilter{
							DoesNotEqual: "Critical",
						},
					},
				},
				{
					Or: []notionapi.DatabaseQueryFilter{
						// And 'Rank' is greater than five and date is on or before 1 week from now
						{
							Property: "Rank",
							DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
								Formula: &notionapi.FormulaDatabaseQueryFilter{
									Number: &notionapi.NumberDatabaseQueryFilter{
										GreaterThanOrEqualTo: intPointer(5),
									},
								},
							},
						},
						// Or 'Date' is on or before one week from now
						{
							Property: "Date",
							DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
								Date: &notionapi.DatePropertyFilter{
									NextWeek: &struct{}{},
								},
							},
						},
						{
							Property: "Date",
							DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
								Date: &notionapi.DatePropertyFilter{
									OnOrBefore: timePointer(time.Now()),
								},
							},
						},
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

var CRITICAL_TASKS = NotionDatabase{
	ID: "",
	Query: notionapi.DatabaseQuery{
		Filter: &notionapi.DatabaseQueryFilter{
			And: []notionapi.DatabaseQueryFilter{
				// 'Priority' equals 'Critical'
				{
					Property: "Priority",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Select: &notionapi.SelectDatabaseQueryFilter{
							Equals: "Critical",
						},
					},
				},
				// And 'Canceled' is unchecked
				{
					Property: "Canceled",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Formula: &notionapi.FormulaDatabaseQueryFilter{
							Checkbox: &notionapi.CheckboxDatabaseQueryFilter{
								Equals: boolPointer(false),
							},
						},
					},
				},
				// And 'Complete' is unchecked
				{
					Property: "Complete",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Checkbox: &notionapi.CheckboxDatabaseQueryFilter{
							Equals: boolPointer(false),
						},
					},
				},
			},
		},
		// Sort by descending effort
		Sorts: []notionapi.DatabaseQuerySort{
			{
				Property:  "Effort",
				Direction: notionapi.SortDirDesc,
			},
		},
	},
}

var SCHEDULE_ITEMS = NotionDatabase{
	ID: "",
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

var RECURRING_TASKS = NotionDatabase{
	ID: "",
	Query: notionapi.DatabaseQuery{
		Filter: &notionapi.DatabaseQueryFilter{
			And: []notionapi.DatabaseQueryFilter{
				// 'Active' is checked
				{
					Property: "Active",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Checkbox: &notionapi.CheckboxDatabaseQueryFilter{
							Equals: boolPointer(true),
						},
					},
				},
				// And 'Upcoming' is checked
				{
					Property: "Upcoming",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Formula: &notionapi.FormulaDatabaseQueryFilter{
							Checkbox: &notionapi.CheckboxDatabaseQueryFilter{
								Equals: boolPointer(true),
							},
						},
					},
				},
				// And 'Done' is unchecked
				{
					Property: "Done",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Checkbox: &notionapi.CheckboxDatabaseQueryFilter{
							Equals: boolPointer(false),
						},
					},
				},
				// And 'Type' is ___
				{
					Property: "Type",
					DatabaseQueryPropertyFilter: notionapi.DatabaseQueryPropertyFilter{
						Select: &notionapi.SelectDatabaseQueryFilter{
							Equals: "",
						},
					},
				},
			},
		},
		// Sort by descending type and ascending name
		Sorts: []notionapi.DatabaseQuerySort{
			{
				Property:  "Name",
				Direction: notionapi.SortDirAsc,
			},
		},
	},
}

var MORNING_AFFIRMATIONS_PAGE = ""

var NIGHT_AFFIRMATIONS_PAGE = ""

/* ---- GLOBALS ---- */

// Calendars for parsing iCal formats
var calendars []*ics.Calendar

// Preferred timezone for formatting
var formatLoc *time.Location

/* ---- INIT ---- */

func NotionInit(c *config.Config) error {
	// Initialize constants
	NORMAL_TASKS.ID = c.Getenv("NOTION_DATABASE_TASKS_ID")
	CRITICAL_TASKS.ID = c.Getenv("NOTION_DATABASE_TASKS_ID")
	SCHEDULE_ITEMS.ID = c.Getenv("NOTION_DATABASE_SCHEDULE_ID")
	RECURRING_TASKS.ID = c.Getenv("NOTION_DATABASE_RECURRING_ID")
	MORNING_AFFIRMATIONS_PAGE = c.Getenv("NOTION_PAGE_MORNING_AFFIRMATIONS")
	NIGHT_AFFIRMATIONS_PAGE = c.Getenv("NOTION_PAGE_NIGHT_AFFIRMATIONS")

	// Read in calendar config
	yamlFile, err := os.ReadFile(c.Getenv("BASE_PATH") + c.Getenv("CALENDAR_CONFIG"))
	if err != nil {
		return err
	}

	var config types.CalendarConfig
	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		return err
	}

	// Repeat for each calendar
	for _, cal := range config.Calendars {
		// Get the URL
		resp, err := http.Get(cal.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Create the calendar
		c, err := ics.ParseCalendar(resp.Body)
		if err != nil {
			return err
		}

		calendars = append(calendars, c)
	}

	// Load the format location
	formatLoc, err = time.LoadLocation(config.TimezoneFormat)
	if err != nil {
		return err
	}

	return nil
}

/* ---- METHODS ---- */

// NotionDailyDigest formats a long, digest string for the user to read with tons of information from Notion
func NotionDailyDigest(c *config.Config) string {
	var output string
	var err error

	// Run for a given number of times
	for i := 0; i < DIGEST_ERROR_LIMIT; i++ {
		output, err = getNotionDailyDigest(c)

		// On no error, return the output
		if err == nil {
			return output
		}
		log.Printf("[ERROR]: Error in notion/NotionDailyDigest, retrying (err: %v)\n", err)
	}

	// Return the last error if we only fail
	return fmt.Sprintf("Error getting Notion Digest\n<BLOCKQUOTE>Error Message: %v", err)
}

// helper function to get the Notion Daily Digest with associated errors
func getNotionDailyDigest(c *config.Config) (string, error) {
	var output string

	loc := time.UTC

	// Find 'today' in the calendar's timezone (assuming there is only one)
	year, month, day := time.Now().In(loc).Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, loc)

	// Get calendar information
	calendarEvents := []Event{}
	for _, calendar := range calendars {
		for _, event := range calendar.Events() {
			// Get the name
			name := event.GetProperty(ics.ComponentPropertySummary)

			// Get all times of the event
			start, _ := event.GetStartAt()
			end, endNotPresent := event.GetEndAt()
			startDay, _ := event.GetAllDayStartAt()
			endDay, endDayNotPresent := event.GetAllDayEndAt()

			// Determine whether this is an all day event
			allDayEvent := start.Equal(startDay) && end.Equal(endDay)
			allDayEvent = allDayEvent || endNotPresent != nil && endDayNotPresent != nil // Weird hack for events without end saved -> are all day events

			// Edit the start date if there are repeat rules
			rr := event.GetProperty(ics.ComponentPropertyRrule)
			repeating := rr != nil

			if repeating {
				// Get the recurring rule
				rule, err := rrule.StrToRRule(rr.BaseProperty.Value)
				if err != nil {
					continue // Skip on error, assume malformatted data
				}

				if allDayEvent {
					// For all day events, set to the base date of start
					year, month, day := start.Date()
					rule.DTStart(time.Date(year, month, day, 0, 0, 0, 0, loc))
				} else {
					// For normal events, just set to start
					rule.DTStart(start)
				}

				// Calculate the next occurence
				start = rule.After(today, true)
				startDay = start
				endDay = startDay.Add(24 * time.Hour)

				// Check if the next occurence
				for _, prop := range event.Properties {
					if prop.IANAToken == "EXDATE" {
						// Get the time from the EXDATE
						t, err := time.Parse("20060102T150405", prop.Value)
						if err != nil {
							continue
						}
						t = t.In(loc)

						// If time is equal to start time reject
						if start.Truncate(24 * time.Hour).Equal(t.Truncate(24 * time.Hour)) {
							start = time.Time{}
							break
						}
					}
				}
			}

			if allDayEvent {
				// Convert the start and end days into a form that compares well with 'today'
				startDay = startDay.In(loc)
				endDay = endDay.In(loc)

				// Add an all day event
				if (today.After(startDay) || today.Equal(startDay)) && today.Before(endDay) {
					calendarEvents = append(calendarEvents, Event{
						Start:  start,
						AllDay: true,
						Format: fmt.Sprintf("- All Day: %v\n", name.BaseProperty.Value),
					})
				}
			} else {
				// Normal days have no errors from end days or end times
				if today.Day() == start.Day() && today.Month() == start.Month() && today.Year() == start.Year() {
					calendarEvents = append(calendarEvents, Event{
						Start:  start,
						AllDay: false,
						Format: fmt.Sprintf("- %v → %v: %v\n", start.In(formatLoc).Format("03:04 PM"), end.In(formatLoc).Format("03:04 PM"), name.BaseProperty.Value),
					})
				}
			}
		}
	}

	// Get the schedule database
	schedule, err := c.Notion.QueryDatabase(context.Background(), SCHEDULE_ITEMS.ID, &SCHEDULE_ITEMS.Query)
	if err != nil {
		return "", err
	}

	// Loop for each task page
	for _, p := range schedule.Results {
		// Get the page property IDs from Notion
		page, err := c.Notion.FindPageByID(context.Background(), p.ID)
		if err != nil {
			return "", err
		}

		// Get the page property values from their IDs
		properties, ok := page.Properties.(notionapi.DatabasePageProperties)
		if !ok {
			return "", err
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
		calendarEvents = append(calendarEvents, Event{
			Start:  start,
			Format: fmt.Sprintf("- %v: %v\n", timespan, name),
		})
	}

	// Sort calendar events
	for i := 1; i < len(calendarEvents); i++ {
		event := calendarEvents[i]
		j := i - 1

		for j >= 0 && calendarEvents[j].Start.Compare(event.Start) > 0 {
			calendarEvents[j+1] = calendarEvents[j]
			j--
		}
		calendarEvents[j+1] = event
	}

	// Add calendar events to the output
	if len(schedule.Results) != 0 {
		output += "<STRONG>Schedule:<STRONG>\n"
	}

	for _, event := range calendarEvents {
		output += event.Format
	}

	// Get the tasks page
	tasks, err := c.Notion.QueryDatabase(context.Background(), NORMAL_TASKS.ID, &NORMAL_TASKS.Query)
	if err != nil {
		return "", err
	}

	if len(tasks.Results) != 0 {
		output += "\n<STRONG>Upcoming Tasks:<STRONG>\n"
	}

	// Loop for each task page
	for _, p := range tasks.Results {
		// Get the page property IDs from Notion
		page, err := c.Notion.FindPageByID(context.Background(), p.ID)
		if err != nil {
			return "", err
		}

		// Get the page property values from their IDs
		properties, ok := page.Properties.(notionapi.DatabasePageProperties)
		if !ok {
			return "", err
		}

		// Get the name of the task
		nameField := properties["Name"]
		if len(nameField.Title) == 0 {
			continue
		}
		name := nameField.Title[0].Text.Content

		// Get the project of the task
		projectField := properties["Project Name"]

		project := ""
		if projectField.Formula != nil && projectField.Formula.String != nil {
			project = *projectField.Formula.String
		}

		if project != "" {
			project = "<EM>" + project + "<EM>"
		}

		// Get the date of the task
		dateField := properties["Date"]
		date := ""
		if dateField.Date != nil {
			date = "(" + dateField.Date.Start.Format("Mon Jan 2") + ")"
		}

		output += fmt.Sprintf("- %v %v %v\n", name, project, date)
	}

	// Get the tasks page
	criticalTasks, err := c.Notion.QueryDatabase(context.Background(), CRITICAL_TASKS.ID, &CRITICAL_TASKS.Query)
	if err != nil {
		return "", err
	}

	if len(criticalTasks.Results) != 0 {
		output += "\n<STRONG>Critical Tasks:<STRONG>\n"
	}

	// Loop for each task page
	for _, p := range criticalTasks.Results {
		// Get the page property IDs from Notion
		page, err := c.Notion.FindPageByID(context.Background(), p.ID)
		if err != nil {
			return "", err
		}

		// Get the page property values from their IDs
		properties, ok := page.Properties.(notionapi.DatabasePageProperties)
		if !ok {
			return "", err
		}

		// Get the name of the task
		nameField := properties["Name"]
		if len(nameField.Title) == 0 {
			continue
		}
		name := nameField.Title[0].Text.Content

		// Get the project of the task
		projectField := properties["Tasks -> Project Name"]

		project := ""
		if projectField.Formula != nil && projectField.Formula.String != nil {
			project = *projectField.Formula.String
		}

		if project != "" {
			project = "<EM>" + project + "<EM>"
		}

		// Get the date of the task
		dateField := properties["Date"]
		date := ""
		if dateField.Date != nil {
			date = "(" + dateField.Date.Start.Format("Mon Jan 2") + ")"
		}

		output += fmt.Sprintf("- %v %v %v\n", name, project, date)
	}

	// Get the recurring database sections
	for _, t := range []string{"Connection", "Habit", "Chore"} {
		RECURRING_TASKS.Query.Filter.And[3].Select.Equals = t

		recurring, err := c.Notion.QueryDatabase(context.Background(), RECURRING_TASKS.ID, &RECURRING_TASKS.Query)
		if err != nil {
			return "", err
		}

		if len(recurring.Results) != 0 {
			output += fmt.Sprintf("\n<STRONG>%vs:<STRONG>\n", t)
		}

		// Loop for each task page
		for _, p := range recurring.Results {
			// Get the page property IDs from Notion
			page, err := c.Notion.FindPageByID(context.Background(), p.ID)
			if err != nil {
				return "", err
			}

			// Get the page property values from their IDs
			properties, ok := page.Properties.(notionapi.DatabasePageProperties)
			if !ok {
				return "", err
			}

			// Get the name of the task
			nameField := properties["Name"]
			if len(nameField.Title) == 0 {
				continue
			}
			name := nameField.Title[0].Text.Content

			output += fmt.Sprintf("- %v\n", name)
		}
	}

	return output, nil
}

// NotionNightAffirmations returns the night affirmations stored in Notion
func NotionNightAffirmations(c *config.Config) string {
	var output string
	var err error

	// Run for a given number of times
	for i := 0; i < DIGEST_ERROR_LIMIT; i++ {
		output, err = getAffirmations(c, NIGHT_AFFIRMATIONS_PAGE)

		// On no error, return the output
		if err == nil {
			return output
		}
		log.Printf("[ERROR]: Error in notion/NotionNightAffrimations, retrying (err: %v)\n", err)
	}

	// Return the last error if we only fail
	return fmt.Sprintf("Error getting Notion Night Affirmations\n<BLOCKQUOTE>Error Message: %v", err)
}

// NotionMorningAffirmations returns the morning affirmations stored in Notion
func NotionMorningAffirmations(c *config.Config) string {
	var output string
	var err error

	// Run for a given number of times
	for i := 0; i < DIGEST_ERROR_LIMIT; i++ {
		output, err = getAffirmations(c, MORNING_AFFIRMATIONS_PAGE)

		// On no error, return the output
		if err == nil {
			return output
		}
		log.Printf("[ERROR]: Error in notion/NotionMorningAffrimations, retrying (err: %v)\n", err)
	}

	// Return the last error if we only fail
	return fmt.Sprintf("Error getting Notion Morning Affirmations\n<BLOCKQUOTE>Error Message: %v", err)
}

// getAffirmations sends the night affirmations page in Notion
func getAffirmations(c *config.Config, id string) (string, error) {
	// Get the page from notion
	page, err := c.Notion.FindBlockChildrenByID(context.Background(), id, &notionapi.PaginationQuery{
		StartCursor: "",
		PageSize:    100,
	})
	if err != nil {
		return "", err
	}

	// For each block contribute to the output
	output := "<STRONG>Affirmations:<STRONG>\n\n"
	for _, raw := range page.Results {
		switch block := raw.(type) {
		// For headings, add strong text
		case *notionapi.Heading1Block:
			if len(block.RichText) > 0 {
				output += fmt.Sprintf("<STRONG>%v<STRONG>\n", block.RichText[0].Text.Content)
			}

		case *notionapi.Heading2Block:
			if len(block.RichText) > 0 {
				output += fmt.Sprintf("<STRONG>%v<STRONG>\n", block.RichText[0].Text.Content)
			}

		case *notionapi.Heading3Block:
			if len(block.RichText) > 0 {
				output += fmt.Sprintf("<STRONG>%v<STRONG>\n", block.RichText[0].Text.Content)
			}

		// For normal text
		case *notionapi.ParagraphBlock:
			if len(block.RichText) > 0 {
				output += fmt.Sprintf("%v\n", block.RichText[0].Text.Content)
			}

		// For unknown blocks, print out a log message
		default:
			log.Printf("[WARN]: unhandled block type %v\n", reflect.TypeOf(block))
		}
	}

	return output, nil
}

/* ---- HELPER FUNCTIONS ---- */

func intPointer(num int) *int {
	return &num
}

func boolPointer(val bool) *bool {
	return &val
}

func timePointer(t time.Time) *time.Time {
	return &t
}
