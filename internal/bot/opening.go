package bot

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
)

type OpeningBot struct {
	botId       string
	sseClient   *sse.Client
	gameService *connectfour.GameService
}

func NewOpeningBot(botId string, client *sse.Client, cfSvc *connectfour.GameService) *OpeningBot {
	return &OpeningBot{
		botId:       botId,
		sseClient:   client,
		gameService: cfSvc,
	}
}

func (b *OpeningBot) Play(ctx context.Context) error {
	width := int32(7)
	gameId, errResp, err := b.gameService.OpenGame(ctx, b.botId, width, 6, -1, "move:15000")
	if err != nil {
		return err
	} else if errResp != nil {
		return fmt.Errorf("failed to open game: %v", errResp)
	}

	sseCtx, sseCancel := context.WithCancel(ctx)
	defer sseCancel()
	resCh, err := b.sseClient.Connect(sseCtx, "connect-four-"+gameId)

	for {
		select {
		case res := <-resCh:
			if sseCtx.Err() != nil {
				return nil
			} else if res.Error != nil {
				return res.Error
			}

			switch res.Event.Name {
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
					c := rand.Intn(int(width)) + 1
					errResp, err := b.gameService.MakeMove(sseCtx, gameId, b.botId, int32(c))
					if err != nil {
						return err
					} else if errResp != nil && !errResp.HasViolation("column_already_filled") {
						break
					}
				}
			case "ConnectFour.GameAborted",
				"ConnectFour.GameWon",
				"ConnectFour.GameDrawn",
				"ConnectFour.GameTimedOut",
				"ConnectFour.GameResigned":
				sseCancel()
			}
		}
	}
}
