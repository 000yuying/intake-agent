# intake-agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立一個 webhook-based 多管道訊息收集器，透過可抽換的 AI Provider 產出 spec Markdown 草稿，並在原管道完成人工確認後寫入 git repo。

**Architecture:** HTTP server 接收各管道 webhook，統一轉為 `Message` struct 送入 Core Engine；Engine 呼叫 AIProvider 介面產 spec 草稿，回原管道等確認；確認後 Output Writer 寫入 Markdown。三層（Adapter / Engine / Output）各自透過介面解耦。

**Tech Stack:** Go 1.22+、`gopkg.in/yaml.v3`（config）、`github.com/anthropics/anthropic-sdk-go`（Claude）、`github.com/google/generative-ai-go`（Gemini）、`github.com/sashabaranov/go-openai`（Codex/OpenAI）、`github.com/go-telegram-bot-api/telegram-bot-api/v5`（Telegram）、`github.com/slack-go/slack`（Slack）

## Global Constraints

- Go 1.22+
- 所有介面定義（`Adapter`、`AIProvider`）不可在任務間改變簽名
- config 欄位名稱與 spec 完全一致（snake_case YAML）
- 確認逾時預設 10 分鐘（`600 * time.Second`）
- Markdown spec 檔名格式：`YYYY-MM-DD-HH-MM-SS-<source>.md`
- 每個任務結束必須 `git commit`

---

### Task 1: 專案骨架 + Config 載入

**Files:**
- Create: `go.mod`
- Create: `cmd/intake-agent/main.go`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `configs/config.yaml.example`
- Create: `Makefile`

**Interfaces:**
- Produces:
  - `config.Config` struct（所有後續任務使用）
  - `config.Load(path string) (*Config, error)`

- [ ] **Step 1: 初始化 Go module**

```bash
cd /home/yuying/intake-agent
go mod init github.com/yuying/intake-agent
```

- [ ] **Step 2: 建立 config_test.go（先寫測試）**

```go
// internal/config/config_test.go
package config_test

import (
    "os"
    "testing"
    "github.com/yuying/intake-agent/internal/config"
)

func TestLoad(t *testing.T) {
    content := `
server:
  port: 9090
ai:
  provider: claude
  model: claude-sonnet-4-6
output:
  repo_path: /tmp/specs
  dir: specs/
adapters:
  telegram:
    enabled: true
    token: "test-token"
  slack:
    enabled: false
    signing_secret: ""
    bot_token: ""
`
    f, _ := os.CreateTemp("", "config-*.yaml")
    f.WriteString(content)
    f.Close()
    defer os.Remove(f.Name())

    cfg, err := config.Load(f.Name())
    if err != nil {
        t.Fatalf("Load error: %v", err)
    }
    if cfg.Server.Port != 9090 {
        t.Errorf("expected port 9090, got %d", cfg.Server.Port)
    }
    if cfg.AI.Provider != "claude" {
        t.Errorf("expected provider claude, got %s", cfg.AI.Provider)
    }
    if !cfg.Adapters.Telegram.Enabled {
        t.Error("expected telegram enabled")
    }
    if cfg.Adapters.Telegram.Token != "test-token" {
        t.Errorf("expected token test-token, got %s", cfg.Adapters.Telegram.Token)
    }
}
```

- [ ] **Step 3: 執行測試確認失敗**

```bash
cd /home/yuying/intake-agent
go test ./internal/config/...
```
Expected: `cannot find package` 或 compile error

- [ ] **Step 4: 實作 config.go**

```go
// internal/config/config.go
package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    AI       AIConfig       `yaml:"ai"`
    Output   OutputConfig   `yaml:"output"`
    Adapters AdaptersConfig `yaml:"adapters"`
}

type ServerConfig struct {
    Port int `yaml:"port"`
}

type AIConfig struct {
    Provider string `yaml:"provider"`
    Model    string `yaml:"model"`
}

type OutputConfig struct {
    RepoPath string `yaml:"repo_path"`
    Dir      string `yaml:"dir"`
}

type AdaptersConfig struct {
    Telegram TelegramConfig `yaml:"telegram"`
    Slack    SlackConfig    `yaml:"slack"`
    Discord  DiscordConfig  `yaml:"discord"`
}

type TelegramConfig struct {
    Enabled bool   `yaml:"enabled"`
    Token   string `yaml:"token"`
}

type SlackConfig struct {
    Enabled       bool   `yaml:"enabled"`
    SigningSecret string `yaml:"signing_secret"`
    BotToken      string `yaml:"bot_token"`
}

type DiscordConfig struct {
    Enabled bool   `yaml:"enabled"`
    Token   string `yaml:"token"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

