package main

import (
	"context"
	"log"

	"github.com/gaming-platform/connect-four-bot/internal/config"
	"github.com/gaming-platform/connect-four-bot/internal/identity"
	"github.com/gaming-platform/connect-four-bot/internal/rpcclient"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	rpcClient, err := rpcclient.NewAmqpRpcClient(cfg.RabbitMqDsn, cfg.RpcTimeout)
	if err != nil {
		log.Fatal(err)
	}
	defer rpcClient.Close()

	botId, err := requestBotId(ctx, identity.NewBotService(rpcClient), cfg.Username)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Continuing with bot id: %s", botId)
}

func requestBotId(ctx context.Context, botSvc *identity.BotService, username string) (string, error) {
	bot, err := botSvc.GetBotByUsername(ctx, username)
	if err != nil {
		return "", err
	} else if bot == nil {
		return botSvc.RegisterBot(ctx, username)
	} else {
		return bot.BotId, nil
	}
}
