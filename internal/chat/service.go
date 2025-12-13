package chat

import (
	"context"
	"errors"

	chatv1 "github.com/gaming-platform/api/go/chat/v1"
	commonv1 "github.com/gaming-platform/api/go/common/v1"
	"github.com/gaming-platform/connect-four-bot/internal/api"
	"github.com/gaming-platform/connect-four-bot/internal/rpcclient"
	"google.golang.org/protobuf/proto"
)

type ChatService struct {
	rpcClient rpcclient.RpcClient
}

func NewChatService(rpcCl rpcclient.RpcClient) *ChatService {
	return &ChatService{
		rpcClient: rpcCl,
	}
}

func (s *ChatService) WriteMessage(
	ctx context.Context,
	chatId string,
	authorId string,
	message string,
	idempotencyKey string,
) (*api.ErrorResponse, error) {
	req := chatv1.WriteMessage{IdempotencyKey: idempotencyKey, ChatId: chatId, AuthorId: authorId, Message: message}
	reqBody, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}

	resp, err := s.rpcClient.Call(ctx, rpcclient.Message{Name: chatv1.WriteMessageType, Body: reqBody})
	if err != nil {
		return nil, err
	}

	switch resp.Name {
	case chatv1.WriteMessageResponseType:
		return nil, nil
	case commonv1.ErrorResponseType:
		return api.NewErrorResponse(resp.Body)
	default:
		return nil, errors.New("unknown response")
	}
}
