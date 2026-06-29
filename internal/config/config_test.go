package config_test

import (
	"os"
	"testing"
	"github.com/yuying/intake-agent/internal/config"
)

func TestLoad(t *testing.T) {
	content := `
server:
  port: 9090
ai:
  provider: claude
  model: claude-sonnet-4-6
output:
  repo_path: /tmp/specs
  dir: specs/
adapters:
  telegram:
    enabled: true
    token: "test-token"
  slack:
    enabled: false
    signing_secret: ""
    bot_token: ""
`
	f, _ := os.CreateTemp("", "config-*.yaml")
	f.WriteString(content)
	f.Close()
	defer os.Remove(f.Name())

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.AI.Provider != "claude" {
		t.Errorf("expected provider claude, got %s", cfg.AI.Provider)
	}
	if !cfg.Adapters.Telegram.Enabled {
		t.Error("expected telegram enabled")
	}
	if cfg.Adapters.Telegram.Token != "test-token" {
		t.Errorf("expected token test-token, got %s", cfg.Adapters.Telegram.Token)
	}
}
