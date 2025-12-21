package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gaming-platform/connect-four-bot/internal/bot"
	"github.com/gaming-platform/connect-four-bot/internal/chat"
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
		rpcclient.NewRouteMessagesToTransientRpcQueue(cfg.RpcExchange, "TransientRpc"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rpcClient.Close()

	sseClient := sse.NewClient(cfg.NchanSubUrl)
	chatSvc := chat.NewChatService(rpcClient)
	botSvc := identity.NewBotService(rpcClient)
	gameSvc := connectfour.NewGameService(rpcClient)

	botId, err := requestBotId(ctx, botSvc, cfg.Username)
	if err != nil {
		log.Fatal(err)
	}

	resumingBot, err := bot.NewResumingBot(ctx, botId, sseClient, chatSvc, gameSvc)
	if err != nil {
		log.Fatal(err)
	}

	bots := [...]bot.Bot{
		bot.NewOpeningBot(botId, sseClient, chatSvc, gameSvc),
		resumingBot,
	}

	wg := sync.WaitGroup{}
	wg.Add(len(bots))
	for _, bt := range bots {
		go (func() {
			if err := bt.Play(ctx); err != nil && ctx.Err() == nil {
				log.Fatal(err)
			}
			wg.Done()
		})()
	}

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
