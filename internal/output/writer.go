package output

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Writer struct {
	repoPath string
	dir      string
}

func NewWriter(repoPath, dir string) *Writer {
	return &Writer{repoPath: repoPath, dir: dir}
}

func (w *Writer) Write(source, content string) (string, error) {
	if err := os.MkdirAll(filepath.Join(w.repoPath, w.dir), 0755); err != nil {
		return "", err
	}
	ts := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("%s-%s.md", ts, source)
	relPath := filepath.Join(w.dir, filename)
	fullPath := filepath.Join(w.repoPath, relPath)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", err
	}
	return relPath, nil
}
