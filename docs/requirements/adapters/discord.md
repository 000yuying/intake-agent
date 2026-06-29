# Discord Adapter

## 實作方式

WebSocket Gateway（discordgo），不需 public URL。

## 設定

```yaml
adapters:
  discord:
    enabled: true
    token: "your-bot-token"
```

## 建立 Discord Bot

1. 前往 https://discord.com/developers/applications → New Application
2. 左側 Bot → Add Bot → 複製 Token
3. OAuth2 → URL Generator：勾選 `bot`，Bot Permissions 勾選 `Send Messages`、`Read Message History`
4. 用產生的 URL 邀請 Bot 進入 Server

## 訊息過濾規則

- `m.Author == nil`：略過
- `m.Author.Bot == true`：略過（防止 bot 迴圈）

## 已知限制

- 需要在 Discord Developer Portal 啟用 Message Content Intent（Privileged Gateway Intents）
- 無重連邏輯，ctx 取消後不會自動重啟
