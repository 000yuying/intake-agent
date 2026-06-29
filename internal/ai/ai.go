package ai

import "context"

// AIProvider is the interface for AI-backed spec generation.
type AIProvider interface {
	Name() string
	GenerateSpec(ctx context.Context, userMessage string) (string, error)
}
