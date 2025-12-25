package engine_marein

import (
	"strconv"
	"strings"
	"testing"

	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
)

var iterationsPerCase = 1000

type boardCase struct {
	board   string
	allowed []int
}

func TestBoardCases(t *testing.T) {
	boardCases := map[string]boardCase{
		"WinIfPossible": {
			board: `0 0 0 0 0 0 0
					0 0 0 0 0 0 0
					0 0 0 0 0 0 0
					1 0 0 0 0 0 0
					1 0 0 0 0 0 0
					1 2 2 2 0 0 0`,
			allowed: []int{1},
		},
		"PreventWin": {
			board: `0 0 0 0 0 0 0
					1 0 0 0 0 0 0
					2 0 0 0 0 0 0
					1 0 0 0 0 0 0
					1 0 0 0 0 0 0
					1 2 2 2 0 0 0`,
			allowed: []int{5},
		},
		"IgnoreLosingColumns": {
			board: `0 0 0 0 0 0 0
					0 0 0 0 0 0 0
					0 0 1 0 0 0 0
					0 0 1 0 0 0 0
					0 0 2 2 2 0 0
					0 0 1 1 2 0 0`,
			allowed: []int{1, 3, 4, 5, 7},
		},
		"UseFirstWhenOnlyOptionIsLosing": {
			board: `0 1 2 2 2 1 0
					0 2 1 2 2 2 0
					0 2 2 2 1 1 0
					0 1 1 1 2 1 0
					0 2 1 2 1 1 1
					2 2 1 1 1 2 1`,
			allowed: []int{1},
		},
		"CreateHorizontalFork1": {
			board:   `0 X 1 1 0 2 2`,
			allowed: []int{2},
		},
		"PreventHorizontalFork1": {
			board:   `0 X 1 1 0 2 0`,
			allowed: []int{2},
		},
		"CreateHorizontalFork2": {
			board: `0 0 0 0 0 0 0
					0 0 0 0 0 0 2
					0 X 2 0 0 0 1
					0 1 1 0 0 0 2
					0 1 1 1 0 2 2
					0 2 1 1 0 2 2`,
			allowed: []int{2},
		},
		"PreventHorizontalFork2": {
			board: `0 0 0 0 0 0 0
					0 0 0 0 0 0 0
					0 X 2 0 0 2 1
					0 1 1 0 0 1 2
					0 1 1 1 0 2 2
					0 2 1 1 0 2 2`,
			allowed: []int{2},
		},
		"CreateHorizontalFork3": {
			board: `0 0 0 0 0 0 0
					0 0 0 1 0 0 0
					0 0 0 2 2 0 0
					0 0 X 1 1 0 0
					0 0 1 2 1 1 0
					0 0 2 1 2 1 2`,
			allowed: []int{3},
		},
		"PreventHorizontalFork3": {
			// Y would prevent the fork, but would lose immediately.
			board: `0 0 0 2 0 0 0
					0 0 0 1 0 0 0
					0 0 0 2 2 0 0
					0 0 Y 1 1 X 0
					0 0 1 2 1 1 0
					0 0 2 1 2 1 2`,
			allowed: []int{6},
		},
		"CreateVerticalFork1": {
			board: `0 0 0 2 0 0 0
					0 0 0 2 0 0 0
					0 0 0 1 0 0 0
					0 0 1 2 1 0 0
					2 0 1 2 2 0 0
					1 0 2 1 1 0 X`,
			allowed: []int{7},
		},
		"PreventVerticalFork1": {
			board: `0 0 0 2 0 0 0
					0 0 0 2 0 0 0
					0 0 0 1 0 0 0
					0 0 1 2 1 0 0
					0 0 1 2 2 0 0
					1 0 2 1 1 0 X`,
			allowed: []int{7},
		},
		"CreateVerticalFork2": {
			board: `0 0 0 1 2 0 0
					0 0 1 2 2 0 0
					0 0 2 2 2 0 0
					0 0 1 2 1 0 0
					0 X 1 1 2 1 0
					2 1 1 1 2 1 0`,
			allowed: []int{2},
		},
		"CreateVerticalFork3": {
			// Y would be a fork, but would lose immediately.
			board: `0 0 0 2 0 0 0
					0 0 2 2 2 0 0
					0 0 2 2 2 0 0
					0 Y 1 2 1 0 0
					X 1 2 1 1 0 0
					1 1 2 1 1 0 0`,
			allowed: []int{1},
		},
		"DoNotCreateLosingForks1": {
			board: `0 0 1 2 0 0 0
					0 0 2 1 1 0 0
					0 0 1 1 1 0 0
					0 X 2 1 2 0 0
					0 2 1 2 1 0 0
					0 2 2 1 2 0 0`,
			allowed: []int{1, 5, 6, 7},
		},
		"DoNotCreateLosingForks2": {
			board: `0 0 0 1 0 0 0
					0 0 0 1 0 0 0
					0 0 0 1 2 0 0
					0 2 X 2 1 0 0
					0 1 2 1 2 1 0
					0 1 2 1 2 2 2`,
			allowed: []int{1, 2, 5, 6, 7},
		},
		//"ForcingMoveLeadsToFork1": {
		//	// X forces Y, then Z can create a fork.
		//	board: `0 0 0 1 2 0 0
		//			0 0 0 1 1 0 0
		//			0 0 0 1 1 2 1
		//			0 0 0 2 2 1 2
		//			0 Z 1 1 2 2 2
		//			Y X 1 1 2 2 2`,
		//	allowed: []int{2},
		//},
		//"ForcingMoveLeadsToFork2": {
		//	// X forces Y, then Z creates a future fork.
		//	// A, B, C, D shows an example continuation.
		//	board: `0 0 2 1 2 0 0
		//			0 0 1 1 2 A D
		//			0 0 2 1 1 Z C
		//			0 0 2 2 1 X B
		//			0 0 1 1 1 2 Y
		//			2 0 2 1 2 1 2`,
		//	allowed: []int{2},
		//},
	}

	for name, c := range boardCases {
		t.Run(name, func(t *testing.T) { runBoardCase(t, c) })
	}
}

func runBoardCase(t *testing.T, c boardCase) {
	t.Helper()
	for i := 0; i < iterationsPerCase; i++ {
		game := newGameFromAscii(c.board)
		x, ok := calculateNextMove(game, NewOptions(100))

		found := false
		for _, allowedX := range c.allowed {
			if x == allowedX && ok {
				found = true
			}
		}

		if !found {
			t.Fatalf("iteration %d returned %d, %v; allowed one of %v", i+1, x, ok, c.allowed)
		}
	}
}

func newGameFromAscii(board string) *connectfour.Game {
	lines := make([]string, 0)
	for _, l := range strings.Split(strings.TrimSpace(board), "\n") {
		lines = append(lines, strings.TrimSpace(l))
	}

	fields := strings.Fields(lines[0])

	game := connectfour.NewGame("", "", "", len(fields), len(lines))

	for y, line := range lines {
		fields := strings.Fields(line)
		for x, color := range fields {
			color, err := strconv.Atoi(color)
			if err != nil || color == 0 {
				continue
			}

			game.ForceMove(x+1, y+1, color)
		}
	}

	return game
}