- [ ] **Step 5: 安裝依賴**

```bash
cd /home/yuying/intake-agent
go get gopkg.in/yaml.v3
```

- [ ] **Step 6: 執行測試確認通過**

```bash
go test ./internal/config/... -v
```
Expected: `PASS`

- [ ] **Step 7: 建立 configs/config.yaml.example**

```yaml
# configs/config.yaml.example
server:
  port: 8080

ai:
  provider: claude          # claude | gemini | codex
  model: claude-sonnet-4-6  # 依 provider 填對應 model name

output:
  repo_path: /home/yuying/specs
  dir: specs/

adapters:
  telegram:
    enabled: true
    token: ""
  slack:
    enabled: true
    signing_secret: ""
    bot_token: ""
  discord:
    enabled: false
    token: ""
```

- [ ] **Step 8: 建立 Makefile**

```makefile
# Makefile
.PHONY: start stop test build

CONFIG ?= configs/config.yaml

start:
	go run cmd/intake-agent/main.go --config $(CONFIG)

build:
	go build -o bin/intake-agent cmd/intake-agent/main.go

test:
	go test ./... -v

stop:
	pkill -f intake-agent || true
```

- [ ] **Step 9: 建立 main.go 骨架**

```go
// cmd/intake-agent/main.go
package main

import (
    "flag"
    "log"
    "github.com/yuying/intake-agent/internal/config"
)

func main() {
    configPath := flag.String("config", "configs/config.yaml", "path to config file")
    flag.Parse()

    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }
    log.Printf("intake-agent starting on port %d", cfg.Server.Port)
}
```

- [ ] **Step 10: 確認 build 通過**

```bash
go build ./...
```
Expected: no errors

- [ ] **Step 11: Commit**

```bash
git add .
git commit -m "feat: project scaffold with config loading"
```

---

### Task 2: Adapter 介面 + Message struct

**Files:**
- Create: `internal/adapter/adapter.go`
- Create: `internal/adapter/adapter_test.go`

**Interfaces:**
- Produces:
  - `adapter.Message` struct
  - `adapter.Adapter` interface（`Name() string`、`Start(ctx, out chan<- Message) error`、`Reply(ctx, msg Message, text string) error`）

- [ ] **Step 1: 建立 adapter_test.go**

```go
// internal/adapter/adapter_test.go
package adapter_test

import (
    "testing"
    "time"
    "github.com/yuying/intake-agent/internal/adapter"
)

func TestMessageFields(t *testing.T) {
    msg := adapter.Message{
        ID:        "123",
        Source:    "telegram",
        ChannelID: "chan-1",
        UserID:    "user-1",
        Text:      "需要新功能",
        Timestamp: time.Now(),
    }
    if msg.Source != "telegram" {
        t.Errorf("expected source telegram, got %s", msg.Source)
    }
}
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
go test ./internal/adapter/... -v
```
Expected: compile error

- [ ] **Step 3: 實作 adapter.go**

```go
// internal/adapter/adapter.go
package adapter

import (
    "context"
    "time"
)

type Message struct {
    ID        string
    Source    string
    ChannelID string
    UserID    string
    Text      string
    Timestamp time.Time
}

type Adapter interface {
    Name() string
    Start(ctx context.Context, out chan<- Message) error
    Reply(ctx context.Context, msg Message, text string) error
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
go test ./internal/adapter/... -v
```
Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/
git commit -m "feat: adapter interface and Message struct"
```

---

### Task 3: AIProvider 介面 + Claude 實作

**Files:**
- Create: `internal/ai/ai.go`
- Create: `internal/ai/claude.go`
- Create: `internal/ai/claude_test.go`

**Interfaces:**
- Consumes: 無
- Produces:
  - `ai.AIProvider` interface（`Name() string`、`GenerateSpec(ctx, userMessage string) (string, error)`）
  - `ai.NewClaude(model, apiKey string) AIProvider`

- [ ] **Step 1: 建立 claude_test.go**

```go
// internal/ai/claude_test.go
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
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
go test ./internal/ai/... -v
```
Expected: compile error

- [ ] **Step 3: 建立 ai.go（介面）**

```go
// internal/ai/ai.go
package ai

