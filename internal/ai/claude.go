package ai

import (
	"context"
	"errors"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type claudeProvider struct {
	model  string
	apiKey string
}

// NewClaude returns an AIProvider that calls the Anthropic Claude API.
func NewClaude(model, apiKey string) AIProvider {
	return &claudeProvider{model: model, apiKey: apiKey}
}

func (c *claudeProvider) Name() string { return "claude" }

func (c *claudeProvider) GenerateSpec(ctx context.Context, userMessage string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("api key is required")
	}
	client := anthropic.NewClient(option.WithAPIKey(c.apiKey))
	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(specPrompt(userMessage))),
		},
	})
	if err != nil {
		return "", err
	}
	if len(msg.Content) == 0 {
		return "", errors.New("empty response from Claude")
	}
	return msg.Content[0].Text, nil
}

func specPrompt(userMessage string) string {
	return `你是一個需求分析師。根據以下訊息，產出一份簡潔的 spec Markdown。

格式：
## 需求概述
（一段話說明這個需求是什麼）

## 驗收條件
- （條列式 AC，每條以 Given/When/Then 或明確的可測量標準描述）

## 範圍外
- （不在本次範圍的相關項目）

---
訊息：` + userMessage
}
