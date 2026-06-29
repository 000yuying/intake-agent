# 專案全貌：intake-agent

> **核心願景**：將散落在各訊息管道（Telegram、Slack、Discord、Jira 等）的需求訊息，透過 AI 自動轉化為標準化 spec Markdown，並在原管道輕量確認後寫入 git repo，最大化 AI 工作流程的輸入品質。

---

## 1. 系統架構概要

```
訊息來源                    核心層                    輸出
─────────────────────────────────────────────────────────
Telegram  ──┐
Slack     ──┤   Adapters   →   Engine   →   Output Writer
Discord   ──┤   (統一介面)      ↕                ↓
Jira      ──┘              AI Provider     spec .md 檔案
G.Chat    ──                (可抽換)        寫入 git repo
Notion    ──
GitHub    ──
```

**三層職責：**

| 層 | 職責 | 可擴充性 |
| :--- | :--- | :--- |
| **Adapters** | 各管道訊息接收與回覆，統一轉為 `Message` struct | 新增管道只需實作 `Adapter` 介面 |
| **Engine** | 協調 AI 產草稿、確認狀態機、寫檔 | 核心邏輯不感知管道或 AI 實作 |
| **AI Provider** | 呼叫 Claude / Gemini / Codex 產 spec | config 切換，不改程式碼 |

---

## 2. 功能模組藍圖

### A. 訊息接收（MVP 已實作）
- **Telegram adapter**：長輪詢方式接收訊息，webhook 回覆
- **Slack adapter**：Events API webhook，signing secret 驗證

### B. 訊息接收（介面預留，待實作）
- Discord、Jira、Google Chat、Notion、GitHub Issues

### C. AI 產 spec（已實作）
- Claude（`claude-sonnet-4-6` 預設）
- Gemini（`gemini-2.0-flash`）
- Codex / OpenAI（`gpt-4o`）
- 統一 prompt：產出「需求概述 + 驗收條件 + 範圍外」三段式 Markdown

### D. 確認流程（已實作）
- 收到訊息 → AI 草稿 → 回原管道等 `ok`/`no`
- 逾時 10 分鐘自動捨棄
- 確認後寫入 `YYYY-MM-DD-HH-MM-SS-<source>.md`

### E. 部署
- 本機：`make start` + ngrok
- 雲端：`docker-compose up -d`

---

## 3. 核心開發原則

1. **介面隔離**：`Adapter` 和 `AIProvider` 是唯一兩個跨層介面。任何新管道或新 AI 只需實作對應介面，不得修改 Engine。

2. **Engine 不感知管道**：Engine 透過 `msg.Source` 找對應 adapter 回覆，不直接 import 任何 adapter package。

3. **確認路由規則**：訊息文字為 `ok` 或 `no`（大小寫不限）才進確認流程；其他所有文字視為新需求。不得用「有無 pending item」作為路由依據。

4. **AI 回應必須是純字串**：各 provider 的 `GenerateSpec` 必須回傳可直接寫入 Markdown 的字串，不得含有格式包裝（如 JSON、struct 序列化）。

5. **Webhook 必須即時回應**：Slack 要求 3 秒內回應，channel send 必須用非阻塞 `select/default`，不得同步等待 Engine 處理完畢。

6. **空 API Key 必須在建構時報錯**：provider 的 `GenerateSpec` 若 key 為空需回傳含 `"api key"` 的 error，讓啟動時失敗快速顯現。

---

## 4. 資料夾結構速覽

```
intake-agent/
├── cmd/intake-agent/main.go       # 程式進入點：讀 config、組裝所有元件
├── internal/
│   ├── adapter/                   # 介面定義 + 各管道實作
│   ├── ai/                        # AIProvider 介面 + Claude/Gemini/Codex
│   ├── config/                    # YAML config 載入
│   ├── engine/                    # ConfirmEngine（狀態機）+ Engine（協調器）
│   └── output/                    # Markdown 寫入
├── configs/config.yaml.example    # 設定範本
├── docs/                          # 本目錄：系統文檔
├── Dockerfile / docker-compose.yml
└── Makefile
```
