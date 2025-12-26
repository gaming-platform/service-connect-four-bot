package bot

import (
	"context"

	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/engine"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var runningGamesGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "connect_four_bot_running_games",
	Help: "The number of running games the bot is currently playing.",
})

type Bot interface {
	Play(ctx context.Context) error
}

func playThrough(
	ctx context.Context,
	sseClient *sse.Client,
	gameService *connectfour.GameService,
	chatService *chat.ChatService,
	calculateNextMove engine.CalculateNextMove,
	botId string,
	game *connectfour.Game,
) error {
	runningGamesGauge.Inc()
	defer runningGamesGauge.Dec()

	sseCtx, sseCancel := context.WithCancel(ctx)
	defer sseCancel()
	resCh, err := sseClient.Connect(sseCtx, "connect-four-"+game.GameId)
	if err != nil {
		return err
	}

	if game.CurrentPlayerId == botId {
		if err := makeMove(botId, game, sseCtx, gameService, calculateNextMove); err != nil {
			return err
		}
	}

	for {
		select {
		case <-sseCtx.Done():
			return nil
		case res := <-resCh:
			if res.Error != nil {
				return res.Error
			}

			switch e := res.Event.(type) {
			case sse.ChatAssigned:
				game.ChatId = e.ChatId

				go chatService.WriteMessage(
					ctx,
					game.ChatId,
					botId,
					"Good luck, have fun!",
					"opening",
				)
			case sse.PlayerJoined:
				if e.RedPlayerId != botId {
					continue
				}

				if err := makeMove(botId, game, sseCtx, gameService, calculateNextMove); err != nil {
					return err
				}
			case sse.PlayerMoved:
				if ok := game.ApplyMove(e.X, e.Y); !ok {
					continue
				}

				if e.NextPlayerId != botId {
					continue
				}

				if err := makeMove(botId, game, sseCtx, gameService, calculateNextMove); err != nil {
					return err
				}
			case sse.GameAborted:
				if game.ChatId != "" {
					go chatService.WriteMessage(
						ctx,
						game.ChatId,
						botId,
						"Next time, perhaps!",
						"ending",
					)
				}
				sseCancel()
			case sse.GameWon,
				sse.GameDrawn,
				sse.GameTimedOut,
				sse.GameResigned:
				if game.ChatId != "" {
					go chatService.WriteMessage(
						ctx,
						game.ChatId,
						botId,
						"Good game! Well played.",
						"ending",
					)
				}
				sseCancel()
			}
		}
	}
}

func makeMove(
	botId string,
	game *connectfour.Game,
	sseCtx context.Context,
	gameService *connectfour.GameService,
	calculateNextMove engine.CalculateNextMove,
) error {
	c, ok := calculateNextMove(game)
	if !ok {
		return nil // The engine couldn't find a valid move. The game is probably already finished.
	}

	// Ignoring errResp, probably the game is already finished if that's returned.
	_, err := gameService.MakeMove(sseCtx, game.GameId, botId, int32(c))

	return err
}
