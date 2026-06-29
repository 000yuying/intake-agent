// internal/engine/confirm.go
package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
	"github.com/yuying/intake-agent/internal/ai"
	"github.com/yuying/intake-agent/internal/output"
)

type pendingItem struct {
	msg      adapter.Message
	draft    string
	expireAt time.Time
}

// ConfirmEngine manages a pending-confirmation state machine per channel+user.
type ConfirmEngine struct {
	ai      ai.AIProvider
	writer  *output.Writer
	timeout time.Duration
	mu      sync.Mutex
	pending map[string]*pendingItem // key: channelID+":"+userID
}

// NewConfirmEngine creates a new ConfirmEngine.
func NewConfirmEngine(aiProvider ai.AIProvider, writer *output.Writer, timeout time.Duration) *ConfirmEngine {
	return &ConfirmEngine{
		ai:      aiProvider,
		writer:  writer,
		timeout: timeout,
		pending: make(map[string]*pendingItem),
	}
}

func (e *ConfirmEngine) key(msg adapter.Message) string {
	return msg.ChannelID + ":" + msg.UserID
}

// HandleMessage generates a spec draft from the user's message and stores it pending confirmation.
func (e *ConfirmEngine) HandleMessage(ctx context.Context, msg adapter.Message) (string, error) {
	draft, err := e.ai.GenerateSpec(ctx, msg.Text)
	if err != nil {
		return "", err
	}
	e.mu.Lock()
	e.pending[e.key(msg)] = &pendingItem{
		msg:      msg,
		draft:    draft,
		expireAt: time.Now().Add(e.timeout),
	}
	e.mu.Unlock()
	return fmt.Sprintf("以下是我理解的 spec，回覆 ok 確認 / no 捨棄：\n\n%s", draft), nil
}

// HandleConfirm processes a confirmation reply. "ok" writes the file; "no" discards.
// Returns (replyText, wrote, err). If no pending item exists, returns 找不到待確認..., false, nil.
func (e *ConfirmEngine) HandleConfirm(ctx context.Context, msg adapter.Message) (string, bool, error) {
	k := e.key(msg)
	e.mu.Lock()
	item, ok := e.pending[k]
	if ok {
		delete(e.pending, k)
	}
	e.mu.Unlock()

	if !ok {
		return "找不到待確認的 spec，請重新發送需求。", false, nil
	}

	if time.Now().After(item.expireAt) {
		return "找不到待確認的 spec，請重新發送需求。", false, nil
	}

	text := strings.TrimSpace(strings.ToLower(msg.Text))
	if text != "ok" {
		return "已捨棄。請重新描述需求。", false, nil
	}

	path, err := e.writer.Write(item.msg.Source, item.draft)
	if err != nil {
		return "", false, err
	}
	return fmt.Sprintf("spec 已建立：%s", path), true, nil
}
