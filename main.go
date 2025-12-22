package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/gaming-platform/connect-four-bot/internal/bot"
	"github.com/gaming-platform/connect-four-bot/internal/chat"
	"github.com/gaming-platform/connect-four-bot/internal/config"
	"github.com/gaming-platform/connect-four-bot/internal/connectfour"
	"github.com/gaming-platform/connect-four-bot/internal/identity"
	"github.com/gaming-platform/connect-four-bot/internal/rpcclient"
	"github.com/gaming-platform/connect-four-bot/internal/sse"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
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
	rpcClient = rpcclient.NewPrometheusClient(rpcClient)
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

	eg, egCtx := errgroup.WithContext(ctx)

	for _, bt := range bots {
		eg.Go(func() error {
			return bt.Play(egCtx)
		})
	}

	eg.Go(func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		srv := &http.Server{Addr: ":80", Handler: mux}

		go func() {
			<-egCtx.Done()
			_ = srv.Shutdown(context.Background())
		}()

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	if err := eg.Wait(); err != nil && ctx.Err() == nil {
		log.Fatalf("fatal error: %v", err)
	}
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
