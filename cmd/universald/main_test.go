package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureRuntimePathsCreatesDBParentAndUploadsDir(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "data", "nested", "universald.db")
	uploadsPath := filepath.Join(root, "panel_uploads")

	if err := ensureRuntimePaths(dbPath, uploadsPath); err != nil {
		t.Fatalf("ensure runtime paths: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(dbPath)); err != nil {
		t.Fatalf("expected db parent dir to exist: %v", err)
	}
	if _, err := os.Stat(uploadsPath); err != nil {
		t.Fatalf("expected uploads dir to exist: %v", err)
	}
}
