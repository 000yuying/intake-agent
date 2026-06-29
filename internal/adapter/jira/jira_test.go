// internal/adapter/jira/jira_test.go
package jira_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yuying/intake-agent/internal/adapter"
	jiraadapter "github.com/yuying/intake-agent/internal/adapter/jira"
)

func TestJiraName(t *testing.T) {
	a := jiraadapter.New("https://company.atlassian.net", "user@email.com", "api-token")
	if a.Name() != "jira" {
		t.Errorf("expected jira, got %s", a.Name())
	}
}

func TestJiraHandleIssueCreated(t *testing.T) {
	mux := http.NewServeMux()
	a := jiraadapter.NewWithMux("https://company.atlassian.net", "user@email.com", "api-token", mux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan adapter.Message, 5)
	if err := a.Start(ctx, out); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	payload := map[string]interface{}{
		"webhookEvent": "jira:issue_created",
		"issue": map[string]interface{}{
			"key": "PROJ-42",
			"fields": map[string]interface{}{
				"summary":     "新功能需求",
				"description": "希望增加登入功能",
				"reporter":    map[string]interface{}{"accountId": "user-acc-id"},
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/jira/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	select {
	case msg := <-out:
		if msg.Source != "jira" {
			t.Errorf("expected source jira, got %s", msg.Source)
		}
		if msg.ChannelID != "PROJ-42" {
			t.Errorf("expected channelID PROJ-42, got %s", msg.ChannelID)
		}
	default:
		t.Error("expected message in out channel")
	}
}
