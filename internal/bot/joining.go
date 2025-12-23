package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/engine"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
	"golang.org/x/sync/errgroup"
)

type openGame struct {
	gameId string
	width  int
	height int
	joinAt time.Time
}

type JoiningBot struct {
	botId             string
	calculateNextMove engine.CalculateNextMove
	games             sync.Map
	joinAfter         time.Duration
	sseClient         *sse.Client
	chatService       *chat.ChatService
	gameService       *connectfour.GameService
}

func NewJoiningBot(
	ctx context.Context,
	botId string,
	calculateNextMove engine.CalculateNextMove,
	joinAfter time.Duration,
	client *sse.Client,
	chatSvc *chat.ChatService,
	gameSvc *connectfour.GameService,
) (Bot, error) {
	openGamesResp, err := gameSvc.GetOpenGames(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("joining bot: could not fetch open games: %w", err)
	}

	games := sync.Map{}
	for _, game := range openGamesResp.Games {
		if game.PlayerId == botId {
			continue
		}

		games.Store(
			game.GameId,
			openGame{
				joinAt: time.Now().Add(joinAfter),
				width:  int(game.Width),
				height: int(game.Height),
			},
		)
	}

	return &JoiningBot{
		botId:             botId,
		calculateNextMove: calculateNextMove,
		joinAfter:         joinAfter,
		games:             games,
		sseClient:         client,
		chatService:       chatSvc,
		gameService:       gameSvc,
	}, nil
}

func (b *JoiningBot) Play(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return b.joinGames(ctx, eg) })
	eg.Go(func() error { return b.watchLobby(ctx) })

	return eg.Wait()
}

func (b *JoiningBot) joinGames(ctx context.Context, eg *errgroup.Group) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			closestJoinAt := time.Now().Add(b.joinAfter)
			joiningEg, _ := errgroup.WithContext(ctx)
			b.games.Range(func(key, value any) bool {
				gameId := key.(string)
				game := value.(openGame)

				if time.Now().Before(game.joinAt) {
					if game.joinAt.Before(closestJoinAt) {
						closestJoinAt = game.joinAt
					}
					return true
				}

				joiningEg.Go(func() error {
					defer b.games.Delete(gameId) // We don't retry, it's best-effort.

					errorResp, err := b.gameService.JoinGame(ctx, gameId, b.botId)
					if err != nil {
						return fmt.Errorf("joining bot: could not join game %s: %w", gameId, err)
					} else if errorResp != nil {
						return nil // Could not join, likely somebody else did in the meantime.
					}

					eg.Go(func() error {
						return playThrough(
							ctx,
							b.sseClient,
							b.gameService,
							b.chatService,
							b.calculateNextMove,
							b.botId,
							connectfour.NewGame(gameId, "", "", game.width, game.height),
						)
					})

					return nil
				})

				return true
			})

			if err := joiningEg.Wait(); err != nil {
				return err
			}

			time.Sleep(time.Until(closestJoinAt))
		}
	}
}

func (b *JoiningBot) watchLobby(ctx context.Context) error {
	sseCtx, sseCancel := context.WithCancel(ctx)
	defer sseCancel()
	resCh, err := b.sseClient.Connect(sseCtx, "lobby")
	if err != nil {
		return err
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
			case sse.GameOpened:
				if e.PlayerId == b.botId {
					continue
				}

				b.games.Store(
					e.GameId,
					openGame{joinAt: time.Now().Add(b.joinAfter), width: e.Width, height: e.Height},
				)
			case sse.GameAborted:
				b.games.Delete(e.GameId)
			case sse.PlayerJoined:
				b.games.Delete(e.GameId)
			}
		}
	}
}
