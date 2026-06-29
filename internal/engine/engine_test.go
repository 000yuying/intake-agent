package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
)

// fakeAdapter is a test double for adapter.Adapter.
type fakeAdapter struct {
	name    string
	replies []string
	mu      sync.Mutex
}

func (f *fakeAdapter) Name() string { return f.name }

func (f *fakeAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	// Send one message then wait for context cancellation.
	out <- adapter.Message{
		ID:        "msg1",
		Source:    f.name,
		ChannelID: "chan1",
		UserID:    "user1",
		Text:      "new requirement",
		Timestamp: time.Now(),
	}
	<-ctx.Done()
	return nil
}

func (f *fakeAdapter) Reply(_ context.Context, _ adapter.Message, text string) error {
	f.mu.Lock()
	f.replies = append(f.replies, text)
	f.mu.Unlock()
	return nil
}

// fakeConfirmEngine is a minimal stand-in that records calls.
type fakeConfirmEngine struct {
	handleConfirmFn func(ctx context.Context, msg adapter.Message) (string, bool, error)
	handleMessageFn func(ctx context.Context, msg adapter.Message) (string, error)
}

func (f *fakeConfirmEngine) HandleConfirm(ctx context.Context, msg adapter.Message) (string, bool, error) {
	if f.handleConfirmFn != nil {
		return f.handleConfirmFn(ctx, msg)
	}
	return "", false, nil
}

func (f *fakeConfirmEngine) HandleMessage(ctx context.Context, msg adapter.Message) (string, error) {
	if f.handleMessageFn != nil {
		return f.handleMessageFn(ctx, msg)
	}
	return "spec draft", nil
}

// confirmEngineIface is the interface Engine depends on (defined in engine.go).
type confirmEngineIface interface {
	HandleConfirm(ctx context.Context, msg adapter.Message) (string, bool, error)
	HandleMessage(ctx context.Context, msg adapter.Message) (string, error)
}

// TestEngine_NewEngine verifies that NewEngine returns a non-nil Engine.
func TestEngine_NewEngine(t *testing.T) {
	fake := &fakeConfirmEngine{}
	eng := NewEngine(fake)
	if eng == nil {
		t.Fatal("expected non-nil Engine")
	}
}

// TestEngine_handleMsg_newRequirement verifies that an unmatched message is treated as a new requirement
// and the reply is sent back via the matching adapter.
func TestEngine_handleMsg_newRequirement(t *testing.T) {
	replied := make(chan string, 1)
	fa := &fakeAdapter{name: "test"}
	// Override Reply to capture the reply text
	fa2 := &replyCapturingAdapter{name: "test", replied: replied}

	fc := &fakeConfirmEngine{
		handleConfirmFn: func(ctx context.Context, msg adapter.Message) (string, bool, error) {
			// no pending item
			return "", false, nil
		},
		handleMessageFn: func(ctx context.Context, msg adapter.Message) (string, error) {
			return "spec: " + msg.Text, nil
		},
	}

	_ = fa // suppress unused
	eng := NewEngine(fc, fa2)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	msg := adapter.Message{
		ID:        "1",
		Source:    "test",
		ChannelID: "c",
		UserID:    "u",
		Text:      "build a login page",
		Timestamp: time.Now(),
	}
	go eng.handleMsg(ctx, msg)

	select {
	case got := <-replied:
		if got != "spec: build a login page" {
			t.Errorf("unexpected reply: %q", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for reply")
	}
}

// TestEngine_handleMsg_confirmation verifies that a message with a pending confirmation is handled
// by HandleConfirm and the reply is sent back.
func TestEngine_handleMsg_confirmation(t *testing.T) {
	replied := make(chan string, 1)
	fa := &replyCapturingAdapter{name: "telegram", replied: replied}

	fc := &fakeConfirmEngine{
		handleConfirmFn: func(ctx context.Context, msg adapter.Message) (string, bool, error) {
			return "spec 已建立：output/spec.md", true, nil
		},
	}

	eng := NewEngine(fc, fa)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	msg := adapter.Message{
		ID:     "2",
		Source: "telegram",
		Text:   "ok",
	}
	go eng.handleMsg(ctx, msg)

	select {
	case got := <-replied:
		if got != "spec 已建立：output/spec.md" {
			t.Errorf("unexpected reply: %q", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for reply")
	}
}

// TestEngine_handleMsg_confirmError verifies that errors from HandleConfirm are logged and an error reply is sent.
func TestEngine_handleMsg_confirmError(t *testing.T) {
	replied := make(chan string, 1)
	fa := &replyCapturingAdapter{name: "slack", replied: replied}

	fc := &fakeConfirmEngine{
		handleConfirmFn: func(ctx context.Context, msg adapter.Message) (string, bool, error) {
			return "", false, errors.New("storage unavailable")
		},
	}

	eng := NewEngine(fc, fa)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	msg := adapter.Message{Source: "slack", Text: "ok"}
	go eng.handleMsg(ctx, msg)

	select {
	case got := <-replied:
		if !strings.Contains(got, "無法") {
			t.Errorf("expected error reply containing '無法', got: %q", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("timeout waiting for error reply")
	}
}

// replyCapturingAdapter captures reply text.
type replyCapturingAdapter struct {
	name    string
	replied chan string
}

func (r *replyCapturingAdapter) Name() string { return r.name }
func (r *replyCapturingAdapter) Start(_ context.Context, _ chan<- adapter.Message) error {
	return nil
}
func (r *replyCapturingAdapter) Reply(_ context.Context, _ adapter.Message, text string) error {
	r.replied <- text
	return nil
}

// errorAI always returns an error
type errorAI struct{}

func (e *errorAI) HandleConfirm(ctx context.Context, msg adapter.Message) (string, bool, error) {
	return "", false, fmt.Errorf("AI service unavailable")
}

func (e *errorAI) HandleMessage(ctx context.Context, msg adapter.Message) (string, error) {
	return "", fmt.Errorf("AI service unavailable")
}

// TestEngineRepliesOnAIError verifies that the Engine replies to the user when AI processing fails
func TestEngineRepliesOnAIError(t *testing.T) {
	replied := make(chan string, 1)
	fa := &replyCapturingAdapter{name: "test", replied: replied}

	fc := &errorAI{}

	eng := NewEngine(fc, fa)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	msg := adapter.Message{
		ID:        "1",
		Source:    "test",
		ChannelID: "c1",
		UserID:    "u1",
		Text:      "新功能需求",
		Timestamp: time.Now(),
	}
	go eng.handleMsg(ctx, msg)

	select {
	case reply := <-replied:
		if !strings.Contains(reply, "無法") && !strings.Contains(reply, "錯誤") && !strings.Contains(reply, "error") {
			t.Errorf("expected error reply to user, got: %s", reply)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for error reply to user")
	}
}
