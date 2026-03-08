package panel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncUploadsDirRejectsEmptyTarget(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "sample.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	if err := syncUploadsDir(source, ""); err == nil {
		t.Fatalf("expected error for empty target")
	}
}

func TestSyncUploadsDirCopiesFiles(t *testing.T) {
	source := t.TempDir()
	target := filepath.Join(t.TempDir(), "uploads")

	nested := filepath.Join(source, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "sample.txt"), []byte("payload"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	if err := syncUploadsDir(source, target); err != nil {
		t.Fatalf("sync uploads: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(target, "nested", "sample.txt"))
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if string(raw) != "payload" {
		t.Fatalf("unexpected copied payload: %q", string(raw))
	}
}
