package engine

import "github.com/gaming-platform/connect-four-bot/internal/connectfour"

type CalculateNextMove func(game *connectfour.Game) (int, bool)
