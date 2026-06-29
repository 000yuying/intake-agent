package output_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuying/intake-agent/internal/output"
)

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	w := output.NewWriter(dir, "specs/")
	path, err := w.Write("telegram", "## 需求概述\n測試需求")
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if !strings.HasPrefix(path, "specs/") {
		t.Errorf("expected path to start with specs/, got %s", path)
	}
	if !strings.HasSuffix(path, "-telegram.md") {
		t.Errorf("expected path to end with -telegram.md, got %s", path)
	}
	full := filepath.Join(dir, path)
	data, err := os.ReadFile(full)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if !strings.Contains(string(data), "測試需求") {
		t.Error("file content does not contain expected text")
	}
}
