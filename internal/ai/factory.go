package ai

import "fmt"

// New returns an AIProvider for the given provider name.
// Supported providers: "claude", "gemini", "codex".
func New(provider, model, apiKey string) (AIProvider, error) {
	switch provider {
	case "claude":
		return NewClaude(model, apiKey), nil
	case "gemini":
		return NewGemini(model, apiKey), nil
	case "codex":
		return NewCodex(model, apiKey), nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", provider)
	}
}
