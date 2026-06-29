package ai

import (
	"context"
	"errors"

	openai "github.com/sashabaranov/go-openai"
)

type codexProvider struct {
	model  string
	apiKey string
}

// NewCodex returns an AIProvider that calls the OpenAI API.
func NewCodex(model, apiKey string) AIProvider {
	return &codexProvider{model: model, apiKey: apiKey}
}

func (c *codexProvider) Name() string { return "codex" }

func (c *codexProvider) GenerateSpec(ctx context.Context, userMessage string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("api key is required")
	}
	client := openai.NewClient(c.apiKey)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: specPrompt(userMessage)},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty response from Codex")
	}
	return resp.Choices[0].Message.Content, nil
}
