// internal/adapter/telegram/telegram_test.go
package telegram_test

import (
	"testing"

	"github.com/yuying/intake-agent/internal/adapter/telegram"
)

func TestTelegramName(t *testing.T) {
	a := telegram.New("fake-token")
	if a.Name() != "telegram" {
		t.Errorf("expected telegram, got %s", a.Name())
	}
}
