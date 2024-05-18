package types

type CalendarConfig struct {
	Calendars []struct {
		URL string `yaml:"url"`
	} `yaml:"calendars"`
	TimezoneFormat string `yaml:"timezone-format"`
}
