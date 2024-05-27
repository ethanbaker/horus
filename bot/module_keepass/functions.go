package module_keepass

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	horus "github.com/ethanbaker/horus/bot"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/ethanbaker/horus/utils/validation"
)

/* ---- TYPES ---- */

// Represents a keepass profile
type Profile struct {
	Path     string `json:"path"`
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	Url      string `json:"url"`
	Notes    string `json:"notes"`
}

// Represents a message from the API
type apiError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

/* ---- GLOBAL CONSTANTS ---- */

// Confirmation message
const CONFIRM_MESSAGE = `Value saved successfully.

<STRONG>Title<STRONG>: %v
<STRONG>Path<STRONG>: %v
<STRONG>Username<STRONG>: %v
<STRONG>Password<STRONG>: <SPOILER>%v<SPOILER>
<STRONG>URL<STRONG>: %v
<STRONG>Notes<STRONG>:
%v

Is this the profile you want to save?`

// Step names to update
var stepNames []string = []string{"path", "username", "password", "URL", "notes (type 'none' if empty)"}

// Regex testing
var titleRegex = regexp.MustCompile(`\W`)
var pathRegex = regexp.MustCompile(`^(/[a-z-]+)+/?$`)

/* ---- FUNCTIONS ---- */

// A list of all enabled functions in the module
var functions = map[string]func(bot *horus.Bot, input *types.Input) any{
	"keepass_get":    get_keepass,
	"keepass_create": create_keepass,
	"keepass_update": update_keepass,
	"keepass_delete": delete_keepass,
}

// Start the get keepass process
func get_keepass(bot *horus.Bot, input *types.Input) any {
	// Get the keepass database
	client := &http.Client{}

	req, err := http.NewRequest("GET", bot.Config.Getenv("KEEPASS_BASE_URL"), nil)
	if err != nil {
		return &types.Output{Error: errors.New("cannot create http request")}
	}
	req.Header.Set("token", bot.Config.Getenv("KEEPASS_TOKEN"))

	res, err := client.Do(req)
	if err != nil {
		return &types.Output{Error: errors.New("error fetching database")}
	}

	// Read the file content
	body, err := io.ReadAll(res.Body)
	if err != nil || len(body) == 0 {
		return &types.Output{Error: errors.New("error reading database file")}
	}

	// Send the output to the user
	output := types.Output{}
	output.Message = "File successfully sent!"
	output.Data = types.FileOutput{Filename: "database.kdbx", Content: body}

	return &output
}

// Start the create keepass process
func create_keepass(bot *horus.Bot, input *types.Input) any {
	// Initialize the profile
	bot.EditVariable("keepass_profile", Profile{})
	bot.EditVariable("keepass_index", 0)

	// Start the step process
	bot.AddQueuedFunctions(create_keepass_step)

	// Return a success message
	return &types.Output{Message: "New password profile started. Please enter the title: "}
}

// Helper method to repeatedly get information from a password
func create_keepass_step(bot *horus.Bot, input *types.Input) *types.Output {
	output := types.Output{}

	// Get the saved profile
	profile, ok := bot.GetVariable("keepass_profile").(Profile)
	if !ok {
		output.Error = errors.New("cannot get saved keepass profile")
		return &output
	}

	// Get the saved index
	idx, ok := bot.GetVariable("keepass_index").(int)
	if !ok {
		output.Error = errors.New("cannot get saved keepass index")
		return &output
	}

	// Update the correlaing field
	switch idx {

	// Update title
	case 0:
		if !titleRegex.MatchString(input.Message) {
			profile.Title = input.Message
		} else {
			output.Error = errors.New("invalid title")
		}

	// Update path
	case 1:
		if pathRegex.MatchString(input.Message) {
			profile.Path = input.Message
		} else {
			output.Error = errors.New("invalid path")
		}

	// Update username
	case 2:
		profile.Username = input.Message

	// Update password
	case 3:
		profile.Password = input.Message

	// Update bot.Config.Getenv("KEEPASS_BASE_URL")
	case 4:
		profile.Url = input.Message

	// Update notes
	case 5:
		if input.Message != "none" {
			profile.Notes = input.Message
		} else {
			profile.Notes = ""
		}
	}

	// If there is an error return
	if output.Error != nil {
		return &output
	}

	// If there is a next item, save and prompt the user for the next field
	bot.EditVariable("keepass_profile", profile)
	if idx < len(stepNames) {
		bot.EditVariable("keepass_index", idx+1)
		bot.AddQueuedFunctions(create_keepass_step)

		output.Message = fmt.Sprintf(`Value saved successfully. Please enter the %v:`, stepNames[idx])
		return &output
	}

	// Otherwise, ask the user for confirmation
	bot.AddQueuedFunctions(create_keepass_confirm)

	output.Message = fmt.Sprintf(CONFIRM_MESSAGE, profile.Title, profile.Path, profile.Username, profile.Password, profile.Url, profile.Notes)
	return &output
}

