# GitHub Issues Adapter

## 實作方式

HTTP webhook 接收 issues/issue_comment 事件，透過 GitHub REST API 新增 issue comment 回覆。

## 設定

```yaml
adapters:
  github:
    enabled: true
    webhook_secret: "your-secret"
    token: "ghp_..."         # Personal Access Token (issues:write)
    repo: "owner/repo"
```

## 設定 GitHub Webhook

1. 前往 repo → Settings → Webhooks → Add webhook
2. Payload URL：`https://<your-domain>/github/events`
3. Content type：`application/json`
4. Secret：填入 `webhook_secret`
5. Events：選 `Issues` + `Issue comments`

## 處理的事件

| 事件 | Action | 說明 |
| :--- | :--- | :--- |
| `issues` | `opened` | 新 Issue 建立 |
| `issue_comment` | `created` | Issue 新增留言 |

## ChannelID 格式

`owner/repo/issues/123`（Reply 用此解析 API endpoint）

## 簽名驗證

HMAC-SHA256，header：`X-Hub-Signature-256: sha256=<hex>`
