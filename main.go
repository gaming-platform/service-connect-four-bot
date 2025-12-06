package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gaming-platform/connect-four-bot/internal/bot"
	"github.com/gaming-platform/connect-four-bot/internal/config"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/identity"
	"github.com/gaming-platform/connect-four-bot/internal/rpcclient"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	rpcClient, err := rpcclient.NewAmqpRpcClient(
		cfg.RabbitMqDsn,
		cfg.RpcTimeout,
		rpcclient.NewRouteMessagesToExchange(cfg.RpcExchange),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rpcClient.Close()

	sseClient := sse.NewClient(cfg.NchanSubUrl)
	idSvc := identity.NewBotService(rpcClient)
	gSvc := connectfour.NewGameServiceService(rpcClient)

	botId, err := requestBotId(ctx, idSvc, cfg.Username)
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go (func() {
		bt := bot.NewOpeningBot(botId, sseClient, gSvc)

		for {
			if err := bt.Play(ctx); err != nil {
				log.Fatal(err)
			} else if ctx.Err() != nil {
				break
			}
		}

		wg.Done()
	})()

	<-ctx.Done()
	wg.Wait()
}

func requestBotId(ctx context.Context, botSvc *identity.BotService, username string) (string, error) {
	bt, err := botSvc.GetBotByUsername(ctx, username)
	if err != nil {
		return "", err
	} else if bt == nil {
		return botSvc.RegisterBot(ctx, username)
	} else {
		return bt.BotId, nil
	}
}
