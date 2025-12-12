package bot

import (
	"context"
	"math/rand"

	connectfourv1 "github.com/gaming-platform/api/go/connectfour/v1"
	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
)

type OpeningBot struct {
	botId       string
	sseClient   *sse.Client
	chatService *chat.ChatService
	gameService *connectfour.GameService
}

func NewOpeningBot(
	botId string,
	client *sse.Client,
	chatSvc *chat.ChatService,
	gameSvc *connectfour.GameService,
) *OpeningBot {
	return &OpeningBot{botId: botId, sseClient: client, chatService: chatSvc, gameService: gameSvc}
}

func (b *OpeningBot) Play(ctx context.Context) error {
	var chatId string
	width := 7
	gameId, err := b.getGameIdForPlay(ctx, width)
	if err != nil {
		return err
	}

	sseCtx, sseCancel := context.WithCancel(ctx)
	defer sseCancel()
	resCh, err := b.sseClient.Connect(sseCtx, "connect-four-"+gameId)

	for {
		select {
		case <-sseCtx.Done():
			return nil
		case res := <-resCh:
			if res.Error != nil {
				return res.Error
			}

			switch res.Event.Name {
			case "ConnectFour.ChatAssigned":
				v, ok := res.Event.Payload["chatId"].(string)
				if !ok {
					continue
				}
				chatId = v

				go b.chatService.WriteMessage(
					ctx,
					chatId,
					b.botId,
					"Good luck, have fun!",
					"opening",
				)
			case "ConnectFour.PlayerJoined":
				redPlayerId, ok := res.Event.Payload["redPlayerId"].(string)
				if !ok || redPlayerId != b.botId {
					continue
				}

				if _, err := b.gameService.MakeMove(sseCtx, gameId, b.botId, int32(4)); err != nil {
					return err
				}
			case "ConnectFour.PlayerMoved":
				nextPlayerId, ok := res.Event.Payload["nextPlayerId"].(string)
				if !ok || nextPlayerId != b.botId {
					continue
				}

				for {
					c := rand.Intn(width) + 1
					errResp, err := b.gameService.MakeMove(sseCtx, gameId, b.botId, int32(c))
					if err != nil {
						return err
					} else if errResp != nil && !errResp.HasViolation("column_already_filled") {
						break
					} else if errResp == nil && err == nil {
						break
					}
				}
			case "ConnectFour.GameAborted":
				if chatId == "" {
					continue
				}
				go b.chatService.WriteMessage(
					ctx,
					chatId,
					b.botId,
					"Next time, perhaps!",
					"ending",
				)
				sseCancel()
			case "ConnectFour.GameWon",
				"ConnectFour.GameDrawn",
				"ConnectFour.GameTimedOut",
				"ConnectFour.GameResigned":
				if chatId == "" {
					continue
				}
				go b.chatService.WriteMessage(
					ctx,
					chatId,
					b.botId,
					"Good game! Well played.",
					"ending",
				)
				sseCancel()
			}
		}
	}
}

func (b *OpeningBot) getGameIdForPlay(ctx context.Context, width int) (string, error) {
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

	return b.gameService.OpenGame(ctx, b.botId, int32(width), 6, -1, "move:15000")
}
