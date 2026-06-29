package ai_test

import (
	"testing"

	"github.com/yuying/intake-agent/internal/ai"
)

func TestGeminiName(t *testing.T) {
	p := ai.NewGemini("gemini-2.0-flash", "fake-key")
	if p.Name() != "gemini" {
		t.Errorf("expected gemini, got %s", p.Name())
	}
}

func TestGeminiEmptyKey(t *testing.T) {
	p := ai.NewGemini("gemini-2.0-flash", "")
	_, err := p.GenerateSpec(nil, "test")
	if err == nil {
		t.Error("expected error with empty API key")
	}
}
