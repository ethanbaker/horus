package dynamic_test

import (
	"log"
	"testing"
	"time"

	"github.com/ethanbaker/horus/outreach/dynamic"
	"github.com/joho/godotenv"
)

/* ---- UPDATE TESTS ---- */

func TestNotionScheduleRemindersUpdate(t *testing.T) {
	log.Println(dynamic.NotionScheduleRemindersUpdate(&dynamic.DynamicOutreachMessage{}))
}

/* ---- MESSAGE TESTS ---- */

func TestNotionScheduleReminders(t *testing.T) {
	log.Println(dynamic.NotionScheduleReminders(&dynamic.DynamicOutreachMessage{}, time.Now()))
}

/* ---- MAIN ---- */

func TestMain(m *testing.M) {
	// Initalize the environment
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}

	// Initialize the module
	if err := dynamic.Init(); err != nil {
		log.Fatal(err)
	}

	// Run the tests
	m.Run()
}
