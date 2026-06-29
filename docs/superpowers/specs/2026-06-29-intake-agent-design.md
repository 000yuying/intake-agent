# intake-agent Design Doc

**Date:** 2026-06-29  
**Status:** Approved  
**Author:** yuying

---

## 概覽

`intake-agent` 是一個 AI 最大化工作流程的訊息收集與 spec 自動產生工具。從多個訊息管道（Telegram、Slack、Discord、Jira、Google Chat、Notion、GitHub Issues）讀取需求訊息，透過 Claude AI 產生 spec 草稿，經人工在原管道輕量確認後，寫入 Markdown 檔案至指定 git repo。

---

## 目標

- 從多管道自動收集需求訊息
- AI 解析意圖並產出標準化 spec Markdown
- 人工在原管道回覆「ok」即可確認，低摩擦
- Docker-ready，本機開發後可直接部署至雲端

---

## 架構

```
Telegram / Slack / Discord / Jira / G.Chat / Notion / GitHub
        │
        ▼
┌───────────────┐
│   Adapters    │  統一轉換為內部 Message struct
└───────┬───────┘
        │
        ▼
┌───────────────┐
│  Core Engine  │  AI 產草稿 → 回原管道確認 → 收到 ok → 寫檔
└───────┬───────┘
        │
        ▼
┌───────────────┐
│    Output     │  寫入 Markdown 至指定 repo
└───────────────┘
```

---

## 核心資料結構

```go
type Message struct {
    ID        string
    Source    string    // "telegram" | "slack" | "discord" | ...
    ChannelID string    // 用來回覆確認訊息
    UserID    string
    Text      string
    Timestamp time.Time
}
```

---

## Adapter 介面

```go
type Adapter interface {
    Name() string
    Start(ctx context.Context, out chan<- Message) error
    Reply(ctx context.Context, msg Message, text string) error
}
```

每個管道實作此介面，Core Engine 與管道實作完全解耦。新增管道只需新增一個 adapter package。

---

## 確認狀態機

```
收到 Message
  → Core Engine 呼叫 Claude API 產 spec 草稿
  → 透過 Adapter.Reply() 回原管道：
    「這是我理解的 spec，回覆 ok 確認 / no 捨棄」
  → 等待同一 ChannelID + UserID 回覆
    → "ok"  → 寫入 Markdown → 回覆檔案路徑
    → "no"  → 捨棄，等待下一則訊息
    → 逾時（預設 10 分鐘）→ 捨棄
```

---

## Config（YAML）

```yaml
server:
  port: 8080

ai:
  provider: claude
  model: claude-sonnet-4-6

output:
  repo_path: /home/yuying/specs   # spec 寫入的 git repo 路徑
  dir: specs/                     # repo 內的子目錄

adapters:
  telegram:
    enabled: true
    token: ""
  slack:
    enabled: true
    signing_secret: ""
    bot_token: ""
  discord:
    enabled: false
    token: ""
  jira:
    enabled: false
    host: ""
    email: ""
    api_token: ""
  gchat:
    enabled: false
    # Google Chat 使用 webhook，不需 token
  notion:
    enabled: false
    token: ""
    database_id: ""
  github:
    enabled: false
    token: ""
    repo: ""  # owner/repo
```

---

## 目錄結構

```
intake-agent/
├── cmd/
│   └── intake-agent/
│       └── main.go
├── internal/
│   ├── adapter/
│   │   ├── adapter.go        # Adapter interface + Message struct
│   │   ├── telegram/
│   │   ├── slack/
│   │   ├── discord/
│   │   ├── jira/
│   │   ├── gchat/
│   │   ├── notion/
│   │   └── github/
│   ├── engine/
│   │   ├── engine.go         # Core Engine：協調 adapter → AI → output
│   │   └── confirm.go        # 確認狀態機
│   ├── ai/
│   │   └── claude.go         # Anthropic Claude API 呼叫
│   └── output/
│       └── writer.go         # Markdown 寫入 + 檔名產生
├── configs/
│   └── config.yaml.example
├── docs/
│   └── superpowers/specs/
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

---

## MVP 範圍（第一版）

| 項目 | 狀態 |
|------|------|
| Telegram adapter | ✅ 實作 |
| Slack adapter | ✅ 實作 |
| Discord adapter | 介面預留，實作留後 |
| Jira adapter | 介面預留，實作留後 |
| Google Chat adapter | 介面預留，實作留後 |
| Notion adapter | 介面預留，實作留後 |
| GitHub Issues adapter | 介面預留，實作留後 |
| Core Engine + 確認狀態機 | ✅ 實作 |
| Claude API 產 spec | ✅ 實作 |
| Markdown 寫入本機 repo | ✅ 實作 |
| Dockerfile + docker-compose | ✅ 實作 |
| ngrok 本機開發指引 | ✅ 文件 |

---

## 部署

**本機開發**
```bash
# 安裝 ngrok，開 public URL
ngrok http 8080

# 啟動服務
make start
```

**雲端**
```bash
docker-compose up -d
```

---

## 非目標（Not in scope）

- Web UI / Dashboard
- 多人確認流程（目前只支援原始發訊者確認）
- spec 版本管理（由 git 負責）
- 自動執行 spec（交由 agent-flow 負責）
