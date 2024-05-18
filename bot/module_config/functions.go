package module_config

import (
	"fmt"

	horus "github.com/ethanbaker/horus/bot"
	"github.com/ethanbaker/horus/utils/types"
)

// A list of all enabled functions in the module
var functions = map[string]func(bot *horus.Bot, input *types.Input) any{
	"set_timezone":         set_timezone,
	"set_city":             set_city,
	"set_temperature_unit": set_temperature_unit,
}

// Set the user's preferred timezone
func set_timezone(bot *horus.Bot, input *types.Input) any {
	// Get the timezone from the user
	timezone, ok := input.GetString("timezone", "")
	if !ok {
		return fmt.Errorf(`{"error": "timezone not formatted correctly"}`)
	}

	// Save the timezone
	bot.Memory.Timezone = timezone

	return `{"message": "successfully saved timezone"}`
}

// Set the user's home city
func set_city(bot *horus.Bot, input *types.Input) any {
	// Get the city from the user
	city, ok := input.GetString("city", "")
	if !ok {
		return fmt.Errorf(`{"error": "city not formatted correctly"}`)
	}

	// Save the timezone
	bot.Memory.City = city

	return `{"message": "successfully saved city"}`
}

// Set the user's preferred temperature unit
func set_temperature_unit(bot *horus.Bot, input *types.Input) any {
	// Get the temperature_unit from the user
	unit, ok := input.GetString("unit", "")
	if !ok {
		return fmt.Errorf(`{"error": "unit not formatted correctly"}`)
	}

	// Save the timezone
	bot.Memory.TemperatureUnit = unit

	return `{"message": "successfully saved unit"}`
}
