package engine_random

import (
	"math/rand"

	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
)

func CalculateNextMove(game *connectfour.Game) (int, bool) {
	freeColumns := game.GetFreeColumns()
	if len(freeColumns) == 0 {
		return 0, false
	}

	return freeColumns[rand.Intn(len(freeColumns))], true
}
