package bot

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
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
	botId       string
	games       sync.Map
	joinAfter   time.Duration
	sseClient   *sse.Client
	chatService *chat.ChatService
	gameService *connectfour.GameService
}

func NewJoiningBot(
	botId string,
	joinAfter time.Duration,
	client *sse.Client,
	chatSvc *chat.ChatService,
	gameSvc *connectfour.GameService,
) Bot {
	return &JoiningBot{
		botId:       botId,
		joinAfter:   joinAfter,
		games:       sync.Map{}, // todo: Fetch currently open games, but this is not exposed yet.
		sseClient:   client,
		chatService: chatSvc,
		gameService: gameSvc,
	}
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
							b.botId,
							gameId,
							game.width,
							"",
							"",
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

			switch res.Event.Name {
			case "ConnectFour.GameOpened":
				playerId, ok := res.Event.Payload["playerId"].(string)
				if !ok {
					return errors.New("joining bot: opened: could not extract playerId")
				} else if playerId == b.botId {
					continue
				}

				gameId, ok := res.Event.Payload["gameId"].(string)
				if !ok {
					return errors.New("joining bot: opened: could not extract gameId")
				}

				width, ok := res.Event.Payload["width"].(float64)
				if !ok {
					return errors.New("joining bot: opened: could not extract width")
				}

				height, ok := res.Event.Payload["height"].(float64)
				if !ok {
					return errors.New("joining bot: opened: could not extract height")
				}

				b.games.Store(
					gameId,
					openGame{joinAt: time.Now().Add(b.joinAfter), width: int(width), height: int(height)},
				)
			case "ConnectFour.PlayerJoined":
			case "ConnectFour.GameAborted":
				gameId, ok := res.Event.Payload["gameId"].(string)
				if !ok {
					return errors.New("joining bot: closed: could not extract gameId")
				}

				b.games.Delete(gameId)
			}
		}
	}
}