import "context"

type AIProvider interface {
    Name() string
    GenerateSpec(ctx context.Context, userMessage string) (string, error)
}
```

- [ ] **Step 4: 安裝 Anthropic SDK**

```bash
go get github.com/anthropics/anthropic-sdk-go
```

- [ ] **Step 5: 實作 claude.go**

```go
// internal/ai/claude.go
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
```

- [ ] **Step 6: 執行測試確認通過**

```bash
go test ./internal/ai/... -v -run TestClaude
```
Expected: `PASS`（空 key 測試通過，真實 API 呼叫略過）

- [ ] **Step 7: Commit**

```bash
git add internal/ai/ai.go internal/ai/claude.go internal/ai/claude_test.go go.mod go.sum
git commit -m "feat: AIProvider interface and Claude implementation"
```

---

### Task 4: Gemini + Codex AI Provider 實作

**Files:**
- Create: `internal/ai/gemini.go`
- Create: `internal/ai/gemini_test.go`
- Create: `internal/ai/codex.go`
- Create: `internal/ai/codex_test.go`
- Create: `internal/ai/factory.go`
- Create: `internal/ai/factory_test.go`

**Interfaces:**
- Consumes: `ai.AIProvider`（Task 3）
- Produces:
  - `ai.NewGemini(model, apiKey string) AIProvider`
  - `ai.NewCodex(model, apiKey string) AIProvider`
  - `ai.New(provider, model, apiKey string) (AIProvider, error)`

- [ ] **Step 1: 建立 gemini_test.go**

```go
// internal/ai/gemini_test.go
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
```

- [ ] **Step 2: 安裝 Gemini SDK**

```bash
go get github.com/google/generative-ai-go/genai
go get google.golang.org/api/option
```

- [ ] **Step 3: 實作 gemini.go**

```go
// internal/ai/gemini.go
package ai

import (
    "context"
    "errors"
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

type geminiProvider struct {
    model  string
    apiKey string
}

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
```

- [ ] **Step 4: 建立 codex_test.go**

```go
// internal/ai/codex_test.go
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
```

- [ ] **Step 5: 安裝 OpenAI SDK**

```bash
go get github.com/sashabaranov/go-openai
```

- [ ] **Step 6: 實作 codex.go**

```go
// internal/ai/codex.go
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
```

- [ ] **Step 7: 建立 factory_test.go**

```go
// internal/ai/factory_test.go
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
```

- [ ] **Step 8: 實作 factory.go**

```go
// internal/ai/factory.go
package ai

import "fmt"

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
```

- [ ] **Step 9: 執行測試確認通過**

```bash
go test ./internal/ai/... -v
```
Expected: `PASS`

- [ ] **Step 10: Commit**

```bash
git add internal/ai/ go.mod go.sum
git commit -m "feat: Gemini and Codex AI providers with factory"
```

---

### Task 5: Output Writer（Markdown 寫入）

**Files:**
- Create: `internal/output/writer.go`
- Create: `internal/output/writer_test.go`

**Interfaces:**
- Consumes: 無
- Produces:
  - `output.Writer` struct
  - `output.NewWriter(repoPath, dir string) *Writer`
  - `(w *Writer) Write(source, content string) (string, error)` — 回傳寫入的相對檔案路徑

- [ ] **Step 1: 建立 writer_test.go**

```go
// internal/output/writer_test.go
package output_test

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/yuying/intake-agent/internal/output"
)

func TestWrite(t *testing.T) {
    dir := t.TempDir()
    w := output.NewWriter(dir, "specs/")
    path, err := w.Write("telegram", "## 需求概述\n測試需求")
    if err != nil {
        t.Fatalf("Write error: %v", err)
    }
    if !strings.HasPrefix(path, "specs/") {
        t.Errorf("expected path to start with specs/, got %s", path)
    }
    if !strings.HasSuffix(path, "-telegram.md") {
        t.Errorf("expected path to end with -telegram.md, got %s", path)
    }
    full := filepath.Join(dir, path)
    data, err := os.ReadFile(full)
    if err != nil {
        t.Fatalf("ReadFile error: %v", err)
    }
    if !strings.Contains(string(data), "測試需求") {
        t.Error("file content does not contain expected text")
    }
}
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
go test ./internal/output/... -v
```
Expected: compile error

- [ ] **Step 3: 實作 writer.go**

```go
// internal/output/writer.go
package output

