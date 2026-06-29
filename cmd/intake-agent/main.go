package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
	discordadapter "github.com/yuying/intake-agent/internal/adapter/discord"
	gchatadapter "github.com/yuying/intake-agent/internal/adapter/gchat"
	slackadapter "github.com/yuying/intake-agent/internal/adapter/slack"
	telegramadapter "github.com/yuying/intake-agent/internal/adapter/telegram"
	"github.com/yuying/intake-agent/internal/ai"
	"github.com/yuying/intake-agent/internal/config"
	"github.com/yuying/intake-agent/internal/engine"
	"github.com/yuying/intake-agent/internal/output"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	apiKey := os.Getenv("AI_API_KEY")
	aiProvider, err := ai.New(cfg.AI.Provider, cfg.AI.Model, apiKey)
	if err != nil {
		log.Fatalf("failed to create AI provider: %v", err)
	}

	writer := output.NewWriter(cfg.Output.RepoPath, cfg.Output.Dir)
	confirm := engine.NewConfirmEngine(aiProvider, writer, 600*time.Second)

	var adapters []adapter.Adapter
	if cfg.Adapters.Telegram.Enabled {
		adapters = append(adapters, telegramadapter.New(cfg.Adapters.Telegram.Token))
	}
	if cfg.Adapters.Slack.Enabled {
		adapters = append(adapters, slackadapter.New(cfg.Adapters.Slack.SigningSecret, cfg.Adapters.Slack.BotToken))
	}
	if cfg.Adapters.Discord.Enabled {
		adapters = append(adapters, discordadapter.New(cfg.Adapters.Discord.Token))
	}
	if cfg.Adapters.GChat.Enabled {
		adapters = append(adapters, gchatadapter.New(cfg.Adapters.GChat.WebhookURL))
	}

	eng := engine.NewEngine(confirm, adapters...)

	log.Printf("intake-agent starting on :%d (AI: %s)", cfg.Server.Port, aiProvider.Name())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	if err := eng.Run(context.Background()); err != nil {
		log.Fatalf("engine error: %v", err)
	}
}
