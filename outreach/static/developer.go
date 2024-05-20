package static

import "github.com/ethanbaker/horus/utils/config"

// Ping is a testing function that sends the user "Ping!"
func Ping(_ *config.Config) string {
	return "Ping!"
}
