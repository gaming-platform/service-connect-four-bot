package bot

import (
	"context"

	connectfourv1 "github.com/gaming-platform/api/go/connectfour/v1"
	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
	"golang.org/x/sync/errgroup"
)

type ResumingBot struct {
	botId       string
	games       []*connectfourv1.Game
	sseClient   *sse.Client
	chatService *chat.ChatService
	gameService *connectfour.GameService
}

func NewResumingBot(
	ctx context.Context,
	botId string,
	client *sse.Client,
	chatSvc *chat.ChatService,
	gameSvc *connectfour.GameService,
) (*ResumingBot, error) {
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

	return &ResumingBot{botId: botId, games: games, sseClient: client, chatService: chatSvc, gameService: gameSvc}, nil
}

func (b *ResumingBot) Play(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, game := range b.games {
		eg.Go(func() error {
			return playThrough(
				ctx,
				b.sseClient,
				b.gameService,
				b.chatService,
				b.botId,
				game.GameId,
				7,  // Hard coded width, because it's not exposed yet.
				"", // No chat id, because it's not exposed yet.
				game.CurrentPlayerId,
			)
		})
	}

	return eg.Wait()
}
