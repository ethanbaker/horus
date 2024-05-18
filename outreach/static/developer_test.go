package static_test

import (
	"log"
	"testing"

	"github.com/ethanbaker/horus/outreach/static"
	"github.com/joho/godotenv"
)

/* ---- MESSAGE TESTS ---- */

func TestPing(t *testing.T) {
	// Initalize the environment
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}

	// Initialize the module
	if err := static.Init(); err != nil {
		log.Fatal(err)
	}

	// Run the test
	log.Println(static.Ping())
}
