package dynamic_test

import (
	"log"
	"testing"
	"time"

	"github.com/ethanbaker/horus/outreach/dynamic"
	"github.com/ethanbaker/horus/utils/config"
	"github.com/stretchr/testify/assert"
)

func TestNotionScheduleReminders(t *testing.T) {
	assert := assert.New(t)

	// initialize the environment
	config, errs := config.NewConfigFromFile("./testing/.env")
	assert.Equal(0, len(errs))

	err := dynamic.Init(config)
	assert.Nil(err)

	m := dynamic.DynamicOutreachMessage{}

	// Run the update function
	err = dynamic.NotionScheduleRemindersUpdate(config, &m)
	assert.Nil(err)

	// Run the content function
	output := dynamic.NotionScheduleReminders(config, &m, time.Now())
	assert.NotNil(output)

	log.Println("Received output: ")
	log.Printf("%#v\n", output)
}
