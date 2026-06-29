# Spec 輸出格式規範

## 檔名格式

```
YYYY-MM-DD-HH-MM-SS-<source>.md
```

範例：`2026-06-29-14-30-00-telegram.md`

- `YYYY-MM-DD-HH-MM-SS`：寫入時間（`time.Now().Format("2006-01-02-15-04-05")`）
- `<source>`：訊息來源管道名稱（`msg.Source`，如 `telegram`、`slack`）

## 儲存路徑

```
{output.repo_path}/{output.dir}/{filename}
```

config 範例：
```yaml
output:
  repo_path: /home/yuying/specs   # 絕對路徑
  dir: specs/                     # repo 內子目錄，結尾加 /
```

實際路徑：`/home/yuying/specs/specs/2026-06-29-14-30-00-telegram.md`

## Markdown 內容格式

AI 產出的 spec 遵循以下三段式結構（由 prompt 強制）：

```markdown
## 需求概述
（一段話）

## 驗收條件
- 條件一
- 條件二

## 範圍外
- 項目一
```

## Writer 行為

- 目錄不存在時自動建立（`os.MkdirAll`，permission 0755）
- 檔案 permission：0644
- 回傳相對路徑（從 `repo_path` 起算），用於回覆給使用者
