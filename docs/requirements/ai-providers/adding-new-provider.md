# 新增 AI Provider Checklist

以新增 `mistral` 為例。

## Step 1：建立實作檔

```
internal/ai/mistral.go
internal/ai/mistral_test.go
```

## Step 2：實作 AIProvider 介面

```go
package ai

import (
    "context"
    "errors"
)

type mistralProvider struct {
    model  string
    apiKey string
}

func NewMistral(model, apiKey string) AIProvider {
    return &mistralProvider{model: model, apiKey: apiKey}
}

func (m *mistralProvider) Name() string { return "mistral" }

func (m *mistralProvider) GenerateSpec(ctx context.Context, userMessage string) (string, error) {
    if m.apiKey == "" {
        return "", errors.New("api key is required")
    }
    // 呼叫 Mistral API，使用 specPrompt(userMessage)
}
```

## Step 3：寫測試

必須覆蓋：
- `Name()` 回傳 `"mistral"`
- 空 key 回傳含 `"api key"` 的 error

## Step 4：加入 factory

`internal/ai/factory.go`：

```go
case "mistral":
    return NewMistral(model, apiKey), nil
```

## Step 5：更新文件

- 在 `docs/requirements/ai-providers/` 新增 `mistral.md`
- 更新 `docs/requirements/README.md`
- 更新 `configs/config.yaml.example` 的 provider 欄位說明

## 規範限制

- `specPrompt()` helper 不得複製，直接呼叫 `claude.go` 中定義的版本（package 內可見）
- `GenerateSpec` 回傳值必須是純字串，不含任何包裝
- 不得修改 `Engine` 或 `ConfirmEngine` 的任何程式碼
