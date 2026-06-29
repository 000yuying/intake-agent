package ai

import (
	"context"
	"errors"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type geminiProvider struct {
	model    string
	apiKey   string
	once     sync.Once
	client   *genai.Client
	clientErr error
}

// NewGemini returns an AIProvider that calls the Google Gemini API.
func NewGemini(model, apiKey string) AIProvider {
	return &geminiProvider{model: model, apiKey: apiKey}
}

func (g *geminiProvider) Name() string { return "gemini" }

func (g *geminiProvider) initClient(ctx context.Context) error {
	g.once.Do(func() {
		g.client, g.clientErr = genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
	})
	return g.clientErr
}

func (g *geminiProvider) GenerateSpec(ctx context.Context, userMessage string) (string, error) {
	if g.apiKey == "" {
		return "", errors.New("api key is required")
	}
	if err := g.initClient(ctx); err != nil {
		return "", err
	}
	model := g.client.GenerativeModel(g.model)
	resp, err := model.GenerateContent(ctx, genai.Text(specPrompt(userMessage)))
	if err != nil {
		return "", err
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from Gemini")
	}
	part, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", errors.New("unexpected response type from Gemini")
	}
	return string(part), nil
}

// Close releases the underlying gRPC connection held by the Gemini client.
func (g *geminiProvider) Close() error {
	g.once.Do(func() {}) // ensure once is marked done so client state is consistent
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}
