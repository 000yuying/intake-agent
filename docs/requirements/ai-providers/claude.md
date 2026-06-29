# Claude AI Provider

## 設定

```yaml
ai:
  provider: claude
  model: claude-sonnet-4-6   # 推薦；也可用 claude-haiku-4-5-20251001（快速省錢）
```

```bash
export AI_API_KEY=sk-ant-...   # Anthropic Console 取得
```

## SDK

`github.com/anthropics/anthropic-sdk-go`

## 可用 Model

| Model ID | 說明 |
| :--- | :--- |
| `claude-sonnet-4-6` | 預設，品質與速度平衡 |
| `claude-haiku-4-5-20251001` | 最快、最省，適合簡單需求 |
| `claude-opus-4-8` | 最強，適合複雜需求分析 |
