package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
	discordadapter "github.com/yuying/intake-agent/internal/adapter/discord"
	gchatadapter "github.com/yuying/intake-agent/internal/adapter/gchat"
	githubadapter "github.com/yuying/intake-agent/internal/adapter/github"
	jiraadapter "github.com/yuying/intake-agent/internal/adapter/jira"
	notionadapter "github.com/yuying/intake-agent/internal/adapter/notion"
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	confirm.StartCleanup(ctx)

	mux := http.NewServeMux()
	var adapters []adapter.Adapter
	if cfg.Adapters.Telegram.Enabled {
		adapters = append(adapters, telegramadapter.New(cfg.Adapters.Telegram.Token))
	}
	if cfg.Adapters.Slack.Enabled {
		adapters = append(adapters, slackadapter.NewWithMux(cfg.Adapters.Slack.SigningSecret, cfg.Adapters.Slack.BotToken, mux))
	}
	if cfg.Adapters.Discord.Enabled {
		adapters = append(adapters, discordadapter.New(cfg.Adapters.Discord.Token))
	}
	if cfg.Adapters.GChat.Enabled {
		adapters = append(adapters, gchatadapter.NewWithMux(cfg.Adapters.GChat.WebhookURL, mux))
	}
	if cfg.Adapters.GitHub.Enabled {
		adapters = append(adapters, githubadapter.NewWithMux(
			cfg.Adapters.GitHub.WebhookSecret,
			cfg.Adapters.GitHub.Token,
			cfg.Adapters.GitHub.Repo,
			mux,
		))
	}
	if cfg.Adapters.Jira.Enabled {
		adapters = append(adapters, jiraadapter.NewWithMux(
			cfg.Adapters.Jira.Host,
			cfg.Adapters.Jira.Email,
			cfg.Adapters.Jira.APIToken,
			mux,
		))
	}
	if cfg.Adapters.Notion.Enabled {
		interval := time.Duration(cfg.Adapters.Notion.PollIntervalSeconds) * time.Second
		if interval == 0 {
			interval = 60 * time.Second
		}
		adapters = append(adapters, notionadapter.New(
			cfg.Adapters.Notion.Token,
			cfg.Adapters.Notion.DatabaseID,
			interval,
		))
	}

	eng := engine.NewEngine(confirm, adapters...)

	log.Printf("intake-agent starting on :%d (AI: %s)", cfg.Server.Port, aiProvider.Name())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), mux); err != nil && ctx.Err() == nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	if err := eng.Run(ctx); err != nil {
		log.Fatalf("engine error: %v", err)
	}
	log.Println("intake-agent stopped gracefully")
}
