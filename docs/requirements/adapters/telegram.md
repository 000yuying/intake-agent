# Telegram Adapter

## 實作方式

長輪詢（long-polling），不需 public URL。

## 設定

```yaml
adapters:
  telegram:
    enabled: true
    token: "123456:ABC-..."   # BotFather 取得
```

環境變數：`AI_API_KEY`（AI provider 用，非 Telegram token）

## 取得 Bot Token

1. 開啟 Telegram，搜尋 `@BotFather`
2. 輸入 `/newbot`，依指示命名
3. 複製 token 填入 config

## 本機開發

長輪詢不需 ngrok（Telegram 主動推訊息給 bot，不需 inbound webhook）。

```bash
make start
```

## 訊息過濾規則

- `update.Message == nil`：略過
- `update.Message.From == nil`：略過（頻道轉發訊息無發送者）
- 其他：正常處理

## 已知限制

- `Reply()` 每次呼叫重新建立 BotAPI 連線（可接受，效能非瓶頸）
- 無重連 / backoff 邏輯（網路中斷需手動重啟）
