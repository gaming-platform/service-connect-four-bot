package engine_random

import (
	"testing"

	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
)

var iterationsPerCase = 1000

func TestIgnoreFullColumns(t *testing.T) {
	for i := 0; i < iterationsPerCase; i++ {
		game := connectfour.NewGame("", "", "", 2, 1)
		game.ApplyMove(1, 1)
		x, ok := CalculateNextMove(game)

		if x == 1 || !ok {
			t.Fatalf("Unexpected move %d, %v", x, ok)
		}
	}
}

func TestFull(t *testing.T) {
	for i := 0; i < iterationsPerCase; i++ {
		game := connectfour.NewGame("", "", "", 1, 1)
		game.ApplyMove(1, 1)
		x, ok := CalculateNextMove(game)

		if x != 0 && ok {
			t.Fatalf("Unexpected move %d, %v", x, ok)
		}
	}
}
