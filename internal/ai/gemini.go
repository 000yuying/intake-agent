package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type geminiProvider struct {
	model  string
	apiKey string
}

// NewGemini returns an AIProvider that calls the Google Gemini API.
func NewGemini(model, apiKey string) AIProvider {
	return &geminiProvider{model: model, apiKey: apiKey}
}

func (g *geminiProvider) Name() string { return "gemini" }

func (g *geminiProvider) GenerateSpec(ctx context.Context, userMessage string) (string, error) {
	if g.apiKey == "" {
		return "", errors.New("api key is required")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(g.apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()
	model := client.GenerativeModel(g.model)
	resp, err := model.GenerateContent(ctx, genai.Text(specPrompt(userMessage)))
	if err != nil {
		return "", err
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from Gemini")
	}
	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}
