# Google Chat Adapter

## 實作方式

HTTP webhook 接收訊息，透過 Google Chat App Webhook URL 回覆。

## 設定

```yaml
adapters:
  gchat:
    enabled: true
    webhook_url: "https://chat.googleapis.com/v1/spaces/xxx/messages?key=yyy&token=zzz"
```

## 建立 Google Chat App

1. 前往 Google Chat API → 設定 → 建立應用程式
2. 連線設定選 HTTP endpoint，填入 webhook URL（需 public URL，本機開發用 ngrok）
   - URL 格式：`https://<your-domain>/gchat/events`
3. 啟用後，從 Google Chat 空間的 Apps 取得 Webhook URL（用於回覆）

## ChannelID 格式

`spaces/<space-name>`，例如 `spaces/AAAAbc123`

## 已知限制

- `Reply()` 透過固定 webhook_url 回覆，所有空間共用同一個 webhook
- 無簽名驗證（Google Chat 不提供 HMAC 方式）
