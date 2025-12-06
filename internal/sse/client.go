package sse

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	externalsse "github.com/tmaxmax/go-sse"
)

type EventCallback func(string, map[string]interface{})

type ConnectChannelResult struct {
	Event Event
	Error error
}

type Event struct {
	Name    string
	Payload map[string]interface{}
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
		conn := externalsse.NewConnection(req)
		unsubscribe := conn.SubscribeToAll(func(e externalsse.Event) {
			parts := strings.SplitN(e.Data, ":", 3)
			if len(parts) < 3 {
				return
			}

			var msg map[string]interface{}
			if err := json.Unmarshal([]byte(parts[2]), &msg); err != nil {
				return
			}

			resChan <- ConnectChannelResult{Event: Event{Name: parts[0], Payload: msg}}
		})
		defer unsubscribe()

		resChan <- ConnectChannelResult{Error: conn.Connect()}
	})()

	return resChan, nil
}
