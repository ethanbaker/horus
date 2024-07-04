package outreach

import (
	"fmt"
	"time"

	"github.com/ethanbaker/horus/outreach/dynamic"
	"github.com/ethanbaker/horus/outreach/static"
	"github.com/ethanbaker/horus/utils/config"
	"github.com/ethanbaker/horus/utils/types"
	"github.com/robfig/cron/v3"
)

/* ---- GLOBALS ---- */

// The global manager of outreach messages
var manager = types.OutreachManager{}

/* ---- METHODS ---- */

// Setup outreach
func Setup(config *config.Config) error {
	// Setup maps in the manager
	manager.Modules = map[types.OutreachModule]func(*types.OutreachManager, []chan string, map[string]any) (types.OutreachMessage, error){
		types.DynamicModule: dynamic.New,
		types.StaticModule:  static.New,
	}
	manager.Channels = map[types.OutreachMethod]chan string{}
	manager.Messages = map[string]types.OutreachMessage{}

	// Create the cron service
	c := cron.New()
	c.Start()

	// Populate the services struct
	manager.Services = &types.OutreachServices{
		Config: config,
		DB:     config.Gorm,
		Cron:   c,
		Clock:  time.NewTicker(time.Minute),
	}

	// Run init functions in submodules
	if err := static.Init(config); err != nil {
		return err
	}
	if err := dynamic.Init(config); err != nil {
		return err
	}

	return nil
}

// Add a channel to outreach
func AddChannel(method types.OutreachMethod) (chan string, error) {
	// Make sure a channel with the same name doesn't exist
	_, ok := manager.Channels[method]
	if ok {
		return nil, fmt.Errorf("channel with key %v already exists", method)
	}

	// Make a new channel
	ch := make(chan string)
	manager.Channels[method] = ch
	return ch, nil
}

// Start an outreach message
func Start(name string) error {
	m, ok := manager.Messages[name]
	if !ok {
		return fmt.Errorf("message with name '%v' cannot be found", name)
	}

	return m.Start()
}

// Stop an outreach message
func Stop(name string) error {
	m, ok := manager.Messages[name]
	if !ok {
		return fmt.Errorf("message with name '%v' cannot be found", name)
	}

	return m.Stop()
}

// Delete an outreach message
func Delete(name string) error {
	m, ok := manager.Messages[name]
	if !ok {
		return fmt.Errorf("message with name '%v' cannot be found", name)
	}

	return m.Delete()
}

// New creates a new outreach message from an associated module, to given channel keys, with a given name and data
func New(module types.OutreachModule, name string, methods []types.OutreachMethod, data map[string]any) error {
	// Make sure the name is valid
	_, ok := manager.Messages[name]
	if ok || name == "" {
		return fmt.Errorf("invalid name '%v'", name)
	}

	// Get the submodule
	f, ok := manager.Modules[module]
	if !ok {
		return fmt.Errorf("requested module '%v' does not exist", module)
	}

	// Get a list of channels for the message
	chans := []chan string{}
	for _, k := range methods {
		// Make sure the channel exists and then add it
		c, ok := manager.Channels[k]
		if !ok {
			return fmt.Errorf("channel with key %v does not exist", k)
		}

		chans = append(chans, c)
	}

	// There should be at least one channel to send the outreach to
	if len(chans) == 0 {
		return fmt.Errorf("no channels selected")
	}

	// Create the new outreach message from the submodule
	m, err := f(&manager, chans, data)
	if err != nil {
		return err
	}

	// Add the message to the outreach message manager
	manager.Messages[name] = m
	return nil
}
