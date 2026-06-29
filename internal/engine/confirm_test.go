// internal/engine/confirm_test.go
package engine_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
	"github.com/yuying/intake-agent/internal/engine"
	"github.com/yuying/intake-agent/internal/output"
)

type fakeAI struct{}

func (f *fakeAI) Name() string { return "fake" }
func (f *fakeAI) GenerateSpec(_ context.Context, msg string) (string, error) {
	return "## 需求概述\n" + msg, nil
}

func TestHandleMessage(t *testing.T) {
	w := output.NewWriter(t.TempDir(), "specs/")
	e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
	msg := adapter.Message{
		ID: "1", Source: "telegram", ChannelID: "c1", UserID: "u1",
		Text: "新功能：登入頁改版",
	}
	reply, err := e.HandleMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("HandleMessage error: %v", err)
	}
	if !strings.Contains(reply, "ok") {
		t.Errorf("expected reply to mention 'ok', got: %s", reply)
	}
	if !strings.Contains(reply, "登入頁改版") {
		t.Errorf("expected reply to contain user message, got: %s", reply)
	}
}

func TestHandleConfirm_OK(t *testing.T) {
	w := output.NewWriter(t.TempDir(), "specs/")
	e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
	orig := adapter.Message{ID: "1", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "需求A"}
	e.HandleMessage(context.Background(), orig)

	confirm := adapter.Message{ID: "2", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "ok"}
	reply, wrote, err := e.HandleConfirm(context.Background(), confirm)
	if err != nil {
		t.Fatalf("HandleConfirm error: %v", err)
	}
	if !wrote {
		t.Error("expected wrote=true")
	}
	if !strings.Contains(reply, "specs/") {
		t.Errorf("expected reply to contain file path, got: %s", reply)
	}
}

func TestHandleConfirm_No(t *testing.T) {
	w := output.NewWriter(t.TempDir(), "specs/")
	e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
	orig := adapter.Message{ID: "1", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "需求B"}
	e.HandleMessage(context.Background(), orig)

	confirm := adapter.Message{ID: "2", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "no"}
	_, wrote, err := e.HandleConfirm(context.Background(), confirm)
	if err != nil {
		t.Fatalf("HandleConfirm error: %v", err)
	}
	if wrote {
		t.Error("expected wrote=false for 'no'")
	}
}

func TestHandleConfirm_NoPending(t *testing.T) {
	w := output.NewWriter(t.TempDir(), "specs/")
	e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)

	confirm := adapter.Message{ID: "1", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "ok"}
	reply, wrote, err := e.HandleConfirm(context.Background(), confirm)
	if err != nil {
		t.Fatalf("HandleConfirm error: %v", err)
	}
	if wrote {
		t.Error("expected wrote=false when no pending")
	}
	if !strings.Contains(reply, "找不到待確認") {
		t.Errorf("expected reply to mention 找不到待確認, got: %s", reply)
	}
}

func TestHandleConfirm_Timeout(t *testing.T) {
	w := output.NewWriter(t.TempDir(), "specs/")
	e := engine.NewConfirmEngine(&fakeAI{}, w, 1*time.Millisecond)
	orig := adapter.Message{ID: "1", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "需求C"}
	e.HandleMessage(context.Background(), orig)

	time.Sleep(5 * time.Millisecond)

	confirm := adapter.Message{ID: "2", Source: "telegram", ChannelID: "c1", UserID: "u1", Text: "ok"}
	_, wrote, err := e.HandleConfirm(context.Background(), confirm)
	if err != nil {
		t.Fatalf("HandleConfirm error: %v", err)
	}
	if wrote {
		t.Error("expected wrote=false after timeout")
	}
}

func TestStartCleanup_RemovesExpired(t *testing.T) {
	w := output.NewWriter(t.TempDir(), "specs/")
	e := engine.NewConfirmEngine(&fakeAI{}, w, 50*time.Millisecond)
	orig := adapter.Message{ID: "1", Source: "telegram", ChannelID: "c99", UserID: "u99", Text: "需求Z"}
	e.HandleMessage(context.Background(), orig)

	ctx, cancel := context.WithCancel(context.Background())
	e.StartCleanup(ctx)

	// 等超過 timeout，讓 cleanup goroutine 有機會執行
	time.Sleep(200 * time.Millisecond)
	cancel()

	// 過期後 HandleConfirm 應回傳找不到
	confirm := adapter.Message{ID: "2", Source: "telegram", ChannelID: "c99", UserID: "u99", Text: "ok"}
	_, wrote, err := e.HandleConfirm(context.Background(), confirm)
	if err != nil {
		t.Fatalf("HandleConfirm error: %v", err)
	}
	if wrote {
		t.Error("expected wrote=false — item should have been cleaned up")
	}
}
