# Requirements Index

> 先從這裡找對應領域文件，再進入細節。每份文件獨立成立。

## Domain Map

### `architecture/`
- `data-flow.md` — 訊息從進入到寫檔的完整流程與各層邊界

### `adapters/`
- `interface.md` — `Adapter` 介面定義、`Message` struct、實作規範
- `telegram.md` — Telegram Bot API 設定與本地開發指引
- `slack.md` — Slack Events API 設定、signing secret 驗證流程
- `adding-new-adapter.md` — 新增管道的步驟 checklist

### `ai-providers/`
- `interface.md` — `AIProvider` 介面定義、prompt 規範、回應格式
- `claude.md` — Anthropic Claude 設定與 model 選擇
- `gemini.md` — Google Gemini 設定
- `codex.md` — OpenAI Codex / GPT 設定
- `adding-new-provider.md` — 新增 AI provider 的步驟 checklist

### `engine/`
- `confirm-flow.md` — 確認狀態機邏輯、路由規則、逾時行為

### `output/`
- `spec-format.md` — 輸出 Markdown 格式規範、檔名規則

### `deployment/`
- `local.md` — 本機開發、ngrok 設定、環境變數
- `docker.md` — Docker / docker-compose 部署、volume 設定

## Boundary Notes

- Adapter 只負責收發訊息，不呼叫 AI、不寫檔、不維護確認狀態。
- Engine 不 import 任何 adapter package；透過 `msg.Source` 動態查找 adapter。
- AI Provider 只負責產生 spec 文字，不感知來源管道。
- Output Writer 只負責寫檔，不感知 AI 或管道。
