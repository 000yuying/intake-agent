package ai_test

import (
	"testing"

	"github.com/yuying/intake-agent/internal/ai"
)

func TestCodexName(t *testing.T) {
	p := ai.NewCodex("gpt-4o", "fake-key")
	if p.Name() != "codex" {
		t.Errorf("expected codex, got %s", p.Name())
	}
}

func TestCodexEmptyKey(t *testing.T) {
	p := ai.NewCodex("gpt-4o", "")
	_, err := p.GenerateSpec(nil, "test")
	if err == nil {
		t.Error("expected error with empty API key")
	}
}