import (
    "fmt"
    "os"
    "path/filepath"
    "time"
)

type Writer struct {
    repoPath string
    dir      string
}

func NewWriter(repoPath, dir string) *Writer {
    return &Writer{repoPath: repoPath, dir: dir}
}

func (w *Writer) Write(source, content string) (string, error) {
    if err := os.MkdirAll(filepath.Join(w.repoPath, w.dir), 0755); err != nil {
        return "", err
    }
    ts := time.Now().Format("2006-01-02-15-04-05")
    filename := fmt.Sprintf("%s-%s.md", ts, source)
    relPath := filepath.Join(w.dir, filename)
    fullPath := filepath.Join(w.repoPath, relPath)
    if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
        return "", err
    }
    return relPath, nil
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
go test ./internal/output/... -v
```
Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/output/
git commit -m "feat: output writer for Markdown spec files"
```

---

### Task 6: 確認狀態機（Confirm Engine）

**Files:**
- Create: `internal/engine/confirm.go`
- Create: `internal/engine/confirm_test.go`

**Interfaces:**
- Consumes:
  - `adapter.Message`（Task 2）
  - `ai.AIProvider`（Task 3）
  - `output.Writer`（Task 5）
- Produces:
  - `engine.ConfirmEngine` struct
  - `engine.NewConfirmEngine(ai AIProvider, writer *output.Writer, timeout time.Duration) *ConfirmEngine`
  - `(e *ConfirmEngine) HandleMessage(ctx context.Context, msg adapter.Message) (replyText string, err error)`
  - `(e *ConfirmEngine) HandleConfirm(ctx context.Context, msg adapter.Message) (replyText string, wrote bool, err error)`

- [ ] **Step 1: 建立 confirm_test.go**

```go
// internal/engine/confirm_test.go
package engine_test

import (
    "context"
    "strings"
    "testing"
    "time"
    "github.com/yuying/intake-agent/internal/adapter"
    "github.com/yuying/intake-agent/internal/engine"
    "github.com/yuying/intake-agent/internal/output"
)

type fakeAI struct{}

func (f *fakeAI) Name() string { return "fake" }
func (f *fakeAI) GenerateSpec(_ context.Context, msg string) (string, error) {
    return "## 需求概述\n" + msg, nil
}

func TestHandleMessage(t *testing.T) {
    w := output.NewWriter(t.TempDir(), "specs/")
    e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
    msg := adapter.Message{
        ID: "1", Source: "telegram", ChannelID: "c1", UserID: "u1",
        Text: "新功能：登入頁改版",
    }
    reply, err := e.HandleMessage(context.Background(), msg)
    if err != nil {
        t.Fatalf("HandleMessage error: %v", err)
    }
    if !strings.Contains(reply, "ok") {
        t.Errorf("expected reply to mention 'ok', got: %s", reply)
    }
    if !strings.Contains(reply, "登入頁改版") {
        t.Errorf("expected reply to contain user message, got: %s", reply)
    }
}

func TestHandleConfirm_OK(t *testing.T) {
    w := output.NewWriter(t.TempDir(), "specs/")
    e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
    orig := adapter.Message{ID: "1", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "需求A"}
    e.HandleMessage(context.Background(), orig)

    confirm := adapter.Message{ID: "2", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "ok"}
    reply, wrote, err := e.HandleConfirm(context.Background(), confirm)
    if err != nil {
        t.Fatalf("HandleConfirm error: %v", err)
    }
    if !wrote {
        t.Error("expected wrote=true")
    }
    if !strings.Contains(reply, "specs/") {
        t.Errorf("expected reply to contain file path, got: %s", reply)
    }
}

func TestHandleConfirm_No(t *testing.T) {
    w := output.NewWriter(t.TempDir(), "specs/")
    e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
    orig := adapter.Message{ID: "1", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "需求B"}
    e.HandleMessage(context.Background(), orig)

    confirm := adapter.Message{ID: "2", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "no"}
    _, wrote, err := e.HandleConfirm(context.Background(), confirm)
    if err != nil {
        t.Fatalf("HandleConfirm error: %v", err)
    }
    if wrote {
        t.Error("expected wrote=false for 'no'")
    }
}
```

