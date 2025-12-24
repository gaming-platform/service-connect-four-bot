package engine_marein

import (
	"testing"

	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
)

var iterations = 1000

func TestWinIfPossible(t *testing.T) {
	// 0 0 0 0 0 0 0
	// 0 0 0 0 0 0 0
	// X 0 0 0 0 0 0
	// 1 0 0 0 0 0 0
	// 1 0 0 0 0 0 0
	// 1 2 2 2 0 0 0
	game := newGameFromColumns([]int{1, 2, 1, 3, 1, 4})

	for i := 0; i < iterations; i++ {
		x, ok := calculateNextMove(game, 1)

		if x != 1 || !ok {
			t.Fatalf("Expected move to be 1, got %d", x)
		}
	}
}

func TestPreventWin(t *testing.T) {
	// 0 0 0 0 0 0 0
	// 1 0 0 0 0 0 0
	// 2 0 0 0 0 0 0
	// 1 0 0 0 0 0 0
	// 1 0 0 0 0 0 0
	// 1 2 2 2 X 0 0
	game := newGameFromColumns([]int{1, 2, 1, 3, 1, 1, 1, 4})

	for i := 0; i < iterations; i++ {
		x, ok := calculateNextMove(game, 1)

		if x != 5 || !ok {
			t.Fatalf("Expected move to be 5, got %d", x)
		}
	}
}

func TestThreatAwareness(t *testing.T) {
	// 0 0 0 0 0 0 0
	// 0 0 0 0 0 0 0
	// 0 0 1 0 0 0 0
	// 0 0 1 0 0 0 0
	// 0 0 2 2 2 0 0
	// 0 X 1 1 2 X 0
	game := newGameFromColumns([]int{4, 4, 3, 3, 3, 5, 3, 5})

	for i := 0; i < iterations; i++ {
		x, ok := calculateNextMove(game, 1)

		if x == 2 || x == 6 || !ok {
			t.Fatalf("Expected move to be anything except 2 or 6, got %d", x)
		}
	}
}

func TestOnlyLoosingMoves(t *testing.T) {
	// 0 1 2 2 2 1 0
	// 0 2 1 2 2 2 0
	// 0 2 2 2 1 1 0
	// 0 1 1 1 2 1 0
	// 0 2 1 2 1 1 1
	// 2 2 1 1 1 2 1
	game := newGameFromColumns([]int{4, 2, 3, 4, 4, 4, 3, 4, 5, 6, 5, 4, 3, 3, 6, 5, 5, 5, 6, 5, 6, 6, 3, 3, 6, 2, 2, 1, 7, 2, 7, 2, 2})

	for i := 0; i < iterations; i++ {
		x, ok := calculateNextMove(game, 1)

		if x != 1 || !ok {
			t.Fatalf("Expected move to be between 1, got %d", x)
		}
	}
}

func TestPreventFork1(t *testing.T) {
	// 0 X 1 1 0 2 0
	game := newGameFromColumns([]int{4, 6, 3})

	for i := 0; i < iterations; i++ {
		x, ok := calculateNextMove(game, 1)

		if x != 2 || !ok {
			t.Fatalf("Expected move to be 2, got %d", x)
		}
	}
}

func TestPreventFork2(t *testing.T) {
	// 0 0 0 0 0 0 0
	// 0 0 0 0 0 0 0
	// 0 X 2 0 0 2 1
	// 0 1 1 0 0 1 2
	// 0 1 1 1 0 2 2
	// 0 2 1 1 0 2 2
	game := newGameFromColumns([]int{4, 6, 3, 2, 2, 6, 6, 7, 3, 7, 4, 7, 7, 6, 3, 3, 2})

	for i := 0; i < iterations; i++ {
		x, ok := calculateNextMove(game, 1)

		if x != 2 || !ok {
			t.Fatalf("Expected move to be 2, got %d", x)
		}
	}
}

func TestPreventFork3(t *testing.T) {
	// 0 0 0 2 0 0 0
	// 0 0 0 1 0 0 0
	// 0 0 0 2 2 0 0
	// 0 0 X 1 1 Y 0  <- Y is how to tackle this fork, because X leads to immediate loss.
	// 0 0 1 2 1 1 0
	// 0 0 2 1 2 1 2
	game := newGameFromColumns([]int{4, 4, 4, 4, 4, 4, 6, 5, 5, 3, 3, 7, 5, 5, 6})

	for i := 0; i < iterations; i++ {
		x, ok := calculateNextMove(game, 1)

		if x != 6 || !ok {
			t.Fatalf("Expected move to be 6, got %d", x)
		}
	}
}

// todo: solution not implemented yet.
//func TestPreventFork4(t *testing.T) {
//	// 0 0 0 2 0 0 0
//	// 0 0 0 2 0 0 0
//	// 0 0 0 1 0 0 0
//	// 0 0 1 2 1 0 0
//	// 0 0 1 2 2 0 0
//	// 1 0 2 1 1 0 X
//	game := newGameFromColumns([]int{4, 4, 5, 3, 3, 4, 4, 4, 3, 4, 1, 5, 5})
//
//	for i := 0; i < iterations; i++ {
//		x, ok := calculateNextMove(game, 1)
//
//		if x != 7 || !ok {
//			t.Fatalf("Expected move to be 7, got %d", x)
//		}
//	}
//}

func TestDoNotCreateForkIfItLeadsToLoss(t *testing.T) {
	// 0 0 1 2 0 0 0
	// 0 0 2 1 1 0 0
	// 0 0 1 1 1 0 0
	// 0 X 2 1 2 0 0
	// 0 2 1 2 1 0 0
	// 0 2 2 1 2 0 0
	game := newGameFromColumns([]int{4, 4, 4, 3, 4, 5, 4, 4, 5, 5, 5, 6, 3, 2, 5, 3, 3, 3, 3, 2})

	for i := 0; i < iterations; i++ {
		x, ok := calculateNextMove(game, 1)

		if (x != 1 && x != 5 && x != 6 && x != 7) || !ok {
			t.Fatalf("Expected move to be 1 or 5 or 6 or 7, got %d", x)
		}
	}
}

// todo: solution not implemented yet.
//func TestForcingMoveLeadsToFork(t *testing.T) {
//	// 0 0 0 1 2 0 0
//	// 0 0 0 1 1 0 0
//	// 0 0 0 1 1 2 1
//	// 0 0 0 2 2 1 2
//	// 0 0 1 1 2 2 2
//	// 0 1 1 1 2 2 2
//	game := newGameFromColumns([]int{4, 5, 4, 5, 3, 5, 5, 4, 4, 6, 3, 6, 6, 7, 4, 7, 5, 7, 7, 6, 4, 5})
//
//	for i := 0; i < iterations; i++ {
//		x, ok := calculateNextMove(game, 1)
//
//		if x != 2 || !ok {
//			t.Fatalf("Expected move to be 2, got %d", x)
//		}
//	}
//}

func newGameFromColumns(columns []int) *connectfour.Game {
	game := connectfour.NewGame("", "", "", 7, 6)
	for _, x := range columns {
		y, _ := game.NextFreeRow(x)
		game.ApplyMove(x, y)
	}

	return game
}
