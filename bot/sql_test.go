package horus_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	horus "github.com/ethanbaker/horus/bot"
	"github.com/stretchr/testify/assert"
)

/* ---- SUITE TESTING ---- */

func (s *Suite) TestGetAllBots() {
	assert := assert.New(s.T())

	// TEST SQL OUTLINE
	// - Preload all of the tables

	s.mock.ExpectQuery("^SELECT (.+) FROM `bots`").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at", "deleted_at", "name", "permissions"}))

	// Getting all bots should result in an empty list
	bots, err := horus.GetAllBots(s.config)
	assert.Nil(err)
	assert.Empty(bots)
}

func (s *Suite) TestGetBotByName() {
	assert := assert.New(s.T())

	// TEST SQL OUTLINE
	// - Preload all of the tables

	s.mock.ExpectQuery("^SELECT (.+) FROM `bots`").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at", "deleted_at", "name", "permissions"}))

	// Getting a bot by name should result in a nil pointer
	bot, err := horus.GetBotByName("invalid", s.config)
	assert.Nil(err)
	assert.Nil(bot)

}
