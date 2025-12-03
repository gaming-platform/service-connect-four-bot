package rpcclient

import (
	"context"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type amqpRpcClient struct {
	connection           *amqp091.Connection
	channel              *amqp091.Channel
	messageRouter        AmqpMessageRouter
	responsesMessages    <-chan amqp091.Delivery
	responseChannels     *responseChannels
	correlationIdCounter *counter
	defaultTimeout       time.Duration
	closeChannel         chan struct{}
	syncClose            sync.Once
}

func NewAmqpRpcClient(url string, timeout time.Duration, msgRouter AmqpMessageRouter) (RpcClient, error) {
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

	client := &amqpRpcClient{
		connection:           conn,
		channel:              ch,
		messageRouter:        msgRouter,
		responsesMessages:    respMsgs,
		correlationIdCounter: newCounter(),
		responseChannels:     newResponseChannels(),
		defaultTimeout:       timeout,
		closeChannel:         make(chan struct{}),
	}

	go client.consumeResponses()

	return client, nil
}

func (c *amqpRpcClient) Call(ctx context.Context, req Message) (Message, error) {
	corrId := c.correlationIdCounter.nextString()

	respChan := c.responseChannels.add(corrId)
	defer c.responseChannels.delete(corrId)

	ctx, cancel := withTimeoutIfNone(ctx, c.defaultTimeout)
	defer cancel()

	route := c.messageRouter.Route(req)

	err := c.channel.PublishWithContext(
		ctx,
		route.exchange,
		route.routingKey,
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

func (c *amqpRpcClient) Close() error {
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

func (c *amqpRpcClient) consumeResponses() {
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

func (rc *responseChannels) add(corrId string) chan amqp091.Delivery {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	respChan := make(chan amqp091.Delivery)
	rc.responses[corrId] = respChan
	return respChan
}

func (rc *responseChannels) get(corrId string) (chan amqp091.Delivery, bool) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	respChan, ok := rc.responses[corrId]
	return respChan, ok
}

func (rc *responseChannels) delete(corrId string) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	delete(rc.responses, corrId)
}

type AmqpMessageRouter interface {
	Route(msg Message) Route
}

type Route struct {
	exchange   string
	routingKey string
}

type RouteMessagesToExchange struct {
	exchange string
}

func NewRouteMessagesToExchange(exchange string) *RouteMessagesToExchange {
	return &RouteMessagesToExchange{exchange: exchange}
}

func (r *RouteMessagesToExchange) Route(msg Message) Route {
	return Route{
		exchange:   r.exchange,
		routingKey: msg.Name,
	}
}