- [ ] **Step 2: 執行測試確認失敗**

```bash
go test ./internal/engine/... -v
```
Expected: compile error

- [ ] **Step 3: 實作 confirm.go**

```go
// internal/engine/confirm.go
package engine

import (
    "context"
    "fmt"
    "strings"
    "sync"
    "time"
    "github.com/yuying/intake-agent/internal/adapter"
    "github.com/yuying/intake-agent/internal/ai"
    "github.com/yuying/intake-agent/internal/output"
)

type pendingItem struct {
    msg      adapter.Message
    draft    string
    expireAt time.Time
}

type ConfirmEngine struct {
    ai      ai.AIProvider
    writer  *output.Writer
    timeout time.Duration
    mu      sync.Mutex
    pending map[string]*pendingItem // key: channelID+":"+userID
}

func NewConfirmEngine(aiProvider ai.AIProvider, writer *output.Writer, timeout time.Duration) *ConfirmEngine {
    return &ConfirmEngine{
        ai:      aiProvider,
        writer:  writer,
        timeout: timeout,
        pending: make(map[string]*pendingItem),
    }
}

func (e *ConfirmEngine) key(msg adapter.Message) string {
    return msg.ChannelID + ":" + msg.UserID
}

func (e *ConfirmEngine) HandleMessage(ctx context.Context, msg adapter.Message) (string, error) {
    draft, err := e.ai.GenerateSpec(ctx, msg.Text)
    if err != nil {
        return "", err
    }
    e.mu.Lock()
    e.pending[e.key(msg)] = &pendingItem{
        msg:      msg,
        draft:    draft,
        expireAt: time.Now().Add(e.timeout),
    }
    e.mu.Unlock()
    return fmt.Sprintf("以下是我理解的 spec，回覆 ok 確認 / no 捨棄：\n\n%s", draft), nil
}

func (e *ConfirmEngine) HandleConfirm(ctx context.Context, msg adapter.Message) (string, bool, error) {
    k := e.key(msg)
    e.mu.Lock()
    item, ok := e.pending[k]
    if ok {
        delete(e.pending, k)
    }
    e.mu.Unlock()

    if !ok || time.Now().After(item.expireAt) {
        return "找不到待確認的 spec，請重新發送需求。", false, nil
    }

    text := strings.TrimSpace(strings.ToLower(msg.Text))
    if text != "ok" {
        return "已捨棄。請重新描述需求。", false, nil
    }

    path, err := e.writer.Write(item.msg.Source, item.draft)
    if err != nil {
        return "", false, err
    }
    return fmt.Sprintf("✅ spec 已建立：%s", path), true, nil
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
go test ./internal/engine/... -v
```
Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/engine/
git commit -m "feat: confirm state machine for spec approval flow"
```

---

### Task 7: Telegram Adapter

**Files:**
- Create: `internal/adapter/telegram/telegram.go`
- Create: `internal/adapter/telegram/telegram_test.go`

**Interfaces:**
- Consumes: `adapter.Adapter`（Task 2）、`engine.ConfirmEngine`（Task 6）
- Produces:
  - `telegram.New(token string, engine *engine.ConfirmEngine) adapter.Adapter`

- [ ] **Step 1: 安裝 Telegram SDK**

```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
```

- [ ] **Step 2: 建立 telegram_test.go**

```go
// internal/adapter/telegram/telegram_test.go
package telegram_test

import (
    "testing"
    "github.com/yuying/intake-agent/internal/adapter/telegram"
    "github.com/yuying/intake-agent/internal/engine"
    "github.com/yuying/intake-agent/internal/output"
    "time"
)

type fakeAI struct{}
func (f *fakeAI) Name() string { return "fake" }
func (f *fakeAI) GenerateSpec(_ interface{}, msg string) (string, error) { return "## spec\n"+msg, nil }

