package identity

import (
	"context"
	"errors"

	commonv1 "github.com/gaming-platform/api/go/common/v1"
	identityv1 "github.com/gaming-platform/api/go/identity/v1"
	"github.com/gaming-platform/connect-four-bot/internal/api"
	"github.com/gaming-platform/connect-four-bot/internal/rpcclient"
	"google.golang.org/protobuf/proto"
)

type BotService struct {
	rpcClient rpcclient.RpcClient
}

func NewBotService(rpcCl rpcclient.RpcClient) *BotService {
	return &BotService{
		rpcClient: rpcCl,
	}
}

func (s *BotService) RegisterBot(ctx context.Context, username string) (string, error) {
	req := identityv1.RegisterBot{Username: username}
	reqBody, err := proto.Marshal(&req)
	if err != nil {
		return "", err
	}

	resp, err := s.rpcClient.Call(ctx, rpcclient.Message{Name: identityv1.RegisterBotType, Body: reqBody})
	if err != nil {
		return "", err
	}

	switch resp.Name {
	case identityv1.RegisterBotResponseType:
		var regBotResp identityv1.RegisterBotResponse
		err = proto.Unmarshal(resp.Body, &regBotResp)
		if err != nil {
			return "", err
		}

		return regBotResp.BotId, nil
	case commonv1.ErrorResponseType:
		return "", api.ErrorResponseToError(resp.Body)
	default:
		return "", errors.New("unknown response")
	}
}

func (s *BotService) GetBotByUsername(ctx context.Context, username string) (*Bot, error) {
	req := identityv1.GetBotByUsername{Username: username}
	reqBody, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}

	resp, err := s.rpcClient.Call(ctx, rpcclient.Message{Name: identityv1.GetBotByUsernameType, Body: reqBody})
	if err != nil {
		return nil, err
	}

	switch resp.Name {
	case identityv1.GetBotByUsernameResponseType:
		var getBotResp identityv1.GetBotByUsernameResponse
		err = proto.Unmarshal(resp.Body, &getBotResp)
		if err != nil {
			return nil, err
		}

		return fromProtoBot(getBotResp.Bot), nil
	case commonv1.ErrorResponseType:
		return nil, api.ErrorResponseToError(resp.Body)
	default:
		return nil, errors.New("unknown response")
	}
}
