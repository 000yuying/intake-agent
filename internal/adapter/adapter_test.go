package adapter_test

import (
	"testing"
	"time"
	"github.com/yuying/intake-agent/internal/adapter"
)

func TestMessageFields(t *testing.T) {
	msg := adapter.Message{
		ID:        "123",
		Source:    "telegram",
		ChannelID: "chan-1",
		UserID:    "user-1",
		Text:      "需要新功能",
		Timestamp: time.Now(),
	}
	if msg.Source != "telegram" {
		t.Errorf("expected source telegram, got %s", msg.Source)
	}
}
