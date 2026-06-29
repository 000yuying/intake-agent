# intake-agent

多管道訊息收集與 AI spec 自動產生工具。

## 快速開始

1. 複製設定檔
   ```bash
   cp configs/config.yaml.example configs/config.yaml
   ```

2. 填入 token（Telegram bot token、Slack signing secret + bot token）

3. 設定 AI API key
   ```bash
   export AI_API_KEY=your_api_key
   ```

4. 啟動
   ```bash
   make start
   ```

## 本機開發（需要 public URL）

```bash
ngrok http 8080
# 將 ngrok URL 設定為 Telegram webhook 或 Slack Events URL
```

## 雲端部署

```bash
docker-compose up -d
```

## AI Provider 切換

在 `configs/config.yaml` 修改：
```yaml
ai:
  provider: gemini   # claude | gemini | codex
  model: gemini-2.0-flash
```

並設定對應的 `AI_API_KEY`。