func TestTelegramName(t *testing.T) {
    w := output.NewWriter(t.TempDir(), "specs/")
    e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
    a := telegram.New("fake-token", e)
    if a.Name() != "telegram" {
        t.Errorf("expected telegram, got %s", a.Name())
    }
}
```

- [ ] **Step 3: 執行測試確認失敗**

```bash
go test ./internal/adapter/telegram/... -v
```
Expected: compile error

- [ ] **Step 4: 實作 telegram.go**

```go
// internal/adapter/telegram/telegram.go
package telegram

import (
    "context"
    "log"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/yuying/intake-agent/internal/adapter"
    "github.com/yuying/intake-agent/internal/engine"
    "strconv"
    "time"
)

type telegramAdapter struct {
    token  string
    engine *engine.ConfirmEngine
}

func New(token string, eng *engine.ConfirmEngine) adapter.Adapter {
    return &telegramAdapter{token: token, engine: eng}
}

func (t *telegramAdapter) Name() string { return "telegram" }

func (t *telegramAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
    bot, err := tgbotapi.NewBotAPI(t.token)
    if err != nil {
        return err
    }
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := bot.GetUpdatesChan(u)
    for {
        select {
        case <-ctx.Done():
            return nil
        case update := <-updates:
            if update.Message == nil {
                continue
            }
            msg := adapter.Message{
                ID:        strconv.Itoa(update.Message.MessageID),
                Source:    "telegram",
                ChannelID: strconv.FormatInt(update.Message.Chat.ID, 10),
                UserID:    strconv.Itoa(update.Message.From.ID),
                Text:      update.Message.Text,
                Timestamp: time.Unix(int64(update.Message.Date), 0),
            }
            out <- msg
        }
    }
}

func (t *telegramAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
    bot, err := tgbotapi.NewBotAPI(t.token)
    if err != nil {
        return err
    }
    chatID, err := strconv.ParseInt(msg.ChannelID, 10, 64)
    if err != nil {
        return err
    }
    reply := tgbotapi.NewMessage(chatID, text)
    _, err = bot.Send(reply)
    if err != nil {
        log.Printf("telegram reply error: %v", err)
    }
    return err
}
```

- [ ] **Step 5: 執行測試確認通過**

```bash
go test ./internal/adapter/telegram/... -v
```
Expected: `PASS`

- [ ] **Step 6: Commit**

```bash
git add internal/adapter/telegram/ go.mod go.sum
git commit -m "feat: Telegram adapter"
```

---

### Task 8: Slack Adapter

**Files:**
- Create: `internal/adapter/slack/slack.go`
- Create: `internal/adapter/slack/slack_test.go`

**Interfaces:**
- Consumes: `adapter.Adapter`（Task 2）
- Produces:
  - `slack.New(signingSecret, botToken string, eng *engine.ConfirmEngine) adapter.Adapter`
  - HTTP handler 掛在 `/slack/events`

- [ ] **Step 1: 安裝 Slack SDK**

```bash
go get github.com/slack-go/slack
```

- [ ] **Step 2: 建立 slack_test.go**

```go
// internal/adapter/slack/slack_test.go
package slack_test

import (
    "testing"
    "time"
    slackadapter "github.com/yuying/intake-agent/internal/adapter/slack"
    "github.com/yuying/intake-agent/internal/engine"
    "github.com/yuying/intake-agent/internal/output"
)

type fakeAI struct{}
func (f *fakeAI) Name() string { return "fake" }
func (f *fakeAI) GenerateSpec(_ interface{}, msg string) (string, error) { return "## spec\n"+msg, nil }

func TestSlackName(t *testing.T) {
    w := output.NewWriter(t.TempDir(), "specs/")
    e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
    a := slackadapter.New("secret", "bot-token", e)
    if a.Name() != "slack" {
        t.Errorf("expected slack, got %s", a.Name())
    }
}
```

- [ ] **Step 3: 實作 slack.go**

```go
// internal/adapter/slack/slack.go
package slack

import (
    "context"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "github.com/slack-go/slack"
    "github.com/slack-go/slack/slackevents"
    "github.com/yuying/intake-agent/internal/adapter"
    "github.com/yuying/intake-agent/internal/engine"
    "time"
)

type slackAdapter struct {
    signingSecret string
    botToken      string
    engine        *engine.ConfirmEngine
    out           chan<- adapter.Message
}

func New(signingSecret, botToken string, eng *engine.ConfirmEngine) adapter.Adapter {
    return &slackAdapter{signingSecret: signingSecret, botToken: botToken, engine: eng}
}

