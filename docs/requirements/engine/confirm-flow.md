# 確認狀態機：ConfirmEngine

## 狀態流程

```
使用者發訊息
    │
    ▼
Engine.handleMsg()
    │
    ├─ Text == "ok" 或 "no" ──→ HandleConfirm()
    │                               │
    │                    ┌──────────┼──────────────┐
    │                    │          │               │
    │                 有 pending  Text=="ok"    Text=="no"
    │                 且未逾時       │               │
    │                    │       Write()         刪除 pending
    │                    │       回覆路徑         回覆"已捨棄"
    │                    │
    │                 無 pending / 已逾時
    │                    │
    │                 回覆"找不到待確認..."
    │
    └─ 其他文字 ──→ HandleMessage()
                        │
                     GenerateSpec()
                        │
                     存入 pending map
                        │
                     回覆草稿 + "ok/no" 指示
```

## Pending Map

```go
key:   msg.ChannelID + ":" + msg.UserID
value: { msg, draft, expireAt }
```

- **同一 key 只能有一筆 pending**：使用者重新描述需求時，新 pending 自動覆蓋舊的
- **逾時**：預設 600 秒（10 分鐘），由 `NewConfirmEngine` 的 `timeout` 參數決定
- **thread-safe**：所有 map 操作都在 `sync.Mutex` 保護下執行

## 路由規則（重要）

路由依據是**訊息文字內容**，不是「有無 pending item」：

| 訊息文字（trim + lowercase） | 路由 |
| :--- | :--- |
| `"ok"` | HandleConfirm |
| `"no"` | HandleConfirm |
| 其他 | HandleMessage |

這樣設計的原因：若以「有無 pending」作為路由依據，HandleConfirm 即使找不到 pending 也會回覆「找不到...」，導致正常需求訊息永遠無法進入 HandleMessage。

## 逾時行為

逾時後使用者再回 `"ok"`：
- 回覆：`"找不到待確認的 spec，請重新發送需求。"`
- `wrote = false`

## 方法簽名

```go
func (e *ConfirmEngine) HandleMessage(ctx context.Context, msg adapter.Message) (replyText string, err error)
func (e *ConfirmEngine) HandleConfirm(ctx context.Context, msg adapter.Message) (replyText string, wrote bool, err error)
```
