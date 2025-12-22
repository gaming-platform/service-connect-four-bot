package sse

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	externalsse "github.com/tmaxmax/go-sse"
)

type ConnectChannelResult struct {
	Event any
	Error error
}

type Client struct {
	NchanSubUrl string
}

func NewClient(nchanSubUrl string) *Client {
	return &Client{
		NchanSubUrl: nchanSubUrl,
	}
}

func (s *Client) Connect(ctx context.Context, sseCh string) (chan ConnectChannelResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.NchanSubUrl+"?id="+sseCh, nil)
	if err != nil {
		return nil, err
	}

	resChan := make(chan ConnectChannelResult, 1)
	go (func() {
		defer close(resChan)
		client := externalsse.Client{Backoff: externalsse.Backoff{MaxRetries: -1}}
		conn := client.NewConnection(req)
		unsubscribe := conn.SubscribeToAll(func(e externalsse.Event) {
			parts := strings.SplitN(e.Data, ":", 3)
			if len(parts) < 3 {
				return
			}

			payload := []byte(parts[2])
			var event any
			var err error

			switch parts[0] {
			case "ConnectFour.GameOpened":
				event, err = castPayloadToEvent[GameOpened](payload)
			case "ConnectFour.PlayerJoined":
				event, err = castPayloadToEvent[PlayerJoined](payload)
			case "ConnectFour.ChatAssigned":
				event, err = castPayloadToEvent[ChatAssigned](payload)
			case "ConnectFour.PlayerMoved":
				event, err = castPayloadToEvent[PlayerMoved](payload)
			case "ConnectFour.GameAborted":
				event, err = castPayloadToEvent[GameAborted](payload)
			case "ConnectFour.GameWon":
				event, err = castPayloadToEvent[GameWon](payload)
			case "ConnectFour.GameDrawn":
				event, err = castPayloadToEvent[GameDrawn](payload)
			case "ConnectFour.GameTimedOut":
				event, err = castPayloadToEvent[GameTimedOut](payload)
			case "ConnectFour.GameResigned":
				event, err = castPayloadToEvent[GameResigned](payload)
			default:
				return
			}

			if err != nil {
				resChan <- ConnectChannelResult{Error: err}
				return
			}

			resChan <- ConnectChannelResult{Event: event}
		})
		defer unsubscribe()

		resChan <- ConnectChannelResult{Error: conn.Connect()}
	})()

	return resChan, nil
}

func castPayloadToEvent[T any](payload []byte) (T, error) {
	var event T
	err := json.Unmarshal(payload, &event)

	return event, err
}
