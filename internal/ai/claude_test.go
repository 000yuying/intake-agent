package ai_test

import (
	"context"
	"strings"
	"testing"

	"github.com/yuying/intake-agent/internal/ai"
)

func TestClaudeName(t *testing.T) {
	p := ai.NewClaude("claude-sonnet-4-6", "fake-key")
	if p.Name() != "claude" {
		t.Errorf("expected claude, got %s", p.Name())
	}
}

func TestClaudeGenerateSpec_EmptyKey(t *testing.T) {
	p := ai.NewClaude("claude-sonnet-4-6", "")
	_, err := p.GenerateSpec(context.Background(), "test message")
	if err == nil {
		t.Error("expected error with empty API key")
	}
	if !strings.Contains(err.Error(), "api key") {
		t.Errorf("expected 'api key' in error, got: %v", err)
	}
}
