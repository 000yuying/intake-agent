# 新增 Adapter Checklist

新增一個訊息管道需要以下步驟。以 `discord` 為例。

## Step 1：建立 package

```
internal/adapter/discord/
├── discord.go
└── discord_test.go
```

## Step 2：實作 Adapter 介面

```go
package discord

import (
    "context"
    "github.com/yuying/intake-agent/internal/adapter"
)

type discordAdapter struct {
    token string
}

func New(token string) adapter.Adapter {
    return &discordAdapter{token: token}
}

func (d *discordAdapter) Name() string { return "discord" }

func (d *discordAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
    // 監聽訊息，非阻塞推送
    // select { case out <- msg: default: log.Printf(...) }
}

func (d *discordAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
    // 回覆到 msg.ChannelID
}
```

## Step 3：寫測試

至少驗證：
- `Name()` 回傳正確字串
- HTTP handler / event parsing 基本行為（可用 fake request）

## Step 4：加入 Config

`internal/config/config.go` 的 `AdaptersConfig` 新增：

```go
Discord DiscordConfig `yaml:"discord"`

type DiscordConfig struct {
    Enabled bool   `yaml:"enabled"`
    Token   string `yaml:"token"`
}
```

`configs/config.yaml.example` 新增對應欄位。

## Step 5：接入 main.go

```go
if cfg.Adapters.Discord.Enabled {
    adapters = append(adapters, discordadapter.New(cfg.Adapters.Discord.Token))
}
```

## Step 6：更新文件

- 在 `docs/requirements/adapters/` 新增 `discord.md`，說明設定方式
- 更新 `docs/requirements/README.md` 的 domain map
- 更新 `docs/PROJECT_MASTER_PLAN.md` 的功能模組藍圖狀態

## 常見錯誤

| 錯誤 | 正確做法 |
| :--- | :--- |
| adapter import engine package | 禁止，會循環依賴 |
| 在 Reply() 裡建立新連線 | 在 struct 初始化時建立，複用 |
| 同步推送到 out channel | 一律 select/default 非阻塞 |
| 處理 bot 自己發的訊息 | 過濾 bot_id 或 sub_type 非空的訊息 |
