// internal/adapter/github/github_test.go
package github_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yuying/intake-agent/internal/adapter"
	githubadapter "github.com/yuying/intake-agent/internal/adapter/github"
)

const testSecret = "test-webhook-secret"

func signGitHub(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestGitHubName(t *testing.T) {
	a := githubadapter.New(testSecret, "token", "owner/repo")
	if a.Name() != "github" {
		t.Errorf("expected github, got %s", a.Name())
	}
}

func TestGitHubHandleIssueOpened(t *testing.T) {
	mux := http.NewServeMux()
	a := githubadapter.NewWithMux(testSecret, "token", "owner/repo", mux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan adapter.Message, 5)
	if err := a.Start(ctx, out); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	payload := map[string]interface{}{
		"action": "opened",
		"issue": map[string]interface{}{
			"number": 42,
			"title":  "新功能需求",
			"body":   "希望增加登入功能",
			"user":   map[string]interface{}{"login": "user1"},
		},
		"repository": map[string]interface{}{
			"full_name": "owner/repo",
		},
	}
	body, _ := json.Marshal(payload)
	sig := signGitHub(body, testSecret)

	req := httptest.NewRequest(http.MethodPost, "/github/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "issues")
	req.Header.Set("X-Hub-Signature-256", sig)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	select {
	case msg := <-out:
		if msg.Source != "github" {
			t.Errorf("expected source github, got %s", msg.Source)
		}
		expected := "owner/repo/issues/42"
		if msg.ChannelID != expected {
			t.Errorf("expected channelID %s, got %s", expected, msg.ChannelID)
		}
		if msg.UserID != "user1" {
			t.Errorf("expected userID user1, got %s", msg.UserID)
		}
	default:
		t.Error("expected message in out channel")
	}
}

func TestGitHubInvalidSignature(t *testing.T) {
	mux := http.NewServeMux()
	a := githubadapter.NewWithMux(testSecret, "token", "owner/repo", mux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan adapter.Message, 5)
	a.Start(ctx, out)

	body := []byte(`{"action":"opened"}`)
	req := httptest.NewRequest(http.MethodPost, "/github/events", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "issues")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalidsig")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	_ = fmt.Sprintf("test done")
}
