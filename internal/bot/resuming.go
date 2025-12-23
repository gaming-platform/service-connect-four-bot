package bot

import (
	"context"

	connectfourv1 "github.com/gaming-platform/api/go/connectfour/v1"
	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/engine"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
	"golang.org/x/sync/errgroup"
)

type ResumingBot struct {
	botId             string
	calculateNextMove engine.CalculateNextMove
	games             []*connectfourv1.Game
	sseClient         *sse.Client
	chatService       *chat.ChatService
	gameService       *connectfour.GameService
}

func NewResumingBot(
	ctx context.Context,
	botId string,
	calculateNextMove engine.CalculateNextMove,
	client *sse.Client,
	chatSvc *chat.ChatService,
	gameSvc *connectfour.GameService,
) (Bot, error) {
	gamesPerPage := 100
	currentPage := 0
	var games []*connectfourv1.Game
	for {
		currentPage++
		runningGames, err := gameSvc.GetGamesByPlayer(
			ctx,
			botId,
			connectfourv1.GetGamesByPlayer_STATE_RUNNING,
			int32(currentPage),
			int32(gamesPerPage),
		)
		if err != nil {
			return nil, err
		}

		for _, game := range runningGames.Games {
			games = append(games, game)
		}

		if len(runningGames.Games) == 0 {
			break
		}
	}

	return &ResumingBot{
		botId:             botId,
		calculateNextMove: calculateNextMove,
		games:             games,
		sseClient:         client,
		chatService:       chatSvc,
		gameService:       gameSvc,
	}, nil
}

func (b *ResumingBot) Play(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, game := range b.games {
		eg.Go(func() error {
			gameModel := connectfour.NewGame(
				game.GameId,
				game.ChatId,
				game.CurrentPlayerId,
				int(game.Width),
				int(game.Height),
			)

			for _, move := range game.Moves {
				gameModel.ForceMove(int(move.X), int(move.Y), int(move.Color))
			}

			return playThrough(
				ctx,
				b.sseClient,
				b.gameService,
				b.chatService,
				b.calculateNextMove,
				b.botId,
				gameModel,
			)
		})
	}

	return eg.Wait()
}
