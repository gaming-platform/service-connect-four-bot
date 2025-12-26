package bot

import (
	"context"

	connectfourv1 "github.com/gaming-platform/api/go/connectfour/v1"
	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/engine"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
)

type OpeningBot struct {
	botId             string
	calculateNextMove engine.CalculateNextMove
	sseClient         *sse.Client
	chatService       *chat.ChatService
	gameService       *connectfour.GameService
}

func NewOpeningBot(
	botId string,
	calculateNextMove engine.CalculateNextMove,
	client *sse.Client,
	chatSvc *chat.ChatService,
	gameSvc *connectfour.GameService,
) Bot {
	return &OpeningBot{
		botId:             botId,
		calculateNextMove: calculateNextMove,
		sseClient:         client,
		chatService:       chatSvc,
		gameService:       gameSvc,
	}
}

func (b *OpeningBot) Play(ctx context.Context) error {
	for {
		width := 7
		height := 6
		gameId, err := b.getGameIdForPlay(ctx, width, height)
		if err != nil {
			return err
		}

		if err := playThrough(
			ctx, b.sseClient,
			b.gameService,
			b.chatService,
			b.calculateNextMove,
			b.botId,
			connectfour.NewGame(gameId, "", "", width, height),
		); err != nil {
			return err
		}
	}
}

func (b *OpeningBot) getGameIdForPlay(ctx context.Context, width int, height int) (string, error) {
	openGames, err := b.gameService.GetGamesByPlayer(
		ctx,
		b.botId,
		connectfourv1.GetGamesByPlayer_STATE_OPEN,
		1,
		1,
	)
	if err != nil {
		return "", err
	} else if len(openGames.Games) > 0 {
		return openGames.Games[0].GameId, nil
	}

	return b.gameService.OpenGame(ctx, b.botId, int32(width), int32(height), -1, "move:15000")
}
