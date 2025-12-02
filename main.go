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

	botSvc := identity.NewBotService(rpcClient)

	bot, err := botSvc.GetBotByUsername(ctx, cfg.Username)
	if err != nil {
		log.Fatal(err)
	}

	var botId string
	switch bot {
	case nil:
		botId, err = botSvc.RegisterBot(ctx, cfg.Username)
		if err != nil {
			log.Fatal(err)
		}
	default:
		botId = bot.BotId
	}

	log.Printf("Continuing with bot id: %s", botId)
}
