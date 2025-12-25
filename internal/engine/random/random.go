package engine_random

import (
	"math/rand"

	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
)

func CalculateNextMove(game *connectfour.Game) (int, bool) {
	availableColumns := game.GetAvailableColumns()
	if len(availableColumns) == 0 {
		return 0, false
	}

	return availableColumns[rand.Intn(len(availableColumns))], true
}
