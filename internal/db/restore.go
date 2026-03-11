package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

var panelRestoreTables = []string{
	"panel_users",
	"panel_sessions",
	"panel_rooms",
	"panel_messages",
	"panel_room_members",
	"panel_message_reactions",
	"panel_message_pins",
	"panel_message_favorites",
	"panel_user_blocks",
	"panel_user_mutes",
	"panel_logs",
	"panel_presence",
	"panel_events",
	"panel_event_rsvps",
	"panel_polls",
	"panel_poll_options",
	"panel_poll_votes",
	"panel_join_requests",
}

func (s *Store) RestoreFromSnapshot(snapshotPath string) error {
	snapshotPath = strings.TrimSpace(snapshotPath)
	if snapshotPath == "" {
		return fmt.Errorf("snapshot path vazio")
	}
	escaped := strings.ReplaceAll(filepath.ToSlash(snapshotPath), "'", "''")
	ctx := context.Background()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open restore connection: %w", err)
	}
	defer conn.Close()

	inTx := false
	attached := false
	defer func() {
		if inTx {
			_, _ = conn.ExecContext(ctx, `ROLLBACK`)
		}
		if attached {
			_, _ = conn.ExecContext(ctx, `DETACH DATABASE imported`)
		}
		_, _ = conn.ExecContext(ctx, `PRAGMA foreign_keys = ON`)
	}()

	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}
	if _, err := conn.ExecContext(ctx, "ATTACH DATABASE '"+escaped+"' AS imported"); err != nil {
		return fmt.Errorf("attach imported snapshot: %w", err)
	}
	attached = true
	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return fmt.Errorf("begin restore snapshot: %w", err)
	}
	inTx = true

	for i := len(panelRestoreTables) - 1; i >= 0; i-- {
		if _, err := conn.ExecContext(ctx, `DELETE FROM `+panelRestoreTables[i]); err != nil {
			return fmt.Errorf("clear table %s: %w", panelRestoreTables[i], err)
		}
	}

	for _, table := range panelRestoreTables {
		exists, err := databaseTableExistsConn(ctx, conn, "imported", table)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}
		columns, err := sharedTableColumns(ctx, conn, table)
		if err != nil {
			return err
		}
		if len(columns) == 0 {
			continue
		}
		columnList := strings.Join(columns, ", ")
		if _, err := conn.ExecContext(ctx, `INSERT INTO `+table+` (`+columnList+`) SELECT `+columnList+` FROM imported.`+table); err != nil {
			return fmt.Errorf("restore table %s: %w", table, err)
		}
	}

	if exists, err := databaseTableExistsConn(ctx, conn, "imported", "sqlite_sequence"); err == nil && exists {
		if _, err := conn.ExecContext(ctx, `DELETE FROM sqlite_sequence WHERE name LIKE 'panel_%'`); err != nil {
			return fmt.Errorf("clear sqlite_sequence: %w", err)
		}
		if _, err := conn.ExecContext(ctx, `INSERT INTO sqlite_sequence(name, seq) SELECT name, seq FROM imported.sqlite_sequence WHERE name LIKE 'panel_%'`); err != nil {
			return fmt.Errorf("restore sqlite_sequence: %w", err)
		}
	}

	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return fmt.Errorf("commit restore snapshot: %w", err)
	}
	inTx = false
	if _, err := conn.ExecContext(ctx, `DETACH DATABASE imported`); err != nil {
		return fmt.Errorf("detach imported snapshot: %w", err)
	}
	attached = false
	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	return nil
}

func databaseTableExistsConn(ctx context.Context, conn *sql.Conn, schemaName, table string) (bool, error) {
	row := conn.QueryRowContext(ctx, `SELECT COUNT(1) FROM `+schemaName+`.sqlite_master WHERE type = 'table' AND name = ?`, table)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("inspect table %s.%s: %w", schemaName, table, err)
	}
	return count > 0, nil
}

func sharedTableColumns(ctx context.Context, conn *sql.Conn, table string) ([]string, error) {
	current, err := tableColumns(ctx, conn, "main", table)
	if err != nil {
		return nil, err
	}
	imported, err := tableColumns(ctx, conn, "imported", table)
	if err != nil {
		return nil, err
	}
	if len(current) == 0 || len(imported) == 0 {
		return nil, nil
	}
	shared := make([]string, 0, len(current))
	for _, name := range current {
		if slices.Contains(imported, name) {
			shared = append(shared, name)
		}
	}
	return shared, nil
}

func tableColumns(ctx context.Context, conn *sql.Conn, schemaName, table string) ([]string, error) {
	rows, err := conn.QueryContext(ctx, `PRAGMA `+schemaName+`.table_info(`+table+`)`)
	if err != nil {
		return nil, fmt.Errorf("inspect columns %s.%s: %w", schemaName, table, err)
	}
	defer rows.Close()

	columns := make([]string, 0, 16)
	for rows.Next() {
		var (
			cid        int
			name       string
			typeName   string
			notNull    int
			defaultV   sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &typeName, &notNull, &defaultV, &primaryKey); err != nil {
			return nil, fmt.Errorf("scan columns %s.%s: %w", schemaName, table, err)
		}
		if strings.TrimSpace(name) != "" {
			columns = append(columns, name)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate columns %s.%s: %w", schemaName, table, err)
	}
	return columns, nil
}
