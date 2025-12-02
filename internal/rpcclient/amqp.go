package rpcclient

import (
	"context"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type AmqpRpcClient struct {
	connection           *amqp091.Connection
	channel              *amqp091.Channel
	responsesMessages    <-chan amqp091.Delivery
	responseChannels     *responseChannels
	correlationIdCounter *counter
	defaultTimeout       time.Duration
	closeChannel         chan struct{}
	syncClose            sync.Once
}

func NewAmqpRpcClient(url string, timeout time.Duration) (*AmqpRpcClient, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	respMsgs, err := ch.Consume(
		"amq.rabbitmq.reply-to",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	client := &AmqpRpcClient{
		connection:           conn,
		channel:              ch,
		responsesMessages:    respMsgs,
		correlationIdCounter: newCounter(),
		responseChannels:     newResponseChannels(),
		defaultTimeout:       timeout,
		closeChannel:         make(chan struct{}),
	}

	go client.consumeResponses()

	return client, nil
}

func (c *AmqpRpcClient) Call(ctx context.Context, req Message) (Message, error) {
	corrId := c.correlationIdCounter.nextString()

	respChan := c.responseChannels.add(corrId)
	defer c.responseChannels.delete(corrId)

	ctx, cancel := withTimeoutIfNone(ctx, c.defaultTimeout)
	defer cancel()

	err := c.channel.PublishWithContext(
		ctx,
		"gaming",
		req.Name,
		false,
		false,
		amqp091.Publishing{
			CorrelationId: corrId,
			ReplyTo:       "amq.rabbitmq.reply-to",
			Type:          req.Name,
			Body:          req.Body,
		},
	)
	if err != nil {
		return Message{}, err
	}

	select {
	case resp := <-respChan:
		return Message{Name: resp.Type, Body: resp.Body}, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

func (c *AmqpRpcClient) Close() error {
	var err error

	c.syncClose.Do(func() {
		close(c.closeChannel)
		if err = c.channel.Close(); err != nil {
			return
		}
		if err = c.connection.Close(); err != nil {
			return
		}
	})

	return err
}

func (c *AmqpRpcClient) consumeResponses() {
	for {
		select {
		case resp := <-c.responsesMessages:
			if respChan, ok := c.responseChannels.get(resp.CorrelationId); ok {
				respChan <- resp
			}
		case <-c.closeChannel:
			return
		}
	}
}

type responseChannels struct {
	responses map[string]chan amqp091.Delivery
	mutex     sync.Mutex
}

func newResponseChannels() *responseChannels {
	return &responseChannels{
		responses: make(map[string]chan amqp091.Delivery),
	}
}

func (cr *responseChannels) add(corrId string) chan amqp091.Delivery {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	respChan := make(chan amqp091.Delivery)
	cr.responses[corrId] = respChan
	return respChan
}

func (cr *responseChannels) get(corrId string) (chan amqp091.Delivery, bool) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	respChan, ok := cr.responses[corrId]
	return respChan, ok
}

func (cr *responseChannels) delete(corrId string) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	delete(cr.responses, corrId)
}
