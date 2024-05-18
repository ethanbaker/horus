package outreach

import (
	"fmt"
	"time"

	"github.com/ethanbaker/horus/outreach/dynamic"
	"github.com/ethanbaker/horus/outreach/static"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/* ---- GLOBALS ---- */

// A list of allowed submodules
var enabledSubmodules = map[string]func(*types.OutreachServices, []chan string, any) (types.OutreachMessage, error){
	"static":  static.New,
	"dynamic": dynamic.New,
}

// Services we can pass to a submodule
var services = types.OutreachServices{}

// Added channels outreach can send messages to
var channels = map[types.OutreachMethod]chan string{}

/* ---- METHODS ---- */

// Setup outreach
func Setup(dsn string) error {
	// Create the database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// Create the cron service
	c := cron.New()
	c.Start()

	// Create a global clock
	services.Clock = time.NewTicker(time.Minute)

	// Add them to the services
	services.DB = db
	services.Cron = c

	// Run init functions in submodules
	if err = static.Init(); err != nil {
		return err
	}
	if err = dynamic.Init(); err != nil {
		return err
	}

	return nil
}

// Add a channel to outreach
func AddChannel(key types.OutreachMethod) (chan string, error) {
	// Make sure a channel with the same name doesn't exist
	_, ok := channels[key]
	if ok {
		return nil, fmt.Errorf("channel with key %v already exists", key)
	}

	// Make a new channel
	ch := make(chan string)
	channels[key] = ch
	return ch, nil
}

// New creates a new message for an associated submodule key
func New(module string, keys []types.OutreachMethod, data any) (types.OutreachMessage, error) {
	f, ok := enabledSubmodules[module]
	if !ok {
		return nil, fmt.Errorf("requested message key %v does not exist", module)
	}

	// Compile a list of channels
	chans := []chan string{}
	for _, k := range keys {
		// Make sure the channel exists and then add it
		c, ok := channels[k]
		if !ok {
			return nil, fmt.Errorf("channel with key %v does not exist", k)
		}

		chans = append(chans, c)
	}

	// Make sure there is more than 0 channels
	if len(chans) == 0 {
		return nil, fmt.Errorf("no channels selected")
	}

	return f(&services, chans, data)
}
