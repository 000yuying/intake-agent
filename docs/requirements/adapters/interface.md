# Adapter 介面規範

## 介面定義

```go
// internal/adapter/adapter.go

type Message struct {
    ID        string
    Source    string    // "telegram" | "slack" | "discord" | ...
    ChannelID string    // 用於回覆與 pending key
    UserID    string    // 用於 pending key
    Text      string
    Timestamp time.Time
}

type Adapter interface {
    Name() string
    Start(ctx context.Context, out chan<- Message) error
    Reply(ctx context.Context, msg Message, text string) error
}
```

## 各方法規範

### `Name() string`
- 回傳小寫固定字串，如 `"telegram"`、`"slack"`
- Engine 用此值比對 `msg.Source` 來決定回哪個 adapter

### `Start(ctx, out chan<- Message) error`
- 開始監聽訊息，收到後推入 `out` channel
- **必須非阻塞推送**：用 `select/default` 避免 HTTP handler 或 goroutine 卡住

```go
select {
case out <- msg:
default:
    log.Printf("adapter %s: out channel full, dropping message", a.Name())
}
```

- `ctx.Done()` 觸發時 graceful 結束，回傳 `nil`
- 不在此方法呼叫 AI 或寫檔

### `Reply(ctx, msg Message, text string) error`
- 回覆文字到 `msg.ChannelID` 指定的頻道
- 若 API 呼叫失敗，log 後回傳 error；Engine 會忽略 reply error（不中斷主流程）
- **不得儲存 bot 連線狀態到 Reply**（避免重複建立連線的問題可在 struct 初始化時處理）

## 規範限制

- Adapter 不得 import `internal/engine`（防止循環依賴）
- Adapter 不得呼叫 `AIProvider`
- Adapter 不得直接呼叫 `output.Writer`
- Adapter 的測試只需驗證 `Name()` 和基本 HTTP handler 行為（無需真實 API token）
