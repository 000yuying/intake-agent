# Notion Adapter

## 實作方式

輪詢（Polling）Notion Database，不需 public URL。每隔 `poll_interval_seconds` 秒查詢資料庫，比對已處理過的 page ID，僅將新頁面送入處理流程。

## 設定

```yaml
adapters:
  notion:
    enabled: true
    token: "secret_..."
    database_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    poll_interval_seconds: 60
```

## 建立 Notion Integration

1. https://www.notion.so/my-integrations → New integration
2. 給予名稱，複製 Integration Token（`secret_...`）
3. 前往目標 Database → 右上角 `...` → Connections → 選擇你的 integration
4. Database URL 中取得 Database ID（`notion.so/<DATABASE_ID>?v=...`）

## 輪詢邏輯

- 每次查詢取最新 10 筆頁面
- 已見過的 page ID 存於記憶體（`seenIDs` map），重啟後重新處理未確認的頁面
- 若資料庫頁面量很大，建議用 Notion filter 限制查詢範圍

## 與 Webhook 型 Adapter 的差異

- 無法即時接收，最多延遲 `poll_interval_seconds` 秒
- 不需 public URL（適合本機長時間運行）
- `seenIDs` 存在記憶體，服務重啟後舊頁面可能重複處理

## ChannelID 格式

Notion Page ID（UUID），例如 `12345678-1234-1234-1234-123456789abc`
