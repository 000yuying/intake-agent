package notion_test

import (
	"testing"
	"time"

	notionadapter "github.com/yuying/intake-agent/internal/adapter/notion"
)

func TestNotionName(t *testing.T) {
	a := notionadapter.New("fake-token", "fake-db-id", 60*time.Second)
	if a.Name() != "notion" {
		t.Errorf("expected notion, got %s", a.Name())
	}
}

func TestNotionBuildText(t *testing.T) {
	title := "需要新功能"
	content := "希望增加登入功能"
	got := notionadapter.BuildText(title, content)
	if got == "" {
		t.Error("expected non-empty text")
	}
}
