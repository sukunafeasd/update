package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestRestoreFromOlderSnapshotHandlesMissingNewColumns(t *testing.T) {
	legacyPath := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite", "file:"+filepath.ToSlash(legacyPath))
	if err != nil {
		t.Fatalf("open legacy sqlite: %v", err)
	}
	legacySchema := `
CREATE TABLE panel_users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	display_name TEXT NOT NULL,
	role TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	theme TEXT NOT NULL DEFAULT 'matrix',
	accent_color TEXT NOT NULL DEFAULT '#7bff00',
	avatar_url TEXT NOT NULL DEFAULT '',
	bio TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'online',
	status_text TEXT NOT NULL DEFAULT '',
	created_by INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	last_login_at TEXT
);
CREATE TABLE panel_sessions (
	id TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	expires_at TEXT NOT NULL
);
CREATE TABLE panel_rooms (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	slug TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	icon TEXT NOT NULL,
	category TEXT NOT NULL DEFAULT 'community',
	scope TEXT NOT NULL DEFAULT 'public',
	sort_order INTEGER NOT NULL DEFAULT 0,
	admin_only INTEGER NOT NULL DEFAULT 0,
	vip_only INTEGER NOT NULL DEFAULT 0,
	password_hash TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE TABLE panel_messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	room_id INTEGER NOT NULL,
	author_id INTEGER NOT NULL,
	author_name TEXT NOT NULL,
	author_role TEXT NOT NULL,
	body TEXT NOT NULL,
	kind TEXT NOT NULL,
	is_ai INTEGER NOT NULL DEFAULT 0,
	reply_to_id INTEGER NOT NULL DEFAULT 0,
	attachment_json TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE panel_room_members (room_id INTEGER NOT NULL, user_id INTEGER NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY (room_id, user_id));
CREATE TABLE panel_message_reactions (message_id INTEGER NOT NULL, user_id INTEGER NOT NULL, emoji TEXT NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY (message_id, user_id, emoji));
CREATE TABLE panel_message_pins (message_id INTEGER PRIMARY KEY, room_id INTEGER NOT NULL, pinned_by INTEGER NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE panel_message_favorites (message_id INTEGER NOT NULL, user_id INTEGER NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY (message_id, user_id));
CREATE TABLE panel_user_blocks (blocker_id INTEGER NOT NULL, blocked_id INTEGER NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY (blocker_id, blocked_id));
CREATE TABLE panel_user_mutes (muter_id INTEGER NOT NULL, muted_id INTEGER NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY (muter_id, muted_id));
CREATE TABLE panel_logs (id INTEGER PRIMARY KEY AUTOINCREMENT, action TEXT NOT NULL, actor_id INTEGER NOT NULL DEFAULT 0, actor_name TEXT NOT NULL DEFAULT '', room_id INTEGER NOT NULL DEFAULT 0, room_slug TEXT NOT NULL DEFAULT '', detail TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE panel_presence (user_id INTEGER PRIMARY KEY, room_id INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'online', last_seen_at TEXT NOT NULL);
CREATE TABLE panel_events (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', room_id INTEGER NOT NULL DEFAULT 0, created_by INTEGER NOT NULL, created_by_name TEXT NOT NULL DEFAULT '', starts_at TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE panel_event_rsvps (event_id INTEGER NOT NULL, user_id INTEGER NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY (event_id, user_id));
CREATE TABLE panel_polls (id INTEGER PRIMARY KEY AUTOINCREMENT, room_id INTEGER NOT NULL, question TEXT NOT NULL, created_by INTEGER NOT NULL, created_by_name TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE panel_poll_options (id INTEGER PRIMARY KEY AUTOINCREMENT, poll_id INTEGER NOT NULL, label TEXT NOT NULL, sort_order INTEGER NOT NULL DEFAULT 0);
CREATE TABLE panel_poll_votes (poll_id INTEGER NOT NULL, option_id INTEGER NOT NULL, user_id INTEGER NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY (poll_id, user_id));
CREATE TABLE panel_join_requests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL,
	display_name TEXT NOT NULL DEFAULT '',
	note TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'pending',
	review_note TEXT NOT NULL DEFAULT '',
	requested_at TEXT NOT NULL,
	reviewed_at TEXT NOT NULL DEFAULT '',
	reviewed_by INTEGER NOT NULL DEFAULT 0,
	reviewer_name TEXT NOT NULL DEFAULT '',
	access_code TEXT NOT NULL DEFAULT '',
	access_code_expires TEXT NOT NULL DEFAULT '',
	approved_user_id INTEGER NOT NULL DEFAULT 0,
	email_sent INTEGER NOT NULL DEFAULT 0
);`
	if _, err := legacy.Exec(legacySchema); err != nil {
		_ = legacy.Close()
		t.Fatalf("seed legacy schema: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := legacy.Exec(`INSERT INTO panel_users (username,email,display_name,role,password_hash,theme,accent_color,avatar_url,bio,status,status_text,created_by,created_at,updated_at,last_login_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		"dief", "dief@test.local", "Dief Antigo", "owner", "hash", "matrix", "#7bff00", "/uploads/avatar.png", "bio antiga", "online", "ativo", 0, now, now, now,
	); err != nil {
		_ = legacy.Close()
		t.Fatalf("seed legacy user: %v", err)
	}
	if _, err := legacy.Exec(`INSERT INTO panel_rooms (slug,name,description,icon,category,scope,sort_order,admin_only,vip_only,password_hash,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		"chat-geral", "Chat Geral", "Sala", "#", "chat", "public", 10, 0, 0, "", now, now,
	); err != nil {
		_ = legacy.Close()
		t.Fatalf("seed legacy room: %v", err)
	}
	if _, err := legacy.Exec(`INSERT INTO panel_messages (room_id,author_id,author_name,author_role,body,kind,is_ai,reply_to_id,attachment_json,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		1, 1, "Dief Antigo", "owner", "mensagem antiga", "text", 0, 0, "", now, "",
	); err != nil {
		_ = legacy.Close()
		t.Fatalf("seed legacy message: %v", err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	current, err := Open(filepath.Join(t.TempDir(), "current.db"))
	if err != nil {
		t.Fatalf("open current store: %v", err)
	}
	defer current.Close()

	if err := current.RestoreFromSnapshot(legacyPath); err != nil {
		t.Fatalf("restore legacy snapshot: %v", err)
	}

	user, err := current.GetPanelUserByLogin("dief")
	if err != nil {
		t.Fatalf("load restored user: %v", err)
	}
	if user.DisplayName != "Dief Antigo" {
		t.Fatalf("expected restored display name, got %q", user.DisplayName)
	}
	if user.BannerPreset != "grid" {
		t.Fatalf("expected missing legacy banner preset to fall back to default grid, got %q", user.BannerPreset)
	}
	var msgCount int
	if err := current.db.QueryRow(`SELECT COUNT(1) FROM panel_messages`).Scan(&msgCount); err != nil {
		t.Fatalf("count restored messages: %v", err)
	}
	if msgCount != 1 {
		t.Fatalf("expected restored messages, got %d", msgCount)
	}
}
