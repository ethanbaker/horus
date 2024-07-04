package module_ambient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	horus "github.com/ethanbaker/horus/bot"
	"github.com/ethanbaker/horus/utils/types"
)

// A list of all enabled functions in the module
var functions = map[string]func(bot *horus.Bot, input *types.Input) any{
	"get_current_time":    get_time,
	"get_current_weather": get_weather,
}

// Get the current time
func get_time(bot *horus.Bot, input *types.Input) any {
	// Type to hold time information
	var timeInformation struct {
		Year    string `json:"year"`
		Month   string `json:"month"`
		Day     string `json:"day"`
		Weekday string `json:"weekday"`
		Hour    string `json:"hour"`
		Minute  string `json:"minute"`
		Second  string `json:"second"`
	}

	// Get the user's timezone
	loc, err := time.LoadLocation(bot.Memory.Timezone)
	if err != nil {
		return `{"error": "could not load timezone"}`
	}

	t := time.Now().In(loc)

	timeInformation.Year = fmt.Sprint(t.Year())
	timeInformation.Month = fmt.Sprint(t.Month())
	timeInformation.Day = fmt.Sprint(t.Day())
	timeInformation.Weekday = fmt.Sprint(t.Weekday())
	timeInformation.Hour = fmt.Sprint(t.Hour())
	timeInformation.Minute = fmt.Sprint(t.Minute())
	timeInformation.Second = fmt.Sprint(t.Second())

	return timeInformation
}

// Get the current weather
func get_weather(bot *horus.Bot, input *types.Input) any {
	// Openweather map request data
	type openweathermapData struct {
		Coord struct {
			Lon float64 `json:"lon"`
			Lat float64 `json:"lat"`
		} `json:"coord"`

		Weather []struct {
			Id          int    `json:"id"`
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`

		Base string `json:"base"`

		Main struct {
			Temp        float64 `json:"temp"`
			FeelsLike   float64 `json:"feels_like"`
			TempMin     float64 `json:"temp_min"`
			TempMax     float64 `json:"temp_max"`
			Pressure    int     `json:"pressure"`
			Humidity    int     `json:"humidity"`
			SeaLevel    int     `json:"sea_level"`
			GroundLevel int     `json:"grnd_level"`
		} `json:"main"`

		Visibility int `json:"visibility"`

		Wind struct {
			Speed  float64 `json:"speed"`
			Degree int     `json:"deg"`
			Gust   float64 `json:"gust"`
		} `json:"wind"`

		Clouds struct {
			CoverPercent int `json:"all"`
		} `json:"clouds"`

		Rain struct {
			OneHour   float64 `json:"1h,omitempty"`
			ThreeHour float64 `json:"3h,omitempty"`
		} `json:"rain"`

		Snow struct {
			OneHour   float64 `json:"1h,omitempty"`
			ThreeHour float64 `json:"3h,omitempty"`
		} `json:"snow"`

		CalculationTime int `json:"dt"`

		System struct {
			Type    int    `json:"type"`
			Id      int    `json:"id"`
			Country string `json:"country"`
			Sunrise int    `json:"sunrise"`
			Sunset  int    `json:"sunset"`
		} `json:"sys"`

		Timezone int    `json:"timezone"`
		Id       int    `json:"id"`
		Name     string `json:"name"`
		Cod      int    `json:"cod"`
	}

	// Weather data to send
	type weatherData struct {
		Overview    string  `json:"overview"`
		Description string  `json:"description"`
		Temperature float64 `json:"temperature"`
		FeelsLike   float64 `json:"feels_like"`
		MaxTemp     float64 `json:"max_temperature"`
		MinTemp     float64 `json:"min_temperature"`
		Humidity    int     `json:"humidity_percent"`
		WindSpeed   float64 `json:"wind_speed"`
		Cloudiness  int     `json:"cloud_cover_percent"`
	}

	// Get the parameters
	location, _ := input.GetString("location", "")
	unit, _ := input.GetString("unit", "")

	if location == "" && bot.Memory.City != "" {
		location = bot.Memory.City
	} else {
		location = "Davidson"
	}

	if unit == "" && bot.Memory.TemperatureUnit != "" {
		unit = bot.Memory.TemperatureUnit
	} else {
		unit = "celsius"
	}

	// Get the weather conditions
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s/data/2.5/weather?q=%s&appid=%s", bot.Config.Getenv("WEATHER_BASE_URL"), location, bot.Config.Getenv("WEATHER_TOKEN"))

	// Send the request
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf(`{"error": "could not access weather database"}`)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf(`{"error": "could not find location '%v'"}`, location)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(`{"error": "unexpected response status %v"}`, resp.Status)
	}

	// Get the data into a struct
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(`{"error": "response from weather database is unreadable"}`)
	}

	var data openweathermapData
	if err = json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf(`{"error": "response from weather database is not formatted correctly"}`)
	}

	// Check for errors
	if len(data.Weather) < 1 {
		return errors.New(`{"error": "no weather elements available for location"}`)
	}

	// Setup a conversion factor based on the requested unit
	tempConversion := func(temp float64) float64 {
		return (temp-273.15)*9/5 + 32
	}
	if unit == "celsius" {
		tempConversion = func(temp float64) float64 {
			return temp - 273.15
		}
	}

	conditions := weatherData{
		Overview:    data.Weather[0].Main,
		Description: data.Weather[0].Description,
		Temperature: tempConversion(data.Main.Temp),
		FeelsLike:   tempConversion(data.Main.FeelsLike),
		MaxTemp:     tempConversion(data.Main.TempMax),
		MinTemp:     tempConversion(data.Main.TempMin),
		Humidity:    data.Main.Humidity,
		WindSpeed:   data.Wind.Speed,
		Cloudiness:  data.Clouds.CoverPercent,
	}

	return conditions
}
