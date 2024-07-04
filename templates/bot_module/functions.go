package module_template

import (
	horus "github.com/ethanbaker/horus/bot"
	"github.com/ethanbaker/horus/utils/types"
)

// A list of all enabled functions in the module
var functions = map[string]func(bot *horus.Bot, input *types.Input) any{
	"function_1":               function_1,
	"get_current_weather_demo": get_current_weather_demo,
}

// Function takes in input from the bot and returns an object that can be json
// marshalled to the model later
func function_1(bot *horus.Bot, input *types.Input) any {
	return nil
}

func get_current_weather_demo(bot *horus.Bot, input *types.Input) any {
	// Get variables from the model using parameters
	location, _ := input.GetString("location", "")
	unit, _ := input.GetString("unit", "")

	// Create demo data to return
	var weatherInfo struct {
		Location    string   `json:"location"`
		Temperature string   `json:"temperature"`
		Unit        string   `json:"unit"`
		Forecast    []string `json:"forecast"`
	}

	weatherInfo.Location = location
	weatherInfo.Temperature = "72"
	weatherInfo.Unit = unit
	weatherInfo.Forecast = []string{"sunny", "windy"}

	return weatherInfo
}
