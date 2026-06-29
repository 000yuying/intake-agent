# 資料流：訊息進入到 spec 寫檔

## 完整流程

```
1. 使用者在 Telegram / Slack 發送需求訊息
        │
        ▼
2. Adapter.Start() 接收，轉為 adapter.Message
   { ID, Source, ChannelID, UserID, Text, Timestamp }
        │
        ▼ (寫入 chan<- Message，buffer=100)
3. Engine.Run() 從 channel 讀取，spawn goroutine 處理
        │
        ▼
4. Engine.handleMsg() 路由判斷：
   - msg.Text == "ok" 或 "no" → HandleConfirm()
   - 其他 → HandleMessage()
        │
   ┌────┴────────────────────────────────┐
   │ HandleMessage()                     │ HandleConfirm()
   │  → 呼叫 AIProvider.GenerateSpec()   │  → 查 pending map
   │  → 存入 pending map (key=            │  → "ok" → Output.Write()
   │    ChannelID:UserID)                │  → "no" → 刪除 pending
   │  → 回覆草稿 + "ok/no 確認"          │  → 逾時 → 回覆"找不到"
   └────┬────────────────────────────────┘
        │
        ▼
5. Adapter.Reply() 回原管道
        │
        ▼ (使用者回 "ok")
6. Output.Write() 寫入 Markdown
   → 路徑：{repo_path}/{dir}/YYYY-MM-DD-HH-MM-SS-{source}.md
        │
        ▼
7. Adapter.Reply() 回覆檔案路徑給使用者
```

## 關鍵邊界

| 邊界 | 說明 |
| :--- | :--- |
| Adapter → Engine | `chan<- adapter.Message`，buffer 100，非阻塞（select/default） |
| Engine → AI | `AIProvider.GenerateSpec(ctx, text)` 介面呼叫 |
| Engine → Output | `Writer.Write(source, content)` 介面呼叫 |
| Engine → Adapter | 透過 `msg.Source` 比對 `Adapter.Name()` 找到對應 adapter 回覆 |

## 確認狀態 pending map

```
key:   ChannelID + ":" + UserID
value: { msg, draft, expireAt }
```

- 同一 ChannelID + UserID 只能有一筆 pending
- 新需求覆蓋舊 pending（使用者重新描述時自然更新）
- 逾時 10 分鐘（預設），逾時後回覆「找不到待確認的 spec」
