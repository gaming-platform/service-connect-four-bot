package engine_marein

import (
	"math"
	"math/rand"

	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
)

// A custom implementation with some heuristics.
// The design goal is to create an engine that plays reasonably well without
// solving the game completely (because Connect Four is a solved game).
// Favors readability over efficiency.

type Options struct {
	ForkCreationProbability int
}

func NewOptions(
	forkCreationProbability int,
) Options {
	return Options{
		ForkCreationProbability: forkCreationProbability,
	}
}

func CreateCalculateNextMove(options Options) func(game *connectfour.Game) (int, bool) {
	return func(game *connectfour.Game) (int, bool) {
		return calculateNextMove(game, options)
	}
}

func calculateNextMove(game *connectfour.Game, options Options) (int, bool) {
	current, opponent := game.GetCurrentPlayerColors()

	availableColumns := game.GetAvailableColumns()
	if len(availableColumns) == 0 {
		return 0, false
	}

	// Win if possible.
	if x, ok := findWinningMove(game, availableColumns, current); ok {
		return x, true
	}

	// Prevent opponent from winning.
	if x, ok := findWinningMove(game, availableColumns, opponent); ok {
		return x, true
	}

	nonLoosingColumns := filterNonLoosingColumns(game, availableColumns)
	if len(nonLoosingColumns) == 0 {
		return availableColumns[0], true
	}

	// Prevent opponent from creating a fork.
	if x, ok := findForkingMove(game, nonLoosingColumns, opponent); ok {
		return x, true
	}

	// Create a fork if possible.
	if rand.Intn(100) < options.ForkCreationProbability {
		if x, ok := findForkingMove(game, nonLoosingColumns, current); ok {
			return x, true
		}
	}

	// ideas:
	// * prefer cluster moves (low weight, probably before random)
	// * check forcing moves (wins the opponent needs to prevent) and see if those are creating threats.

	return findRandomLegalMoveThatPrefersCenter(game, nonLoosingColumns), true
}

func findWinningMove(game *connectfour.Game, columns []int, color int) (int, bool) {
	for _, x := range columns {
		y, _ := game.NextFreeRow(x)
		gameClone := game.Clone()
		gameClone.ForceMove(x, y, color)

		if connectfour.IsWinningMove(gameClone, x, y, color) {
			return x, true
		}
	}

	return 0, false
}

func findForkingMove(game *connectfour.Game, columns []int, color int) (int, bool) {
	for _, x := range columns {
		y, _ := game.NextFreeRow(x)
		gameClone := game.Clone()
		gameClone.ForceMove(x, y, color)

		nextColumns := gameClone.GetAvailableColumns()

		// Test horizontally.
		if firstWinX, ok := findWinningMove(gameClone, nextColumns, color); ok {
			if _, ok := findWinningMove(gameClone, removeColumn(nextColumns, firstWinX), color); ok {
				return x, true // Second winning move shows the forking threat.
			}
		}

		// Test vertically.
		for _, verticalX := range nextColumns {
			verticalY, _ := gameClone.NextFreeRow(verticalX)
			verticalGameClone := gameClone.Clone()
			verticalGameClone.ForceMove(verticalX, verticalY, color)

			if connectfour.IsWinningMove(verticalGameClone, verticalX, verticalY, color) &&
				game.IsInBounds(verticalX, verticalY-1) &&
				!game.HasMoveAt(verticalX, verticalY-1) {
				verticalGameClone := gameClone.Clone() // Clone again to prevent checking vertical wins.
				verticalGameClone.ForceMove(verticalX, verticalY-1, color)
				if connectfour.IsWinningMove(verticalGameClone, verticalX, verticalY-1, color) {
					return x, true // Second winning move shows the forking threat.
				}
			}
		}
	}

	return 0, false
}

func findRandomLegalMoveThatPrefersCenter(game *connectfour.Game, columns []int) int {
	center := math.Ceil(float64(game.Width) / 2)
	weightedColumns := make([]int, 0)
	for _, y := range columns {
		// 4 - abs(4 - col 1) = wgt 1
		// 4 - abs(4 - col 2) = wgt 2
		// 4 - abs(4 - col 3) = wgt 3
		// 4 - abs(4 - col 4) = wgt 4
		// 4 - abs(4 - col 5) = wgt 3
		// 4 - abs(4 - col 6) = wgt 2
		// 4 - abs(4 - col 7) = wgt 1
		weight := math.Pow(5, center-math.Abs(center-float64(y)))
		for i := 0.0; i < weight; i++ {
			weightedColumns = append(weightedColumns, y)
		}
	}

	return weightedColumns[rand.Intn(len(weightedColumns))]
}

func canOpponentWin(game *connectfour.Game) bool {
	current, _ := game.GetCurrentPlayerColors()
	availableColumns := game.GetAvailableColumns()

	for _, x := range availableColumns {
		y, _ := game.NextFreeRow(x)
		gameClone := game.Clone()
		gameClone.ApplyMove(x, y)

		if connectfour.IsWinningMove(gameClone, x, y, current) {
			return true
		}
	}

	return false
}

func filterNonLoosingColumns(game *connectfour.Game, columns []int) []int {
	nonLoosingColumns := make([]int, 0)

	for _, x := range columns {
		y, _ := game.NextFreeRow(x)
		gameClone := game.Clone()
		gameClone.ApplyMove(x, y)

		if !canOpponentWin(gameClone) {
			nonLoosingColumns = append(nonLoosingColumns, x)
		}
	}

	return nonLoosingColumns
}

func removeColumn(columns []int, column int) []int {
	var filteredColumns []int
	for _, c := range columns {
		if c != column {
			filteredColumns = append(filteredColumns, c)
		}
	}
	return filteredColumns
}
