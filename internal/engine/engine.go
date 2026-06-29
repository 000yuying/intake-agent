// internal/engine/engine.go
package engine

import (
	"context"
	"log"
	"strings"

	"github.com/yuying/intake-agent/internal/adapter"
)

// confirmProvider is the interface Engine uses to avoid a circular dependency
// with *ConfirmEngine. Both HandleMessage and HandleConfirm are needed.
type confirmProvider interface {
	HandleConfirm(ctx context.Context, msg adapter.Message) (string, bool, error)
	HandleMessage(ctx context.Context, msg adapter.Message) (string, error)
}

// Engine fans messages from all registered adapters to the ConfirmEngine.
type Engine struct {
	adapters []adapter.Adapter
	confirm  confirmProvider
}

// NewEngine constructs an Engine with the given ConfirmEngine and zero or more adapters.
func NewEngine(confirm confirmProvider, adapters ...adapter.Adapter) *Engine {
	return &Engine{confirm: confirm, adapters: adapters}
}

// Run starts all adapters and dispatches incoming messages until ctx is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	out := make(chan adapter.Message, 100)
	for _, a := range e.adapters {
		go func(a adapter.Adapter) {
			if err := a.Start(ctx, out); err != nil {
				log.Printf("adapter %s error: %v", a.Name(), err)
			}
		}(a)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-out:
			go e.handleMsg(ctx, msg)
		}
	}
}

func (e *Engine) handleMsg(ctx context.Context, msg adapter.Message) {
	text := strings.TrimSpace(strings.ToLower(msg.Text))
	if text == "ok" || text == "no" {
		replyText, _, err := e.confirm.HandleConfirm(ctx, msg)
		if err != nil {
			log.Printf("confirm error: %v", err)
			e.replyTo(msg, "系統暫時無法處理您的請求，請稍後再試。", ctx)
			return
		}
		e.replyTo(msg, replyText, ctx)
		return
	}
	replyText, err := e.confirm.HandleMessage(ctx, msg)
	if err != nil {
		log.Printf("engine error: %v", err)
		e.replyTo(msg, "系統暫時無法處理您的請求，請稍後再試。", ctx)
		return
	}
	e.replyTo(msg, replyText, ctx)
}

func (e *Engine) replyTo(msg adapter.Message, text string, ctx context.Context) {
	for _, a := range e.adapters {
		if a.Name() == msg.Source {
			if err := a.Reply(ctx, msg, text); err != nil {
				log.Printf("reply error: %v", err)
			}
			return
		}
	}
}
