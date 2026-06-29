// internal/adapter/slack/slack_test.go
package slack_test

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
	"strings"
	"testing"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
	slackadapter "github.com/yuying/intake-agent/internal/adapter/slack"
)

const testSecret = "test-signing-secret"

// signRequest adds Slack-compatible signing headers to the request.
func signRequest(t *testing.T, body []byte, r *http.Request) {
	t.Helper()
	ts := fmt.Sprintf("%d", time.Now().Unix())
	sigBase := "v0:" + ts + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(sigBase))
	sig := "v0=" + hex.EncodeToString(mac.Sum(nil))
	r.Header.Set("X-Slack-Request-Timestamp", ts)
	r.Header.Set("X-Slack-Signature", sig)
}

// TestSlackName verifies Name() returns "slack".
func TestSlackName(t *testing.T) {
	a := slackadapter.New(testSecret, "bot-token")
	if a.Name() != "slack" {
		t.Errorf("expected slack, got %s", a.Name())
	}
}

// TestSlackStart_URLVerification verifies the challenge is echoed back.
func TestSlackStart_URLVerification(t *testing.T) {
	mux := http.NewServeMux()
	a := slackadapter.NewWithMux(testSecret, "bot-token", mux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan adapter.Message, 5)
	if err := a.Start(ctx, out); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	body := `{"token":"x","challenge":"my-challenge-value","type":"url_verification"}`
	bodyBytes := []byte(body)
	req := httptest.NewRequest(http.MethodPost, "/slack/events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	signRequest(t, bodyBytes, req)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "my-challenge-value") {
		t.Errorf("expected challenge value in body, got: %s", rr.Body.String())
	}
}

// TestSlackStart_MessageEvent verifies a message event is pushed to the out channel.
func TestSlackStart_MessageEvent(t *testing.T) {
	mux := http.NewServeMux()
	a := slackadapter.NewWithMux(testSecret, "bot-token", mux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan adapter.Message, 5)
	if err := a.Start(ctx, out); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	event := map[string]interface{}{
		"token":      "x",
		"team_id":    "T1",
		"api_app_id": "A1",
		"event": map[string]interface{}{
			"type":    "message",
			"user":    "U123",
			"text":    "hello from slack",
			"ts":      "1234567890.123456",
			"channel": "C456",
		},
		"type":       "event_callback",
		"event_id":   "Ev1",
		"event_time": 1234567890,
	}
	bodyBytes, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/slack/events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	signRequest(t, bodyBytes, req)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	select {
	case msg := <-out:
		if msg.Source != "slack" {
			t.Errorf("expected source=slack, got %s", msg.Source)
		}
		if msg.Text != "hello from slack" {
			t.Errorf("expected text='hello from slack', got %s", msg.Text)
		}
		if msg.UserID != "U123" {
			t.Errorf("expected userID=U123, got %s", msg.UserID)
		}
		if msg.ChannelID != "C456" {
			t.Errorf("expected channelID=C456, got %s", msg.ChannelID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for message on channel")
	}
}

// TestSlackStart_InvalidSignature verifies requests with bad signatures return 401.
func TestSlackStart_InvalidSignature(t *testing.T) {
	mux := http.NewServeMux()
	a := slackadapter.NewWithMux(testSecret, "bot-token", mux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan adapter.Message, 5)
	if err := a.Start(ctx, out); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	body := `{"token":"x","challenge":"c","type":"url_verification"}`
	req := httptest.NewRequest(http.MethodPost, "/slack/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Slack-Request-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	req.Header.Set("X-Slack-Signature", "v0=badsignature")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad signature, got %d", rr.Code)
	}
}
