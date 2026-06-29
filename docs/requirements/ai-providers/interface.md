# AIProvider 介面規範

## 介面定義

```go
// internal/ai/ai.go

type AIProvider interface {
    Name() string
    GenerateSpec(ctx context.Context, userMessage string) (string, error)
}
```

## 各方法規範

### `Name() string`
回傳小寫固定字串：`"claude"`、`"gemini"`、`"codex"`

### `GenerateSpec(ctx, userMessage string) (string, error)`
- 輸入：使用者的原始訊息文字
- 輸出：可直接寫入 Markdown 的 spec 字串
- 空 API key 必須回傳含 `"api key"` 的 error（讓啟動失敗快速顯現）
- 回應必須是純字串，不得含 JSON 包裝或結構體序列化

## Prompt 規範

所有 provider 使用同一個 `specPrompt()` helper（定義在 `internal/ai/claude.go`）：

```
你是一個需求分析師。根據以下訊息，產出一份簡潔的 spec Markdown。

格式：
## 需求概述
（一段話說明這個需求是什麼）

## 驗收條件
- （條列式 AC，每條以明確的可測量標準描述）

## 範圍外
- （不在本次範圍的相關項目）

---
訊息：{userMessage}
```

## 切換 Provider

只需修改 config：

```yaml
ai:
  provider: gemini          # claude | gemini | codex
  model: gemini-2.0-flash
```

並設定對應的 `AI_API_KEY` 環境變數。

## Factory

```go
// internal/ai/factory.go
func New(provider, model, apiKey string) (AIProvider, error)
```

`main.go` 透過此函式建立 provider，不直接 import 各 provider package。