// Helper method to ask for confirmation creating a new profile
func create_keepass_confirm(bot *horus.Bot, input *types.Input) *types.Output {
	output := types.Output{}

	// Check for a yes
	if !validation.ValidateConfirmation(input.Message) {
		output.Message = "Password profile creation abandoned."
		return &output
	}

	// Get the saved profile
	profile, ok := bot.GetVariable("keepass_profile").(Profile)
	if !ok {
		output.Error = errors.New("cannot get saved keepass profile")
		return &output
	}

	// Send password profile to the API
	reqBody, err := json.Marshal(profile)
	if err != nil {
		output.Error = err
		return &output
	}

	// Send the request
	req, err := http.NewRequest("POST", bot.Config.Getenv("KEEPASS_BASE_URL"), bytes.NewBuffer(reqBody))
	if err != nil {
		output.Error = err
		return &output
	}
	req.Header.Set("token", bot.Config.Getenv("KEEPASS_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Error = err
		return &output
	}
	defer resp.Body.Close()

	// Get the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Error = err
		return &output
	}

	// Unmarshal the reponse
	e := apiError{}
	err = json.Unmarshal(body, &e)
	if err != nil {
		output.Error = err
		return &output
	}

	// Look for API errors
	if e.Error {
		bot.AddQueuedFunctions(create_keepass_confirm)

		output.Message = "There was an error saving your password. Try again?"
		output.Error = errors.New(e.Message)

		return &output
	}

	// Return success
	output.Message = "Password profile created successfully!"
	return &output
}

// Start the create keepass process
func update_keepass(bot *horus.Bot, input *types.Input) any {
	// Initialize the profile
	bot.EditVariable("keepass_profile", Profile{})
	bot.EditVariable("keepass_index", 0)

	// Start the step process
	bot.AddQueuedFunctions(update_keepass_step)

	// Return a success message
	return &types.Output{Message: "Updated password profile started. Please enter the title: "}
}

// Helper method to repeatedly get information from a password
func update_keepass_step(bot *horus.Bot, input *types.Input) *types.Output {
	output := types.Output{}

	// Get the saved profile
	profile, ok := bot.GetVariable("keepass_profile").(Profile)
	if !ok {
		output.Error = errors.New("cannot get saved keepass profile")
		return &output
	}

	// Get the saved index
	idx, ok := bot.GetVariable("keepass_index").(int)
	if !ok {
		output.Error = errors.New("cannot get saved keepass index")
		return &output
	}

	// Update the correlaing field
	switch idx {

	// Update title
	case 0:
		if !titleRegex.MatchString(input.Message) {
			profile.Title = input.Message
		} else {
			output.Error = errors.New("invalid title")
		}

	// Update path
	case 1:
		if pathRegex.MatchString(input.Message) {
			profile.Path = input.Message
		} else {
			output.Error = errors.New("invalid path")
		}

	// Update username
	case 2:
		profile.Username = input.Message

	// Update password
	case 3:
		profile.Password = input.Message

	// Update bot.Config.Getenv("KEEPASS_BASE_URL")
	case 4:
		profile.Url = input.Message

	// Update notes
	case 5:
		if input.Message != "none" {
			profile.Notes = input.Message
		} else {
			profile.Notes = ""
		}
	}

	// If there is an error return
	if output.Error != nil {
		return &output
	}

	// If there is a next item, save and prompt the user for the next field
	bot.EditVariable("keepass_profile", profile)
	if idx < len(stepNames) {
		bot.EditVariable("keepass_index", idx+1)
		bot.AddQueuedFunctions(update_keepass_step)

		output.Message = fmt.Sprintf(`Value saved successfully. Please enter the %v:`, stepNames[idx])
		return &output
	}

	// Otherwise, ask the user for confirmation
	bot.AddQueuedFunctions(update_keepass_confirm)

	output.Message = fmt.Sprintf(CONFIRM_MESSAGE, profile.Title, profile.Path, profile.Username, profile.Password, profile.Url, profile.Notes)
	return &output
}

