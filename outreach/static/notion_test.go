package static_test

import (
	"log"
	"testing"

	"github.com/ethanbaker/horus/outreach/static"
	"github.com/joho/godotenv"
)

/* ---- MESSAGE TESTS ---- */

func TestNotionDailyDigest(t *testing.T) {
	log.Println(static.NotionDailyDigest())
}

func TestNotionNightAffirmations(t *testing.T) {
	log.Println(static.NotionNightAffirmations())
}

func TestMain(m *testing.M) {
	// Initalize the environment
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}

	// Initialize the module
	if err := static.Init(); err != nil {
		log.Fatal(err)
	}

	m.Run()
}
