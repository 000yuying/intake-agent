// internal/adapter/jira/jira.go
package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
)

type jiraAdapter struct {
	host     string
	email    string
	apiToken string
	mux      *http.ServeMux
	out      chan<- adapter.Message
}

func New(host, email, apiToken string) adapter.Adapter {
	return NewWithMux(host, email, apiToken, http.DefaultServeMux)
}

func NewWithMux(host, email, apiToken string, mux *http.ServeMux) adapter.Adapter {
	return &jiraAdapter{host: host, email: email, apiToken: apiToken, mux: mux}
}

func (j *jiraAdapter) Name() string { return "jira" }

func (j *jiraAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	j.out = out
	j.mux.HandleFunc("/jira/events", j.handleEvent)
	return nil
}

type jiraEvent struct {
	WebhookEvent string `json:"webhookEvent"`
	Issue        struct {
		Key    string `json:"key"`
		Fields struct {
			Summary     string `json:"summary"`
			Description string `json:"description"`
			Reporter    struct {
				AccountID string `json:"accountId"`
			} `json:"reporter"`
		} `json:"fields"`
	} `json:"issue"`
	Comment struct {
		Body   string `json:"body"`
		Author struct {
			AccountID string `json:"accountId"`
		} `json:"author"`
	} `json:"comment"`
}

func (j *jiraAdapter) handleEvent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var event jiraEvent
	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var text, userID string
	switch event.WebhookEvent {
	case "jira:issue_created", "jira:issue_updated":
		text = fmt.Sprintf("[%s] %s", event.Issue.Fields.Summary, event.Issue.Fields.Description)
		userID = event.Issue.Fields.Reporter.AccountID
	case "comment_created":
		text = event.Comment.Body
		userID = event.Comment.Author.AccountID
	default:
		w.WriteHeader(http.StatusOK)
		return
	}
	if j.out != nil {
		select {
		case j.out <- adapter.Message{
			ID:        event.Issue.Key,
			Source:    "jira",
			ChannelID: event.Issue.Key,
			UserID:    userID,
			Text:      text,
			Timestamp: time.Now(),
		}:
		default:
			log.Printf("jira: out channel full, dropping event for %s", event.Issue.Key)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (j *jiraAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", j.host, msg.ChannelID)
	payload, _ := json.Marshal(map[string]interface{}{
		"body": map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{"type": "text", "text": text},
					},
				},
			},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	creds := base64.StdEncoding.EncodeToString([]byte(j.email + ":" + j.apiToken))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("jira reply error: %v", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}
