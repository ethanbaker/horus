package horus

import "gorm.io/gorm"

// Memory represents a static memory bank used by a Bot. Values in this struct get stored to SQL and are saved past reboot
type Memory struct {
	gorm.Model

	BotID           uint // The foreign key to relate the Memory struct to the Bot
	Timezone        string
	City            string
	TemperatureUnit string
}
