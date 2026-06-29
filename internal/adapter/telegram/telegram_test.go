// internal/adapter/telegram/telegram_test.go
package telegram_test

import (
	"context"
	"testing"
	"time"

	"github.com/yuying/intake-agent/internal/adapter/telegram"
	"github.com/yuying/intake-agent/internal/engine"
	"github.com/yuying/intake-agent/internal/output"
)

type fakeAI struct{}

func (f *fakeAI) Name() string { return "fake" }
func (f *fakeAI) GenerateSpec(_ context.Context, msg string) (string, error) {
	return "## spec\n" + msg, nil
}

func TestTelegramName(t *testing.T) {
	w := output.NewWriter(t.TempDir(), "specs/")
	e := engine.NewConfirmEngine(&fakeAI{}, w, 10*time.Minute)
	a := telegram.New("fake-token", e)
	if a.Name() != "telegram" {
		t.Errorf("expected telegram, got %s", a.Name())
	}
}
