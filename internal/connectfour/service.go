package connectfour

import (
	"context"
	"errors"

	commonv1 "github.com/gaming-platform/api/go/common/v1"
	connectfourv1 "github.com/gaming-platform/api/go/connectfour/v1"
	"github.com/gaming-platform/connect-four-bot/internal/api"
	"github.com/gaming-platform/connect-four-bot/internal/rpcclient"
	"google.golang.org/protobuf/proto"
)

type GameService struct {
	rpcClient rpcclient.RpcClient
}

func NewGameService(rpcCl rpcclient.RpcClient) *GameService {
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
) (string, error) {
	req := connectfourv1.OpenGame{
		PlayerId: playerId,
		Width:    width,
		Height:   height,
		Stone:    connectfourv1.OpenGame_Stone(stone),
		Timer:    timer,
	}
	reqBody, err := proto.Marshal(&req)
	if err != nil {
		return "", err
	}

	resp, err := s.rpcClient.Call(ctx, rpcclient.Message{Name: connectfourv1.OpenGameType, Body: reqBody})
	if err != nil {
		return "", err
	}

	switch resp.Name {
	case connectfourv1.OpenGameResponseType:
		var openGameResp connectfourv1.OpenGameResponse
		err = proto.Unmarshal(resp.Body, &openGameResp)
		if err != nil {
			return "", err
		}

		return openGameResp.GameId, nil
	case commonv1.ErrorResponseType:
		return "", api.ErrorResponseToError(resp.Body)
	default:
		return "", errors.New("unknown response")
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

	resp, err := s.rpcClient.Call(ctx, rpcclient.Message{Name: connectfourv1.MakeMoveType, Body: reqBody})
	if err != nil {
		return nil, err
	}

	switch resp.Name {
	case connectfourv1.MakeMoveResponseType:
		return nil, nil
	case commonv1.ErrorResponseType:
		return api.NewErrorResponse(resp.Body)
	default:
		return nil, errors.New("unknown response")
	}
}

func (s *GameService) GetGamesByPlayer(
	ctx context.Context,
	playerId string,
	state connectfourv1.GetGamesByPlayer_State,
	page int32,
	limit int32,
) (*connectfourv1.GetGamesByPlayerResponse, error) {
	req := connectfourv1.GetGamesByPlayer{PlayerId: playerId, State: state, Page: page, Limit: limit}
	reqBody, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}

	resp, err := s.rpcClient.Call(ctx, rpcclient.Message{Name: connectfourv1.GetGamesByPlayerType, Body: reqBody})
	if err != nil {
		return nil, err
	}

	switch resp.Name {
	case connectfourv1.GetGamesByPlayerResponseType:
		var getGamesByPlayerResp connectfourv1.GetGamesByPlayerResponse
		err = proto.Unmarshal(resp.Body, &getGamesByPlayerResp)
		if err != nil {
			return nil, err
		}

		return &getGamesByPlayerResp, nil
	case commonv1.ErrorResponseType:
		return nil, api.ErrorResponseToError(resp.Body)
	default:
		return nil, errors.New("unknown response")
	}
}
