package connectfour

import (
	"context"
	"errors"

	connectfourv1 "github.com/gaming-platform/api/go/connectfour/v1"
	"github.com/gaming-platform/connect-four-bot/internal/api"
	"github.com/gaming-platform/connect-four-bot/internal/rpcclient"
	"google.golang.org/protobuf/proto"
)

type GameService struct {
	rpcClient rpcclient.RpcClient
}

func NewGameServiceService(rpcCl rpcclient.RpcClient) *GameService {
	return &GameService{
		rpcClient: rpcCl,
	}
}

func (s *GameService) OpenGame(
	ctx context.Context,
	playerId string,
	width int32,
	height int32,
	stone int32,
	timer string,
) (string, *api.ErrorResponse, error) {
	req := connectfourv1.OpenGame{
		PlayerId: playerId,
		Width:    width,
		Height:   height,
		Stone:    connectfourv1.OpenGame_Stone(stone),
		Timer:    timer,
	}
	reqBody, err := proto.Marshal(&req)
	if err != nil {
		return "", nil, err
	}

	resp, err := s.rpcClient.Call(ctx, rpcclient.Message{Name: "ConnectFour.OpenGame", Body: reqBody})
	if err != nil {
		return "", nil, err
	}

	switch resp.Name {
	case "ConnectFour.OpenGameResponse":
		var openGameResp connectfourv1.OpenGameResponse
		err = proto.Unmarshal(resp.Body, &openGameResp)
		if err != nil {
			return "", nil, err
		}

		return openGameResp.GameId, nil, nil
	case "Common.ErrorResponse":
		errResp, err := api.NewErrorResponse(resp.Body)
		if err != nil {
			return "", nil, err
		}

		return "", errResp, nil
	default:
		return "", nil, errors.New("unknown response")
	}
}

func (s *GameService) MakeMove(
	ctx context.Context,
	gameId string,
	playerId string,
	column int32,
) (*api.ErrorResponse, error) {
	req := connectfourv1.MakeMove{GameId: gameId, PlayerId: playerId, Column: column}
	reqBody, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}

	resp, err := s.rpcClient.Call(ctx, rpcclient.Message{Name: "ConnectFour.MakeMove", Body: reqBody})
	if err != nil {
		return nil, err
	}

	switch resp.Name {
	case "ConnectFour.MakeMoveResponse":
		var openGameResp connectfourv1.OpenGameResponse
		err = proto.Unmarshal(resp.Body, &openGameResp)
		if err != nil {
			return nil, err
		}

		return nil, nil
	case "Common.ErrorResponse":
		return api.NewErrorResponse(resp.Body)
	default:
		return nil, errors.New("unknown response")
	}
}
