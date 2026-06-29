# CLAUDE.md

本專案的設計規範與需求文件存於 `docs/` 目錄，請按任務需求讀取對應文件。

## 規範導覽

| 領域 | 路徑 | 核心內容 |
| :--- | :--- | :--- |
| **專案全貌** | `docs/PROJECT_MASTER_PLAN.md` | 系統架構、功能模組、核心開發原則 |
| **需求索引** | `docs/requirements/README.md` | 所有領域文件的唯一入口 |
| **資料流** | `docs/requirements/architecture/data-flow.md` | 訊息進入 → AI 草稿 → 確認 → 寫檔的完整流程 |
| **Adapter 介面** | `docs/requirements/adapters/interface.md` | Adapter 規範、禁止事項 |
| **AIProvider 介面** | `docs/requirements/ai-providers/interface.md` | 介面規範、prompt 格式、切換方式 |
| **確認狀態機** | `docs/requirements/engine/confirm-flow.md` | ok/no 路由規則、逾時行為 |
| **Spec 輸出格式** | `docs/requirements/output/spec-format.md` | 檔名格式、路徑規則 |
| **新增 Adapter** | `docs/requirements/adapters/adding-new-adapter.md` | 新增管道 checklist |
| **新增 AI Provider** | `docs/requirements/ai-providers/adding-new-provider.md` | 新增 AI provider checklist |
| **本機開發** | `docs/requirements/deployment/local.md` | 環境變數、ngrok、Makefile 指令 |
| **Docker 部署** | `docs/requirements/deployment/docker.md` | Volume 設定、config 注意事項 |

## 快速參考

- **專案：** intake-agent（多管道訊息收集 + AI spec 自動產生）
- **模組名：** `github.com/yuying/intake-agent`
- **服務端口：** `:8080`（可由 `server.port` 調整）
- **主要指令：**
  ```bash
  make start              # 啟動服務（預設 configs/config.yaml）
  make test               # 執行所有測試
  make build              # 編譯為 binary
  go test ./...           # 直接執行測試
  go build ./...          # 建置驗證
  ```
- **AI API Key：** 透過環境變數 `AI_API_KEY` 傳入，不寫入 config

## 核心限制（修改前必讀）

1. **Adapter 不得 import `internal/engine`**：會造成循環依賴
2. **Engine 路由依訊息文字**：`"ok"`/`"no"` → HandleConfirm，其他 → HandleMessage；不得用「有無 pending」判斷
3. **Webhook handler 必須非阻塞**：channel send 一律用 `select/default`
4. **AIProvider 回傳純字串**：不得含 JSON 或結構體包裝
5. **空 API Key 必須立即報錯**：error 訊息需含 `"api key"`

> **Agent 提示：** 執行任何開發任務前，先讀 `docs/requirements/README.md` 確認領域邊界，再讀對應文件。修改介面前必讀核心限制。
