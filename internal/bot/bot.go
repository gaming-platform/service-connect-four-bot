package bot

import (
	"context"
	"math/rand"

	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
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
	botId string,
	gameId string,
	width int,
	chatId string,
	currentPlayerId string,
) error {
	runningGamesGauge.Inc()
	defer runningGamesGauge.Dec()

	sseCtx, sseCancel := context.WithCancel(ctx)
	defer sseCancel()
	resCh, err := sseClient.Connect(sseCtx, "connect-four-"+gameId)
	if err != nil {
		return err
	}

	if currentPlayerId == botId {
		if err := makeMove(botId, gameId, width, sseCtx, gameService); err != nil {
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
				chatId = e.ChatId

				go chatService.WriteMessage(
					ctx,
					chatId,
					botId,
					"Good luck, have fun!",
					"opening",
				)
			case sse.PlayerJoined:
				if e.RedPlayerId != botId {
					continue
				}

				if err := makeMove(botId, gameId, width, sseCtx, gameService); err != nil {
					return err
				}
			case sse.PlayerMoved:
				if e.NextPlayerId != botId {
					continue
				}

				if err := makeMove(botId, gameId, width, sseCtx, gameService); err != nil {
					return err
				}
			case sse.GameAborted:
				if chatId != "" {
					go chatService.WriteMessage(
						ctx,
						chatId,
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
				if chatId != "" {
					go chatService.WriteMessage(
						ctx,
						chatId,
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

func makeMove(botId, gameId string, width int, sseCtx context.Context, gameService *connectfour.GameService) error {
	for {
		c := rand.Intn(width) + 1
		errResp, err := gameService.MakeMove(sseCtx, gameId, botId, int32(c))
		if err != nil {
			return err
		} else if errResp != nil && !errResp.HasViolation("column_already_filled") {
			return nil
		} else if errResp == nil && err == nil {
			return nil
		}
	}
}
