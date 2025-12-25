package connectfour

import "strconv"

type Game struct {
	GameId                string
	ChatId                string
	CurrentPlayerId       string
	WinningSequenceLength int
	Width                 int
	Height                int
	moves                 map[string]Move
}

type Move struct {
	X     int
	Y     int
	Color int // 1 (red) or 2 (yellow)
}

func NewGame(
	gameId string,
	chatId string,
	currentPlayerId string,
	width int,
	height int,
) *Game {
	return &Game{
		GameId:                gameId,
		ChatId:                chatId,
		CurrentPlayerId:       currentPlayerId,
		WinningSequenceLength: 4, // Although technically possible, players cannot customize this yet.
		Width:                 width,
		Height:                height,
		moves:                 make(map[string]Move),
	}
}

// ApplyMove ensures idempotency.
func (g *Game) ApplyMove(x int, y int) bool {
	_, ok := g.moves[g.moveKey(x, y)]
	if ok {
		return false
	}

	color, _ := g.GetCurrentPlayerColors()
	g.moves[g.moveKey(x, y)] = Move{X: x, Y: y, Color: color}

	return true
}

func (g *Game) GetMoveAt(x int, y int) (Move, bool) {
	mv, ok := g.moves[g.moveKey(x, y)]
	return mv, ok
}

func (g *Game) HasMoveAt(x int, y int) bool {
	_, ok := g.moves[g.moveKey(x, y)]
	return ok
}

func (g *Game) ForceMove(x int, y int, color int) {
	g.moves[g.moveKey(x, y)] = Move{X: x, Y: y, Color: color}
}

func (g *Game) IsInBounds(x int, y int) bool {
	return x >= 1 && x <= g.Width && y >= 1 && y <= g.Height
}

// GetCurrentPlayerColors returns the first int as the current player's color, second as the opponent's player color.
func (g *Game) GetCurrentPlayerColors() (int, int) {
	if len(g.moves)%2 == 0 {
		return 1, 2
	}
	return 2, 1
}

func (g *Game) GetAvailableColumns() []int {
	availableColumns := make([]int, 0)

	for x := 1; x <= g.Width; x++ {
		if _, ok := g.NextFreeRow(x); ok {
			availableColumns = append(availableColumns, x)
		}
	}

	return availableColumns
}

func (g *Game) NextFreeRow(x int) (int, bool) {
	for y := g.Height; y >= 1; y-- {
		if _, ok := g.GetMoveAt(x, y); !ok {
			return y, true
		}
	}

	return 0, false
}

func (g *Game) Clone() *Game {
	moves := make(map[string]Move)
	for k, v := range g.moves {
		moves[k] = v
	}

	return &Game{
		GameId:                g.GameId,
		ChatId:                g.ChatId,
		CurrentPlayerId:       g.CurrentPlayerId,
		WinningSequenceLength: g.WinningSequenceLength,
		Width:                 g.Width,
		Height:                g.Height,
		moves:                 moves,
	}
}

func (g *Game) moveKey(x int, y int) string {
	return strconv.Itoa(x) + "," + strconv.Itoa(y)
}
