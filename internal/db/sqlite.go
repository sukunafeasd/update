package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	path string
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	normalizedPath := filepath.ToSlash(path)
	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=trusted_schema(OFF)&_pragma=wal_autocheckpoint(1000)&_pragma=journal_size_limit(67108864)",
		normalizedPath,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	store := &Store{db: db, path: path}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Path() string {
	return strings.TrimSpace(s.path)
}

func (s *Store) SnapshotTo(dest string) error {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return fmt.Errorf("snapshot path vazio")
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	escaped := strings.ReplaceAll(filepath.ToSlash(dest), "'", "''")
	if _, err := s.db.Exec("VACUUM INTO '" + escaped + "'"); err != nil {
		return fmt.Errorf("snapshot sqlite: %w", err)
	}
	return nil
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS panel_users (
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

CREATE TABLE IF NOT EXISTS panel_sessions (
	id TEXT PRIMARY KEY,
	user_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	expires_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS panel_rooms (
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

CREATE TABLE IF NOT EXISTS panel_messages (
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

CREATE INDEX IF NOT EXISTS idx_panel_messages_room_created
ON panel_messages (room_id, id DESC);

CREATE TABLE IF NOT EXISTS panel_room_members (
	room_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY (room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_panel_room_members_user
ON panel_room_members (user_id, room_id);

CREATE TABLE IF NOT EXISTS panel_message_reactions (
	message_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	emoji TEXT NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY (message_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_panel_message_reactions_message
ON panel_message_reactions (message_id, created_at DESC);

CREATE TABLE IF NOT EXISTS panel_message_pins (
	message_id INTEGER PRIMARY KEY,
	room_id INTEGER NOT NULL,
	pinned_by INTEGER NOT NULL,
	created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_panel_message_pins_room
ON panel_message_pins (room_id, created_at DESC);

CREATE TABLE IF NOT EXISTS panel_message_favorites (
	message_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY (message_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_panel_message_favorites_user
ON panel_message_favorites (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS panel_user_blocks (
	blocker_id INTEGER NOT NULL,
	blocked_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY (blocker_id, blocked_id)
);

CREATE INDEX IF NOT EXISTS idx_panel_user_blocks_blocked
ON panel_user_blocks (blocked_id, created_at DESC);

CREATE TABLE IF NOT EXISTS panel_user_mutes (
	muter_id INTEGER NOT NULL,
	muted_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY (muter_id, muted_id)
);

CREATE INDEX IF NOT EXISTS idx_panel_user_mutes_muted
ON panel_user_mutes (muted_id, created_at DESC);

CREATE TABLE IF NOT EXISTS panel_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	action TEXT NOT NULL,
	actor_id INTEGER NOT NULL DEFAULT 0,
	actor_name TEXT NOT NULL DEFAULT '',
	room_id INTEGER NOT NULL DEFAULT 0,
	room_slug TEXT NOT NULL DEFAULT '',
	detail TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_panel_logs_created
ON panel_logs (id DESC);

CREATE TABLE IF NOT EXISTS panel_presence (
	user_id INTEGER PRIMARY KEY,
	room_id INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'online',
	last_seen_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS panel_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	room_id INTEGER NOT NULL DEFAULT 0,
	created_by INTEGER NOT NULL,
	created_by_name TEXT NOT NULL DEFAULT '',
	starts_at TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_panel_events_starts
ON panel_events (starts_at ASC, id DESC);

CREATE TABLE IF NOT EXISTS panel_event_rsvps (
	event_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY (event_id, user_id)
);

CREATE TABLE IF NOT EXISTS panel_polls (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	room_id INTEGER NOT NULL,
	question TEXT NOT NULL,
	created_by INTEGER NOT NULL,
	created_by_name TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_panel_polls_room_created
ON panel_polls (room_id, id DESC);

CREATE TABLE IF NOT EXISTS panel_poll_options (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	poll_id INTEGER NOT NULL,
	label TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_panel_poll_options_poll
ON panel_poll_options (poll_id, sort_order ASC, id ASC);

CREATE TABLE IF NOT EXISTS panel_poll_votes (
	poll_id INTEGER NOT NULL,
	option_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY (poll_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_panel_poll_votes_poll
ON panel_poll_votes (poll_id, option_id, created_at DESC);

CREATE TABLE IF NOT EXISTS panel_join_requests (
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
);

CREATE INDEX IF NOT EXISTS idx_panel_join_requests_email
ON panel_join_requests (email, id DESC);

CREATE INDEX IF NOT EXISTS idx_panel_join_requests_status
ON panel_join_requests (status, id DESC);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	if _, err := s.db.Exec(`
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS optimization_runs;
`); err != nil {
		return fmt.Errorf("drop legacy runtime tables: %w", err)
	}
	if err := s.ensureColumnExists("panel_messages", "reply_to_id", "ALTER TABLE panel_messages ADD COLUMN reply_to_id INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := s.ensureColumnExists("panel_messages", "updated_at", "ALTER TABLE panel_messages ADD COLUMN updated_at TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := s.ensureColumnExists("panel_rooms", "scope", "ALTER TABLE panel_rooms ADD COLUMN scope TEXT NOT NULL DEFAULT 'public'"); err != nil {
		return err
	}
	if err := s.ensureColumnExists("panel_users", "status_text", "ALTER TABLE panel_users ADD COLUMN status_text TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensureColumnExists(table, column, ddl string) error {
	rows, err := s.db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return fmt.Errorf("inspect %s columns: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			dataType   string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultVal, &primaryKey); err != nil {
			return fmt.Errorf("scan %s columns: %w", table, err)
		}
		if strings.EqualFold(name, column) {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate %s columns: %w", table, err)
	}

	if _, err := s.db.Exec(ddl); err != nil {
		return fmt.Errorf("alter %s add %s: %w", table, column, err)
	}
	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func intToBool(value int) bool {
	return value != 0
}
