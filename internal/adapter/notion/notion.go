package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
)

const notionAPIBase = "https://api.notion.com/v1"
const notionVersion = "2022-06-28"

type notionAdapter struct {
	token        string
	databaseID   string
	pollInterval time.Duration
	seenIDs      map[string]bool
	mu           sync.Mutex
	out          chan<- adapter.Message
}

func New(token, databaseID string, pollInterval time.Duration) adapter.Adapter {
	return &notionAdapter{
		token:        token,
		databaseID:   databaseID,
		pollInterval: pollInterval,
		seenIDs:      make(map[string]bool),
	}
}

func (n *notionAdapter) Name() string { return "notion" }

func (n *notionAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	n.out = out
	go n.poll(ctx)
	return nil
}

type notionPage struct {
	ID         string `json:"id"`
	Properties map[string]struct {
		Title []struct {
			PlainText string `json:"plain_text"`
		} `json:"title"`
		RichText []struct {
			PlainText string `json:"plain_text"`
		} `json:"rich_text"`
	} `json:"properties"`
	CreatedBy struct {
		ID string `json:"id"`
	} `json:"created_by"`
}

func (n *notionAdapter) poll(ctx context.Context) {
	ticker := time.NewTicker(n.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.fetchNew(ctx)
		}
	}
}

func (n *notionAdapter) fetchNew(ctx context.Context) {
	url := fmt.Sprintf("%s/databases/%s/query", notionAPIBase, n.databaseID)
	payload := []byte(`{"page_size":10}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		log.Printf("notion: build request error: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+n.token)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("notion: query error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("notion: read response error: %v", err)
		return
	}
	var result struct {
		Results []notionPage `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("notion: parse error: %v", err)
		return
	}

	for _, page := range result.Results {
		n.mu.Lock()
		seen := n.seenIDs[page.ID]
		if !seen {
			n.seenIDs[page.ID] = true
		}
		n.mu.Unlock()
		if seen {
			continue
		}
		title := extractTitle(page)
		content := extractContent(page)
		text := BuildText(title, content)
		if n.out != nil {
			select {
			case n.out <- adapter.Message{
				ID:        page.ID,
				Source:    "notion",
				ChannelID: page.ID,
				UserID:    page.CreatedBy.ID,
				Text:      text,
				Timestamp: time.Now(),
			}:
			default:
				log.Printf("notion: out channel full, dropping page %s", page.ID)
			}
		}
	}
}

func extractTitle(page notionPage) string {
	for _, prop := range page.Properties {
		if len(prop.Title) > 0 {
			return prop.Title[0].PlainText
		}
	}
	return page.ID
}

func extractContent(page notionPage) string {
	for _, prop := range page.Properties {
		if len(prop.RichText) > 0 {
			return prop.RichText[0].PlainText
		}
	}
	return ""
}

// BuildText is exported for testing.
func BuildText(title, content string) string {
	if content == "" {
		return title
	}
	return fmt.Sprintf("%s\n\n%s", title, content)
}

func (n *notionAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
	url := fmt.Sprintf("%s/comments", notionAPIBase)
	payload, err := json.Marshal(map[string]interface{}{
		"parent": map[string]string{"page_id": msg.ChannelID},
		"rich_text": []map[string]interface{}{
			{"type": "text", "text": map[string]string{"content": text}},
		},
	})
	if err != nil {
		return fmt.Errorf("notion: marshal reply payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+n.token)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("notion reply error: %v", err)
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notion: reply failed with status %d", resp.StatusCode)
	}
	return nil
}
