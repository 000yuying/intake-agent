// internal/adapter/gchat/gchat_test.go
package gchat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yuying/intake-agent/internal/adapter"
	gchatadapter "github.com/yuying/intake-agent/internal/adapter/gchat"
)

func TestGChatName(t *testing.T) {
	a := gchatadapter.New("https://example.com/webhook")
	if a.Name() != "gchat" {
		t.Errorf("expected gchat, got %s", a.Name())
	}
}

func TestGChatHandleMessage(t *testing.T) {
	mux := http.NewServeMux()
	a := gchatadapter.NewWithMux("https://example.com/webhook", mux)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan adapter.Message, 5)
	if err := a.Start(ctx, out); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	payload := map[string]interface{}{
		"type": "MESSAGE",
		"message": map[string]interface{}{
			"name": "spaces/AAA/messages/111",
			"text": "需要新功能",
			"sender": map[string]interface{}{
				"name": "users/user1",
			},
		},
		"space": map[string]interface{}{
			"name": "spaces/AAA",
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/gchat/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	select {
	case msg := <-out:
		if msg.Source != "gchat" {
			t.Errorf("expected source gchat, got %s", msg.Source)
		}
		if msg.Text != "需要新功能" {
			t.Errorf("expected text, got %s", msg.Text)
		}
		if msg.ChannelID != "spaces/AAA" {
			t.Errorf("expected channelID spaces/AAA, got %s", msg.ChannelID)
		}
	default:
		t.Error("expected message in out channel")
	}
}
