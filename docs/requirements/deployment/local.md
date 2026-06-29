# 本機開發

## 快速啟動

```bash
# 1. 複製設定檔
cp configs/config.yaml.example configs/config.yaml

# 2. 填入 token
#    - Telegram: adapters.telegram.token
#    - Slack: adapters.slack.signing_secret + bot_token

# 3. 設定 AI API Key
export AI_API_KEY=your_anthropic_api_key   # Claude 預設

# 4. 啟動
make start
```

## 環境變數

| 變數 | 說明 | 必填 |
| :--- | :--- | :--- |
| `AI_API_KEY` | AI provider 的 API key | 是 |

不同 provider 對應的 key：
- `claude`：Anthropic API key（`sk-ant-...`）
- `gemini`：Google AI Studio API key
- `codex`：OpenAI API key（`sk-...`）

## Telegram 本機開發

Telegram 長輪詢不需 public URL，直接 `make start` 即可。

## Slack 本機開發（需要 ngrok）

```bash
# 安裝 ngrok：https://ngrok.com/download
ngrok http 8080
```

取得 ngrok URL（如 `https://xxxx.ngrok.io`），填入 Slack App Event Subscriptions：
```
https://xxxx.ngrok.io/slack/events
```

詳細 Slack App 設定見 [adapters/slack.md](../adapters/slack.md)。

## Makefile 指令

| 指令 | 說明 |
| :--- | :--- |
| `make start` | 啟動服務（預設 `configs/config.yaml`） |
| `make start CONFIG=configs/my.yaml` | 指定 config |
| `make build` | 編譯為 binary |
| `make test` | 執行所有測試 |
| `make stop` | 停止服務 |
