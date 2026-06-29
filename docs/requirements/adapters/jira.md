# Jira Adapter

## 實作方式

HTTP webhook 接收 issue 事件，透過 Jira REST API v3 新增 comment 回覆。

## 設定

```yaml
adapters:
  jira:
    enabled: true
    host: "https://company.atlassian.net"
    email: "you@company.com"
    api_token: "your-api-token"
```

## 設定 Jira Webhook

1. Jira Settings → System → WebHooks → Create a WebHook
2. URL：`https://<your-domain>/jira/events`
3. Events：勾選 `Issue → created`、`Issue → updated`、`Comment → created`

## 取得 API Token

1. https://id.atlassian.com/manage-profile/security/api-tokens
2. Create API token，複製 token 值

## 處理的事件

| webhookEvent | 說明 |
| :--- | :--- |
| `jira:issue_created` | Issue 建立 |
| `jira:issue_updated` | Issue 更新 |
| `comment_created` | Issue 新增留言 |

## ChannelID 格式

Jira issue key，例如 `PROJ-42`

## Reply 格式

Jira REST API v3 要求 Atlassian Document Format (ADF) JSON，已在 Reply() 中處理。
