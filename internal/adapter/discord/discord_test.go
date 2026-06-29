// internal/adapter/discord/discord_test.go
package discord_test

import (
	"testing"
	discordadapter "github.com/yuying/intake-agent/internal/adapter/discord"
)

func TestDiscordName(t *testing.T) {
	a := discordadapter.New("fake-token")
	if a.Name() != "discord" {
		t.Errorf("expected discord, got %s", a.Name())
	}
}
