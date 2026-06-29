// internal/adapter/github/github.go
package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
)

type githubAdapter struct {
	webhookSecret string
	token         string
	repo          string
	mux           *http.ServeMux
	out           chan<- adapter.Message
}

func New(webhookSecret, token, repo string) adapter.Adapter {
	return NewWithMux(webhookSecret, token, repo, http.DefaultServeMux)
}

func NewWithMux(webhookSecret, token, repo string, mux *http.ServeMux) adapter.Adapter {
	return &githubAdapter{
		webhookSecret: webhookSecret,
		token:         token,
		repo:          repo,
		mux:           mux,
	}
}

func (g *githubAdapter) Name() string { return "github" }

func (g *githubAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	g.out = out
	g.mux.HandleFunc("/github/events", g.handleEvent)
	return nil
}

type issueEvent struct {
	Action string `json:"action"`
	Issue  struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		User   struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"issue"`
	Comment struct {
		Body string `json:"body"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"comment"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func (g *githubAdapter) verifySignature(body []byte, sig string) bool {
	if !strings.HasPrefix(sig, "sha256=") {
		return false
	}
	mac := hmac.New(sha256.New, []byte(g.webhookSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

func (g *githubAdapter) handleEvent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sig := r.Header.Get("X-Hub-Signature-256")
	if !g.verifySignature(body, sig) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "issues" && eventType != "issue_comment" {
		w.WriteHeader(http.StatusOK)
		return
	}
	var event issueEvent
	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 只處理 opened/created action
	if event.Action != "opened" && event.Action != "created" {
		w.WriteHeader(http.StatusOK)
		return
	}
	var text, userID string
	if eventType == "issues" {
		text = fmt.Sprintf("[%s] %s\n\n%s", event.Issue.Title, "", event.Issue.Body)
		userID = event.Issue.User.Login
	} else {
		text = event.Comment.Body
		userID = event.Comment.User.Login
	}
	channelID := fmt.Sprintf("%s/issues/%d", event.Repository.FullName, event.Issue.Number)
	if g.out != nil {
		select {
		case g.out <- adapter.Message{
			ID:        fmt.Sprintf("%s/%d", event.Repository.FullName, event.Issue.Number),
			Source:    "github",
			ChannelID: channelID,
			UserID:    userID,
			Text:      text,
			Timestamp: time.Now(),
		}:
		default:
			log.Printf("github: out channel full, dropping event from %s", userID)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (g *githubAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
	// ChannelID format: "owner/repo/issues/123"
	parts := strings.Split(msg.ChannelID, "/issues/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid github channelID: %s", msg.ChannelID)
	}
	repoPath := parts[0]
	issueNum := parts[1]
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", repoPath, issueNum)
	payload, _ := json.Marshal(map[string]string{"body": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("github reply error: %v", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}
