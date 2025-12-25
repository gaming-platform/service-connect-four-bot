package connectfour

func IsWinningMove(g *Game, x, y, color int) bool {
	directions := [][2]int{
		{1, 0},  // horizontal
		{0, 1},  // vertical
		{1, 1},  // diagonal \
		{1, -1}, // diagonal /
	}

	for _, d := range directions {
		count := 1
		count += countDirection(g, x, y, d[0], d[1], color)
		count += countDirection(g, x, y, -d[0], -d[1], color)

		if count >= g.WinningSequenceLength {
			return true
		}
	}

	return false
}

func countDirection(g *Game, x, y, dx, dy, color int) int {
	count := 0
	for i := 1; ; i++ {
		nx := x + i*dx
		ny := y + i*dy

		if nx < 1 || nx > g.Width || ny < 1 || ny > g.Height {
			break
		}

		move, ok := g.GetMoveAt(nx, ny)
		if !ok || move.Color != color {
			break
		}

		count++
	}
	return count
}
