package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	AI       AIConfig       `yaml:"ai"`
	Output   OutputConfig   `yaml:"output"`
	Adapters AdaptersConfig `yaml:"adapters"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type AIConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
}

type OutputConfig struct {
	RepoPath string `yaml:"repo_path"`
	Dir      string `yaml:"dir"`
}

type AdaptersConfig struct {
	Telegram TelegramConfig `yaml:"telegram"`
	Slack    SlackConfig    `yaml:"slack"`
	Discord  DiscordConfig  `yaml:"discord"`
	GChat    GChatConfig    `yaml:"gchat"`
	GitHub   GitHubConfig   `yaml:"github"`
}

type TelegramConfig struct {
	Enabled bool   `yaml:"enabled"`
	Token   string `yaml:"token"`
}

type SlackConfig struct {
	Enabled       bool   `yaml:"enabled"`
	SigningSecret string `yaml:"signing_secret"`
	BotToken      string `yaml:"bot_token"`
}

type DiscordConfig struct {
	Enabled bool   `yaml:"enabled"`
	Token   string `yaml:"token"`
}

type GChatConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
}

type GitHubConfig struct {
	Enabled       bool   `yaml:"enabled"`
	WebhookSecret string `yaml:"webhook_secret"`
	Token         string `yaml:"token"`
	Repo          string `yaml:"repo"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