func (s *slackAdapter) Name() string { return "slack" }

func (s *slackAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
    s.out = out
    http.HandleFunc("/slack/events", s.handleEvent)
    return nil
}

func (s *slackAdapter) handleEvent(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    sv, err := slack.NewSecretsVerifier(r.Header, s.signingSecret)
    if err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    sv.Write(body)
    if err := sv.Ensure(); err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    if eventsAPIEvent.Type == slackevents.URLVerification {
        var cr *slackevents.ChallengeResponse
        json.Unmarshal(body, &cr)
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte(cr.Challenge))
        return
    }
    if eventsAPIEvent.Type == slackevents.CallbackEvent {
        innerEvent := eventsAPIEvent.InnerEvent
        if msg, ok := innerEvent.Data.(*slackevents.MessageEvent); ok {
            if s.out != nil {
                s.out <- adapter.Message{
                    ID:        msg.TimeStamp,
                    Source:    "slack",
                    ChannelID: msg.Channel,
                    UserID:    msg.User,
                    Text:      msg.Text,
                    Timestamp: time.Now(),
                }
            }
        }
    }
    w.WriteHeader(http.StatusOK)
}

func (s *slackAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
    api := slack.New(s.botToken)
    _, _, err := api.PostMessage(msg.ChannelID, slack.MsgOptionText(text, false))
    if err != nil {
        log.Printf("slack reply error: %v", err)
    }
    return err
}
```

- [ ] **Step 4: 執行測試確認通過**

```bash
go test ./internal/adapter/slack/... -v
```
Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/slack/ go.mod go.sum
git commit -m "feat: Slack adapter with Events API webhook"
```

---

### Task 9: Core Engine + HTTP Server 整合（main.go）

**Files:**
- Modify: `cmd/intake-agent/main.go`
- Create: `internal/engine/engine.go`

**Interfaces:**
- Consumes: 全部前面 Tasks 的產出

- [ ] **Step 1: 實作 engine.go（整合 adapters）**

```go
// internal/engine/engine.go
package engine

import (
    "context"
    "log"
    "github.com/yuying/intake-agent/internal/adapter"
)

type Engine struct {
    adapters []adapter.Adapter
    confirm  *ConfirmEngine
}

func NewEngine(confirm *ConfirmEngine, adapters ...adapter.Adapter) *Engine {
    return &Engine{confirm: confirm, adapters: adapters}
}

func (e *Engine) Run(ctx context.Context) error {
    out := make(chan adapter.Message, 100)
    for _, a := range e.adapters {
        go func(a adapter.Adapter) {
            if err := a.Start(ctx, out); err != nil {
                log.Printf("adapter %s error: %v", a.Name(), err)
            }
        }(a)
    }
    for {
        select {
        case <-ctx.Done():
            return nil
        case msg := <-out:
            go e.handleMsg(ctx, msg)
        }
    }
}

func (e *Engine) handleMsg(ctx context.Context, msg adapter.Message) {
    var replyText string
    var err error

    // 先嘗試當成確認訊息處理
    replyText, wrote, _ := e.confirm.HandleConfirm(ctx, msg)
    if wrote || replyText != "" {
        // 找到對應的 adapter 回覆
        for _, a := range e.adapters {
            if a.Name() == msg.Source {
                a.Reply(ctx, msg, replyText)
                return
            }
        }
        return
    }

    // 否則當成新需求
    replyText, err = e.confirm.HandleMessage(ctx, msg)
    if err != nil {
        log.Printf("engine error: %v", err)
        return
    }
    for _, a := range e.adapters {
        if a.Name() == msg.Source {
            a.Reply(ctx, msg, replyText)
            return
        }
    }
}
```

- [ ] **Step 2: 更新 main.go**

