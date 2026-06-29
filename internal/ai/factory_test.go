package ai_test

import (
	"testing"

	"github.com/yuying/intake-agent/internal/ai"
)

func TestNewFactory(t *testing.T) {
	tests := []struct {
		provider string
		wantName string
		wantErr  bool
	}{
		{"claude", "claude", false},
		{"gemini", "gemini", false},
		{"codex", "codex", false},
		{"unknown", "", true},
	}
	for _, tt := range tests {
		p, err := ai.New(tt.provider, "model", "key")
		if tt.wantErr {
			if err == nil {
				t.Errorf("provider %s: expected error", tt.provider)
			}
			continue
		}
		if err != nil {
			t.Errorf("provider %s: unexpected error: %v", tt.provider, err)
		}
		if p.Name() != tt.wantName {
			t.Errorf("provider %s: expected name %s, got %s", tt.provider, tt.wantName, p.Name())
		}
	}
}
