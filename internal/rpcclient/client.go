package rpcclient

import "context"

type RpcClient interface {
	Call(ctx context.Context, req Message) (Message, error)
	Close() error
}

type Message struct {
	Name string
	Body []byte
}
