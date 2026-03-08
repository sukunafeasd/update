package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestOpenConfiguresPanelSQLitePragmas(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	var (
		journalMode   string
		foreignKeys   int
		busyTimeout   int
		trustedSchema int
	)

	if err := store.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if err := store.db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if err := store.db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatalf("query busy_timeout: %v", err)
	}
	if err := store.db.QueryRow("PRAGMA trusted_schema").Scan(&trustedSchema); err != nil {
		t.Fatalf("query trusted_schema: %v", err)
	}

	if !strings.EqualFold(strings.TrimSpace(journalMode), "wal") {
		t.Fatalf("expected WAL journal mode, got %q", journalMode)
	}
	if foreignKeys != 1 {
		t.Fatalf("expected foreign_keys=1, got %d", foreignKeys)
	}
	if busyTimeout < 5000 {
		t.Fatalf("expected busy_timeout >= 5000, got %d", busyTimeout)
	}
	if trustedSchema != 0 {
		t.Fatalf("expected trusted_schema=0, got %d", trustedSchema)
	}
}

func TestOpenDoesNotCreateLegacyRuntimeTables(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	for _, table := range []string{"profiles", "optimization_runs"} {
		var count int
		if err := store.db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name = ?`, table).Scan(&count); err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected legacy table %s to be absent, got count=%d", table, count)
		}
	}
}

func TestOpenDropsLegacyRuntimeTablesOnExistingDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panel.db")

	seed, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("seed open sqlite: %v", err)
	}
	if _, err := seed.Exec(`
CREATE TABLE profiles (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL
);
CREATE TABLE optimization_runs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_at TEXT NOT NULL
);
`); err != nil {
		_ = seed.Close()
		t.Fatalf("seed legacy tables: %v", err)
	}
	if err := seed.Close(); err != nil {
		t.Fatalf("close seed db: %v", err)
	}

	store, err := Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	for _, table := range []string{"profiles", "optimization_runs"} {
		var count int
		if err := store.db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name = ?`, table).Scan(&count); err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("expected legacy table %s to be dropped, got count=%d", table, count)
		}
	}
}

func TestSnapshotToCreatesReusableCopy(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "panel.db")
	store, err := Open(sourcePath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if store.Path() != sourcePath {
		t.Fatalf("expected store path %q, got %q", sourcePath, store.Path())
	}

	snapshotPath := filepath.Join(t.TempDir(), "panel.snapshot.db")
	if err := store.SnapshotTo(snapshotPath); err != nil {
		t.Fatalf("snapshot sqlite: %v", err)
	}

	info, err := os.Stat(snapshotPath)
	if err != nil {
		t.Fatalf("stat snapshot: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected non-empty snapshot file")
	}

	snapshotStore, err := Open(snapshotPath)
	if err != nil {
		t.Fatalf("open snapshot store: %v", err)
	}
	defer snapshotStore.Close()

	var count int
	if err := snapshotStore.db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name='panel_users'`).Scan(&count); err != nil {
		t.Fatalf("query snapshot schema: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected panel_users table in snapshot, got count=%d", count)
	}
}