```go
// cmd/intake-agent/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    "github.com/yuying/intake-agent/internal/ai"
    "github.com/yuying/intake-agent/internal/config"
    "github.com/yuying/intake-agent/internal/engine"
    "github.com/yuying/intake-agent/internal/output"
    telegramadapter "github.com/yuying/intake-agent/internal/adapter/telegram"
    slackadapter "github.com/yuying/intake-agent/internal/adapter/slack"
)

func main() {
    configPath := flag.String("config", "configs/config.yaml", "path to config file")
    flag.Parse()

    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    apiKey := os.Getenv("AI_API_KEY")
    aiProvider, err := ai.New(cfg.AI.Provider, cfg.AI.Model, apiKey)
    if err != nil {
        log.Fatalf("failed to create AI provider: %v", err)
    }

    writer := output.NewWriter(cfg.Output.RepoPath, cfg.Output.Dir)
    confirm := engine.NewConfirmEngine(aiProvider, writer, 600*time.Second)

    var adapters []adapter.Adapter
    if cfg.Adapters.Telegram.Enabled {
        adapters = append(adapters, telegramadapter.New(cfg.Adapters.Telegram.Token, confirm))
    }
    if cfg.Adapters.Slack.Enabled {
        adapters = append(adapters, slackadapter.New(cfg.Adapters.Slack.SigningSecret, cfg.Adapters.Slack.BotToken, confirm))
    }

    eng := engine.NewEngine(confirm, adapters...)

    log.Printf("intake-agent starting on :%d (AI: %s)", cfg.Server.Port, aiProvider.Name())
    go func() {
        if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil); err != nil {
            log.Fatalf("http server error: %v", err)
        }
    }()

    if err := eng.Run(context.Background()); err != nil {
        log.Fatalf("engine error: %v", err)
    }
}
```

- [ ] **Step 3: 確認 build 通過**

```bash
go build ./...
```
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add cmd/ internal/engine/engine.go
git commit -m "feat: core engine and HTTP server wiring in main.go"
```

---

### Task 10: Docker + 部署文件

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `configs/config.yaml.example`（更新）
- Create: `README.md`

- [ ] **Step 1: 建立 Dockerfile**

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/intake-agent cmd/intake-agent/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/bin/intake-agent .
COPY configs/ configs/
EXPOSE 8080
ENTRYPOINT ["./intake-agent", "--config", "configs/config.yaml"]
```

- [ ] **Step 2: 建立 docker-compose.yml**

```yaml
# docker-compose.yml
services:
  intake-agent:
    build: .
    ports:
      - "8080:8080"
    environment:
      - AI_API_KEY=${AI_API_KEY}
    volumes:
      - ./configs/config.yaml:/app/configs/config.yaml:ro
      - ${SPECS_REPO_PATH:-./specs}:/specs
    restart: unless-stopped
```

- [ ] **Step 3: 建立 README.md**

```markdown
# intake-agent

多管道訊息收集與 AI spec 自動產生工具。

## 快速開始

1. 複製設定檔
   ```bash
   cp configs/config.yaml.example configs/config.yaml
   ```
2. 填入 token（Telegram bot token、Slack signing secret + bot token）
3. 設定 AI API key
   ```bash
   export AI_API_KEY=your_api_key
   ```
4. 啟動
   ```bash
   make start
   ```

## 本機開發（需要 public URL）

```bash
ngrok http 8080
# 將 ngrok URL 設定為 Telegram webhook 或 Slack Events URL
```

## 雲端部署

```bash
docker-compose up -d
```

## AI Provider 切換

在 `configs/config.yaml` 修改：
```yaml
ai:
  provider: gemini   # claude | gemini | codex
  model: gemini-2.0-flash
```

並設定對應的 `AI_API_KEY`。
```

- [ ] **Step 4: 確認 Docker build**

```bash
docker build -t intake-agent . 2>&1 | tail -5
```
Expected: `Successfully built ...`

- [ ] **Step 5: 執行所有測試**

```bash
go test ./... -v
```
Expected: 全部 `PASS`

- [ ] **Step 6: Commit**

```bash
git add Dockerfile docker-compose.yml README.md
git commit -m "feat: Docker deployment and README"
```

---

## 任務總覽

| Task | 內容 | 依賴 |
|------|------|------|
| 1 | 專案骨架 + Config | — |
| 2 | Adapter 介面 + Message | 1 |
| 3 | AIProvider 介面 + Claude | 1 |
| 4 | Gemini + Codex + Factory | 3 |
| 5 | Output Writer | 1 |
| 6 | 確認狀態機 | 2, 3, 5 |
| 7 | Telegram Adapter | 2, 6 |
| 8 | Slack Adapter | 2, 6 |
| 9 | Core Engine + main.go 整合 | 全部 |
| 10 | Docker + 部署文件 | 9 |
