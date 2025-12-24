package engine_marein

import (
	"math"
	"math/rand"

	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
)

// A custom implementation with some heuristics.
// No big algorithms, favors readability over efficiency, just for fun.
// It currently ignores depth > 1.

func CreateCalculateNextMove(maxDepth int) func(game *connectfour.Game) (int, bool) {
	return func(game *connectfour.Game) (int, bool) {
		return calculateNextMove(game, maxDepth)
	}
}

func calculateNextMove(game *connectfour.Game, maxDepth int) (int, bool) {
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
	if x, ok := findForkingMove(game, nonLoosingColumns, current); ok {
		return x, true
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

// findForkingMove looks for positions that will result in: 2 0 1 1 1 0 2.
func findForkingMove(game *connectfour.Game, columns []int, color int) (int, bool) {
	for _, x := range columns {
		y, _ := game.NextFreeRow(x)
		gameClone := game.Clone()
		gameClone.ForceMove(x, y, color)

		nextColumns := gameClone.GetAvailableColumns()
		if firstWinX, ok := findWinningMove(gameClone, nextColumns, color); ok {
			nextColumns = removeColumn(nextColumns, firstWinX)
			if _, ok := findWinningMove(gameClone, nextColumns, color); ok {
				return x, true // Second winning move shows the forking threat.
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

func filterNonLoosingColumns(game *connectfour.Game, columns []int) []int {
	nonLoosingColumns := make([]int, 0)

	for _, x := range columns {
		y, _ := game.NextFreeRow(x)
		gameClone := game.Clone()
		gameClone.ApplyMove(x, y)

		if !findThreat(gameClone, 1) {
			nonLoosingColumns = append(nonLoosingColumns, x)
		}
	}

	return nonLoosingColumns
}

func findThreat(game *connectfour.Game, depth int) bool {
	if depth <= 0 {
		return false
	}

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

func removeColumn(columns []int, column int) []int {
	var filteredColumns []int
	for _, c := range columns {
		if c != column {
			filteredColumns = append(filteredColumns, c)
		}
	}
	return filteredColumns
}
