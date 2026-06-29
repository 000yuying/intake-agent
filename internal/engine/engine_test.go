package engine

import (
	"context"
	"errors"
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

// TestEngine_handleMsg_confirmError verifies that errors from HandleConfirm are logged (no panic/reply).
func TestEngine_handleMsg_confirmError(t *testing.T) {
	replied := make(chan string, 1)
	fa := &replyCapturingAdapter{name: "slack", replied: replied}

	fc := &fakeConfirmEngine{
		handleConfirmFn: func(ctx context.Context, msg adapter.Message) (string, bool, error) {
			return "", false, errors.New("storage unavailable")
		},
	}

	eng := NewEngine(fc, fa)

	ctx := context.Background()
	msg := adapter.Message{Source: "slack", Text: "ok"}
	eng.handleMsg(ctx, msg) // should not panic, should not reply

	select {
	case got := <-replied:
		t.Errorf("should not have replied, got: %q", got)
	default:
		// good: no reply
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
