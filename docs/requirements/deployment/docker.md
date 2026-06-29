# Docker 部署

## 快速啟動

```bash
# 1. 準備 config
cp configs/config.yaml.example configs/config.yaml
# 編輯 config.yaml，填入 token
# 注意：output.repo_path 改為 /specs（container 內路徑）

# 2. 啟動
AI_API_KEY=your_api_key docker-compose up -d
```

## docker-compose.yml 說明

```yaml
services:
  intake-agent:
    build: .
    ports:
      - "8080:8080"
    environment:
      - AI_API_KEY=${AI_API_KEY}
    volumes:
      - ./configs/config.yaml:/app/configs/config.yaml:ro   # config 唯讀掛入
      - ${SPECS_REPO_PATH:-./specs}:/specs                   # spec 輸出目錄
    restart: unless-stopped
```

## Volume 設定

| Volume | 說明 |
| :--- | :--- |
| `./configs/config.yaml` | config 唯讀掛入 container |
| `${SPECS_REPO_PATH:-./specs}` | spec 輸出目錄，預設為 `./specs` |

**重要**：`configs/config.yaml` 的 `output.repo_path` 必須設為 `/specs`（container 內路徑），否則寫檔會失敗：

```yaml
output:
  repo_path: /specs    # Docker 環境使用此路徑
  dir: specs/
```

## 指定自訂 specs 路徑

```bash
SPECS_REPO_PATH=/home/yuying/my-specs AI_API_KEY=xxx docker-compose up -d
```

## 重新 build

```bash
docker-compose build --no-cache
docker-compose up -d
```

## 查看 log

```bash
docker-compose logs -f intake-agent
```

## 安全注意事項

- `configs/config.yaml` 包含 token，不得 commit 到 git（`.gitignore` 已排除）
- image build 時不會複製 `config.yaml`（`.dockerignore` 排除）
- `AI_API_KEY` 透過環境變數傳入，不寫入 config 或 image