// Helper method to ask for confirmation creating a new profile
func update_keepass_confirm(bot *horus.Bot, input *types.Input) *types.Output {
	output := types.Output{}

	// Check for a yes
	if !validation.ValidateConfirmation(input.Message) {
		output.Message = "Password profile creation abandoned."
		return &output
	}

	// Get the saved profile
	profile, ok := bot.GetVariable("keepass_profile").(Profile)
	if !ok {
		output.Error = errors.New("cannot get saved keepass profile")
		return &output
	}

	// Send password profile to the API
	reqBody, err := json.Marshal(profile)
	if err != nil {
		output.Error = err
		return &output
	}

	// Send the request
	req, err := http.NewRequest("PUT", bot.Config.Getenv("KEEPASS_BASE_URL"), bytes.NewBuffer(reqBody))
	if err != nil {
		output.Error = err
		return &output
	}
	req.Header.Set("token", bot.Config.Getenv("KEEPASS_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Error = err
		return &output
	}
	defer resp.Body.Close()

	// Get the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Error = err
		return &output
	}

	// Unmarshal the response
	e := apiError{}
	err = json.Unmarshal(body, &e)
	if err != nil {
		output.Error = err
		return &output
	}

	// Look for API errors
	if e.Error {
		bot.AddQueuedFunctions(update_keepass_confirm)

		output.Message = "There was an error saving your password. Try again?"
		output.Error = errors.New(e.Message)

		return &output
	}

	// Return success
	output.Message = "Password profile updated successfully!"
	return &output
}

// Start the create keepass process
func delete_keepass(bot *horus.Bot, input *types.Input) any {
	// Initialize the profile
	bot.EditVariable("keepass_profile", Profile{})
	bot.EditVariable("keepass_index", 0)

	// Start the step process
	bot.AddQueuedFunctions(delete_keepass_step)

	// Return a success message
	return &types.Output{Message: "Delete password profile started. Please enter the title: "}
}

// Helper method to repeatedly get information from a password
func delete_keepass_step(bot *horus.Bot, input *types.Input) *types.Output {
	output := types.Output{}

	// Get the saved profile
	profile, ok := bot.GetVariable("keepass_profile").(Profile)
	if !ok {
		output.Error = errors.New("cannot get saved keepass profile")
		return &output
	}

	// Get the saved index
	idx, ok := bot.GetVariable("keepass_index").(int)
	if !ok {
		output.Error = errors.New("cannot get saved keepass index")
		return &output
	}

	// Update the correlaing field
	switch idx {

	// Update title
	case 0:
		if !titleRegex.MatchString(input.Message) {
			profile.Title = input.Message
		} else {
			output.Error = errors.New("invalid title")
		}

	// Update path
	case 1:
		if pathRegex.MatchString(input.Message) {
			profile.Path = input.Message
		} else {
			output.Error = errors.New("invalid path")
		}
	}

	// If there is an error return
	if output.Error != nil {
		return &output
	}

	// If there is a next item, save and prompt the user for the next field
	bot.EditVariable("keepass_profile", profile)
	if idx == 0 {
		bot.EditVariable("keepass_index", idx+1)
		bot.AddQueuedFunctions(delete_keepass_step)

		output.Message = fmt.Sprintf(`Value saved successfully. Please enter the %v:`, stepNames[idx])
		return &output
	}

	// Otherwise, ask the user for confirmation
	bot.AddQueuedFunctions(delete_keepass_confirm)

	output.Message = fmt.Sprintf(`Are you sure you want to delete <STRONG>%v<STRONG>?`, profile.Title)
	return &output
}

// Helper method to ask for confirmation creating a new profile
func delete_keepass_confirm(bot *horus.Bot, input *types.Input) *types.Output {
	output := types.Output{}

	// Check for a yes
	if !validation.ValidateConfirmation(input.Message) {
		output.Message = "Password profile creation abandoned."
		return &output
	}

	// Get the saved profile
	profile, ok := bot.GetVariable("keepass_profile").(Profile)
	if !ok {
		output.Error = errors.New("cannot get saved keepass profile")
		return &output
	}

	// Send password profile to the API
	reqBody, err := json.Marshal(profile)
	if err != nil {
		output.Error = err
		return &output
	}

	// Send the request
	req, err := http.NewRequest("DELETE", bot.Config.Getenv("KEEPASS_URL"), bytes.NewBuffer(reqBody))
	if err != nil {
		output.Error = err
		return &output
	}
	req.Header.Set("token", bot.Config.Getenv("KEEPASS_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Error = err
		return &output
	}
	defer resp.Body.Close()

	// Get the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output.Error = err
		return &output
	}

	// Unmarshal the response
	e := apiError{}
	err = json.Unmarshal(body, &e)
	if err != nil {
		output.Error = err
		return &output
	}

	// Look for API errors
	if e.Error {
		bot.AddQueuedFunctions(delete_keepass_confirm)

		output.Message = "There was an error deleting your password. Try again?"
		output.Error = errors.New(e.Message)

		return &output
	}

	// Return success
	output.Message = "Password profile deleted successfully!"
	return &output
}
