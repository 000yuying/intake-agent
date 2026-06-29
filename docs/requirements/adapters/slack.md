# Slack Adapter

## 實作方式

Events API webhook，需要 public URL（本機用 ngrok）。

## 設定

```yaml
adapters:
  slack:
    enabled: true
    signing_secret: "abc123..."   # Slack App Basic Information
    bot_token: "xoxb-..."         # OAuth & Permissions → Bot Token
```

## 建立 Slack App

1. 前往 https://api.slack.com/apps → Create New App → From scratch
2. **OAuth & Permissions**：新增 Bot Token Scopes
   - `chat:write`（發送訊息）
   - `channels:history`（讀取訊息）
3. **Event Subscriptions**：啟用，填入 webhook URL
   - URL 格式：`https://<your-domain>/slack/events`
   - Subscribe to bot events：`message.channels`
4. 安裝 App 到 workspace，複製 Bot Token
5. 複製 **Signing Secret**（Basic Information 頁面）

## 本機開發（ngrok）

```bash
# 開啟 public URL
ngrok http 8080

# 將 ngrok URL 填入 Slack App Event Subscriptions
# https://xxxx.ngrok.io/slack/events
```

## Webhook 驗證流程

每個 request 都會驗證 Slack signing secret：

```
X-Slack-Request-Timestamp + body → HMAC-SHA256 → 比對 X-Slack-Signature
```

時間差超過 5 分鐘的請求會被拒絕（防 replay attack）。

## 訊息過濾規則

以下訊息會被略過（避免 bot 自我迴圈）：
- `SubType != ""`：系統訊息（如 bot 加入頻道）
- `BotID != ""`：bot 自己發的訊息

## URL 驗證 Challenge

Slack 第一次設定 webhook 時會發送 challenge request，handler 自動處理並回傳 `challenge` 欄位。
