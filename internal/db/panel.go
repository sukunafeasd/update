package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"universald/internal/model"
)

func (s *Store) CountPanelUsers() (int, error) {
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_users`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count panel users: %w", err)
	}
	return count, nil
}

func (s *Store) PanelOpsSummary(now time.Time, onlineWindow time.Duration) (model.PanelOpsSummary, error) {
	summary := model.PanelOpsSummary{
		GeneratedAt: now.UTC(),
	}
	var err error

	if summary.Users, err = s.countByQuery(`SELECT COUNT(1) FROM panel_users`); err != nil {
		return model.PanelOpsSummary{}, fmt.Errorf("count ops users: %w", err)
	}
	if summary.Rooms, err = s.countByQuery(`SELECT COUNT(1) FROM panel_rooms`); err != nil {
		return model.PanelOpsSummary{}, fmt.Errorf("count ops rooms: %w", err)
	}
	if summary.Messages, err = s.countByQuery(`SELECT COUNT(1) FROM panel_messages`); err != nil {
		return model.PanelOpsSummary{}, fmt.Errorf("count ops messages: %w", err)
	}
	if summary.Events, err = s.countByQuery(`SELECT COUNT(1) FROM panel_events`); err != nil {
		return model.PanelOpsSummary{}, fmt.Errorf("count ops events: %w", err)
	}
	if summary.Polls, err = s.countByQuery(`SELECT COUNT(1) FROM panel_polls`); err != nil {
		return model.PanelOpsSummary{}, fmt.Errorf("count ops polls: %w", err)
	}
	if summary.ActiveSessions, err = s.countByQuery(`SELECT COUNT(1) FROM panel_sessions WHERE expires_at > ?`, now.UTC().Format(time.RFC3339Nano)); err != nil {
		return model.PanelOpsSummary{}, fmt.Errorf("count ops sessions: %w", err)
	}
	if summary.OnlineUsers, err = s.countByQuery(`SELECT COUNT(1) FROM panel_presence WHERE last_seen_at >= ?`, now.UTC().Add(-onlineWindow).Format(time.RFC3339Nano)); err != nil {
		return model.PanelOpsSummary{}, fmt.Errorf("count ops online users: %w", err)
	}

	return summary, nil
}

func (s *Store) countByQuery(query string, args ...any) (int, error) {
	row := s.db.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) ListPanelUsers() ([]model.PanelUser, error) {
	rows, err := s.db.Query(`
SELECT id, username, email, display_name, role, theme, accent_color, avatar_url, bio, status, status_text,
       created_by, created_at, updated_at, last_login_at, password_hash
FROM panel_users
ORDER BY CASE role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 WHEN 'vip' THEN 2 ELSE 3 END, display_name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list panel users: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelUser, 0)
	for rows.Next() {
		user, err := scanPanelUser(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel users: %w", err)
	}
	return items, nil
}

func (s *Store) GetPanelUserByID(id int64) (model.PanelUser, error) {
	row := s.db.QueryRow(`
SELECT id, username, email, display_name, role, theme, accent_color, avatar_url, bio, status, status_text,
       created_by, created_at, updated_at, last_login_at, password_hash
FROM panel_users
WHERE id = ?
LIMIT 1`, id)
	user, err := scanPanelUser(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelUser{}, fmt.Errorf("panel user %d not found", id)
		}
		return model.PanelUser{}, err
	}
	return user, nil
}

func (s *Store) GetPanelUserByLogin(login string) (model.PanelUser, error) {
	login = strings.TrimSpace(strings.ToLower(login))
	row := s.db.QueryRow(`
SELECT id, username, email, display_name, role, theme, accent_color, avatar_url, bio, status, status_text,
       created_by, created_at, updated_at, last_login_at, password_hash
FROM panel_users
WHERE LOWER(username) = ? OR LOWER(email) = ?
LIMIT 1`, login, login)
	user, err := scanPanelUser(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelUser{}, fmt.Errorf("panel user %q not found", login)
		}
		return model.PanelUser{}, err
	}
	return user, nil
}

func (s *Store) CreatePanelUser(input model.PanelUser) (model.PanelUser, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(input.Username) == "" {
		return model.PanelUser{}, errors.New("username obrigatorio")
	}
	if strings.TrimSpace(input.Email) == "" {
		return model.PanelUser{}, errors.New("email obrigatorio")
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		input.DisplayName = input.Username
	}
	if strings.TrimSpace(input.Role) == "" {
		input.Role = "member"
	}
	if strings.TrimSpace(input.Theme) == "" {
		input.Theme = "matrix"
	}
	if strings.TrimSpace(input.AccentColor) == "" {
		input.AccentColor = "#7bff00"
	}
	if strings.TrimSpace(input.Status) == "" {
		input.Status = "online"
	}
	if strings.TrimSpace(input.PasswordHash) == "" {
		return model.PanelUser{}, errors.New("password hash obrigatorio")
	}

	result, err := s.db.Exec(`
INSERT INTO panel_users (
	username, email, display_name, role, password_hash, theme, accent_color,
	avatar_url, bio, status, status_text, created_by, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		strings.ToLower(strings.TrimSpace(input.Username)),
		strings.ToLower(strings.TrimSpace(input.Email)),
		strings.TrimSpace(input.DisplayName),
		strings.TrimSpace(strings.ToLower(input.Role)),
		input.PasswordHash,
		strings.TrimSpace(strings.ToLower(input.Theme)),
		strings.TrimSpace(input.AccentColor),
		strings.TrimSpace(input.AvatarURL),
		strings.TrimSpace(input.Bio),
		strings.TrimSpace(strings.ToLower(input.Status)),
		strings.TrimSpace(input.StatusText),
		input.CreatedBy,
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return model.PanelUser{}, fmt.Errorf("create panel user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.PanelUser{}, fmt.Errorf("fetch panel user id: %w", err)
	}
	return s.GetPanelUserByID(id)
}

func (s *Store) UpdatePanelUserProfile(input model.PanelUser) (model.PanelUser, error) {
	if input.ID <= 0 {
		return model.PanelUser{}, errors.New("user id invalido")
	}
	_, err := s.db.Exec(`
UPDATE panel_users
SET display_name = ?, theme = ?, accent_color = ?, avatar_url = ?, bio = ?, status = ?, status_text = ?, updated_at = ?
WHERE id = ?`,
		strings.TrimSpace(input.DisplayName),
		strings.TrimSpace(strings.ToLower(input.Theme)),
		strings.TrimSpace(input.AccentColor),
		strings.TrimSpace(input.AvatarURL),
		strings.TrimSpace(input.Bio),
		strings.TrimSpace(strings.ToLower(input.Status)),
		strings.TrimSpace(input.StatusText),
		time.Now().UTC().Format(time.RFC3339Nano),
		input.ID,
	)
	if err != nil {
		return model.PanelUser{}, fmt.Errorf("update panel user profile: %w", err)
	}
	return s.GetPanelUserByID(input.ID)
}

func (s *Store) SetPanelUserLastLogin(userID int64, at time.Time) error {
	_, err := s.db.Exec(`UPDATE panel_users SET last_login_at = ?, updated_at = ? WHERE id = ?`,
		at.UTC().Format(time.RFC3339Nano),
		at.UTC().Format(time.RFC3339Nano),
		userID,
	)
	if err != nil {
		return fmt.Errorf("set panel user last login: %w", err)
	}
	return nil
}

func (s *Store) SavePanelSession(session model.PanelSession) error {
	_, err := s.db.Exec(`
INSERT INTO panel_sessions (id, user_id, created_at, expires_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	user_id = excluded.user_id,
	created_at = excluded.created_at,
	expires_at = excluded.expires_at`,
		session.ID,
		session.UserID,
		session.CreatedAt.UTC().Format(time.RFC3339Nano),
		session.ExpiresAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("save panel session: %w", err)
	}
	return nil
}

func (s *Store) GetPanelSession(id string) (model.PanelSession, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, created_at, expires_at
FROM panel_sessions
WHERE id = ?
LIMIT 1`, strings.TrimSpace(id))

	var session model.PanelSession
	var createdAt, expiresAt string
	if err := row.Scan(&session.ID, &session.UserID, &createdAt, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelSession{}, fmt.Errorf("panel session not found")
		}
		return model.PanelSession{}, fmt.Errorf("get panel session: %w", err)
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return model.PanelSession{}, fmt.Errorf("parse panel session created_at: %w", err)
	}
	expires, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return model.PanelSession{}, fmt.Errorf("parse panel session expires_at: %w", err)
	}
	session.CreatedAt = created
	session.ExpiresAt = expires
	return session, nil
}

func (s *Store) DeletePanelSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM panel_sessions WHERE id = ?`, strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("delete panel session: %w", err)
	}
	return nil
}

func (s *Store) DeleteExpiredPanelSessions(now time.Time) error {
	_, err := s.db.Exec(`DELETE FROM panel_sessions WHERE expires_at <= ?`, now.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("delete expired panel sessions: %w", err)
	}
	return nil
}

func (s *Store) DeleteStalePanelPresence(before time.Time) error {
	_, err := s.db.Exec(`DELETE FROM panel_presence WHERE last_seen_at <= ?`, before.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("delete stale panel presence: %w", err)
	}
	return nil
}

func (s *Store) UpsertPanelPresence(userID, roomID int64, status string, seenAt time.Time) error {
	_, err := s.db.Exec(`
INSERT INTO panel_presence (user_id, room_id, status, last_seen_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
	room_id = excluded.room_id,
	status = excluded.status,
	last_seen_at = excluded.last_seen_at`,
		userID,
		roomID,
		strings.TrimSpace(strings.ToLower(status)),
		seenAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("upsert panel presence: %w", err)
	}
	return nil
}

func (s *Store) ListPanelPresence(now time.Time, onlineWindow time.Duration) ([]model.PanelPresence, error) {
	rows, err := s.db.Query(`
SELECT u.id, u.username, u.display_name, u.role, u.theme, u.accent_color, u.avatar_url, u.bio, u.status_text,
       COALESCE(p.status, 'offline'), COALESCE(p.room_id, 0), COALESCE(p.last_seen_at, u.updated_at)
FROM panel_users u
LEFT JOIN panel_presence p ON p.user_id = u.id
ORDER BY CASE u.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 WHEN 'vip' THEN 2 ELSE 3 END, u.display_name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list panel presence: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelPresence, 0)
	for rows.Next() {
		var item model.PanelPresence
		var lastSeen string
		if err := rows.Scan(
			&item.UserID,
			&item.Username,
			&item.DisplayName,
			&item.Role,
			&item.Theme,
			&item.AccentColor,
			&item.AvatarURL,
			&item.Bio,
			&item.StatusText,
			&item.Status,
			&item.RoomID,
			&lastSeen,
		); err != nil {
			return nil, fmt.Errorf("scan panel presence: %w", err)
		}
		ts, err := time.Parse(time.RFC3339Nano, lastSeen)
		if err != nil {
			return nil, fmt.Errorf("parse panel presence last_seen_at: %w", err)
		}
		item.LastSeenAt = ts
		item.Online = now.UTC().Sub(ts) <= onlineWindow
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel presence: %w", err)
	}
	return items, nil
}

func (s *Store) GetPanelPresenceByUserID(userID int64, now time.Time, onlineWindow time.Duration) (model.PanelPresence, error) {
	row := s.db.QueryRow(`
SELECT u.id, u.username, u.display_name, u.role, u.theme, u.accent_color, u.avatar_url, u.bio,
       COALESCE(p.status, 'offline'), COALESCE(p.room_id, 0), COALESCE(p.last_seen_at, u.updated_at)
FROM panel_users u
LEFT JOIN panel_presence p ON p.user_id = u.id
WHERE u.id = ?
LIMIT 1`, userID)

	var item model.PanelPresence
	var lastSeen string
	if err := row.Scan(
		&item.UserID,
		&item.Username,
		&item.DisplayName,
		&item.Role,
		&item.Theme,
		&item.AccentColor,
		&item.AvatarURL,
		&item.Bio,
		&item.Status,
		&item.RoomID,
		&lastSeen,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelPresence{}, fmt.Errorf("panel presence %d not found", userID)
		}
		return model.PanelPresence{}, fmt.Errorf("get panel presence: %w", err)
	}
	ts, err := time.Parse(time.RFC3339Nano, lastSeen)
	if err != nil {
		return model.PanelPresence{}, fmt.Errorf("parse panel presence last_seen_at: %w", err)
	}
	item.LastSeenAt = ts
	item.Online = now.UTC().Sub(ts) <= onlineWindow
	return item, nil
}

func (s *Store) UpsertPanelRoom(input model.PanelRoom) (model.PanelRoom, error) {
	now := time.Now().UTC()
	if strings.TrimSpace(input.Slug) == "" {
		return model.PanelRoom{}, errors.New("slug obrigatorio")
	}
	if strings.TrimSpace(input.Name) == "" {
		return model.PanelRoom{}, errors.New("nome da sala obrigatorio")
	}
	if strings.TrimSpace(input.Icon) == "" {
		input.Icon = "#"
	}
	if strings.TrimSpace(input.Category) == "" {
		input.Category = "community"
	}
	if strings.TrimSpace(input.Scope) == "" {
		input.Scope = "public"
	}

	_, err := s.db.Exec(`
INSERT INTO panel_rooms (
	slug, name, description, icon, category, scope, sort_order, admin_only, vip_only, password_hash, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(slug) DO UPDATE SET
	name = excluded.name,
	description = excluded.description,
	icon = excluded.icon,
	category = excluded.category,
	scope = excluded.scope,
	sort_order = excluded.sort_order,
	admin_only = excluded.admin_only,
	vip_only = excluded.vip_only,
	password_hash = excluded.password_hash,
	updated_at = excluded.updated_at`,
		strings.TrimSpace(strings.ToLower(input.Slug)),
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Description),
		strings.TrimSpace(input.Icon),
		strings.TrimSpace(strings.ToLower(input.Category)),
		strings.TrimSpace(strings.ToLower(input.Scope)),
		input.SortOrder,
		boolToInt(input.AdminOnly),
		boolToInt(input.VIPOnly),
		strings.TrimSpace(input.PasswordHash),
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return model.PanelRoom{}, fmt.Errorf("upsert panel room: %w", err)
	}
	return s.GetPanelRoomBySlug(input.Slug)
}

func (s *Store) GetPanelRoomBySlug(slug string) (model.PanelRoom, error) {
	row := s.db.QueryRow(`
SELECT r.id, r.slug, r.name, r.description, r.icon, r.category, r.scope, r.sort_order, r.admin_only, r.vip_only,
       r.password_hash, r.created_at, r.updated_at,
       m.created_at, m.body
FROM panel_rooms r
LEFT JOIN panel_messages m ON m.id = (
	SELECT pm.id FROM panel_messages pm WHERE pm.room_id = r.id ORDER BY pm.id DESC LIMIT 1
)
WHERE r.slug = ?
LIMIT 1`, strings.TrimSpace(strings.ToLower(slug)))
	room, err := scanPanelRoom(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelRoom{}, fmt.Errorf("panel room %q not found", slug)
		}
		return model.PanelRoom{}, err
	}
	return room, nil
}

func (s *Store) GetPanelRoomByID(id int64) (model.PanelRoom, error) {
	row := s.db.QueryRow(`
SELECT r.id, r.slug, r.name, r.description, r.icon, r.category, r.scope, r.sort_order, r.admin_only, r.vip_only,
       r.password_hash, r.created_at, r.updated_at,
       m.created_at, m.body
FROM panel_rooms r
LEFT JOIN panel_messages m ON m.id = (
	SELECT pm.id FROM panel_messages pm WHERE pm.room_id = r.id ORDER BY pm.id DESC LIMIT 1
)
WHERE r.id = ?
LIMIT 1`, id)
	room, err := scanPanelRoom(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelRoom{}, fmt.Errorf("panel room %d not found", id)
		}
		return model.PanelRoom{}, err
	}
	return room, nil
}

func (s *Store) ListPanelRooms() ([]model.PanelRoom, error) {
	rows, err := s.db.Query(`
SELECT r.id, r.slug, r.name, r.description, r.icon, r.category, r.scope, r.sort_order, r.admin_only, r.vip_only,
       r.password_hash, r.created_at, r.updated_at,
       m.created_at, m.body
FROM panel_rooms r
LEFT JOIN panel_messages m ON m.id = (
	SELECT pm.id FROM panel_messages pm WHERE pm.room_id = r.id ORDER BY pm.id DESC LIMIT 1
)
ORDER BY r.sort_order ASC, r.name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list panel rooms: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelRoom, 0)
	for rows.Next() {
		room, err := scanPanelRoom(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel rooms: %w", err)
	}
	return items, nil
}

func (s *Store) AddPanelRoomMembers(roomID int64, userIDs ...int64) error {
	if roomID <= 0 {
		return errors.New("room id invalido")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, userID := range uniqueInt64s(userIDs) {
		if _, err := s.db.Exec(`
INSERT INTO panel_room_members (room_id, user_id, created_at)
VALUES (?, ?, ?)
ON CONFLICT(room_id, user_id) DO NOTHING`, roomID, userID, now); err != nil {
			return fmt.Errorf("add panel room member: %w", err)
		}
	}
	return nil
}

func (s *Store) IsPanelRoomMember(roomID, userID int64) (bool, error) {
	if roomID <= 0 || userID <= 0 {
		return false, errors.New("membro invalido")
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_room_members WHERE room_id = ? AND user_id = ?`, roomID, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check panel room member: %w", err)
	}
	return count > 0, nil
}

func (s *Store) ListPanelRoomMemberIDs(roomID int64) ([]int64, error) {
	if roomID <= 0 {
		return nil, errors.New("room id invalido")
	}
	rows, err := s.db.Query(`SELECT user_id FROM panel_room_members WHERE room_id = ? ORDER BY user_id ASC`, roomID)
	if err != nil {
		return nil, fmt.Errorf("list panel room members: %w", err)
	}
	defer rows.Close()

	items := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan panel room member: %w", err)
		}
		items = append(items, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel room members: %w", err)
	}
	return items, nil
}

func (s *Store) CreatePanelMessage(input model.PanelMessage) (model.PanelMessage, error) {
	if input.RoomID <= 0 {
		return model.PanelMessage{}, errors.New("room id invalido")
	}
	if input.AuthorID <= 0 {
		return model.PanelMessage{}, errors.New("author id invalido")
	}
	if strings.TrimSpace(input.Body) == "" && input.Attachment == nil {
		return model.PanelMessage{}, errors.New("mensagem vazia")
	}
	if strings.TrimSpace(input.Kind) == "" {
		input.Kind = "text"
	}

	attachmentJSON := ""
	if input.Attachment != nil {
		buf, err := json.Marshal(input.Attachment)
		if err != nil {
			return model.PanelMessage{}, fmt.Errorf("marshal panel attachment: %w", err)
		}
		attachmentJSON = string(buf)
	}

	now := time.Now().UTC()
	result, err := s.db.Exec(`
INSERT INTO panel_messages (
	room_id, author_id, author_name, author_role, body, kind, is_ai, reply_to_id, attachment_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.RoomID,
		input.AuthorID,
		strings.TrimSpace(input.AuthorName),
		strings.TrimSpace(strings.ToLower(input.AuthorRole)),
		strings.TrimSpace(input.Body),
		strings.TrimSpace(strings.ToLower(input.Kind)),
		boolToInt(input.IsAI),
		input.ReplyToID,
		attachmentJSON,
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return model.PanelMessage{}, fmt.Errorf("create panel message: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.PanelMessage{}, fmt.Errorf("fetch panel message id: %w", err)
	}
	return s.GetPanelMessageForViewer(id, input.AuthorID)
}

func (s *Store) GetPanelMessageByID(id int64) (model.PanelMessage, error) {
	return s.GetPanelMessageForViewer(id, 0)
}

func (s *Store) GetPanelMessageForViewer(id, viewerID int64) (model.PanelMessage, error) {
	row := s.db.QueryRow(`
SELECT id, room_id, author_id, author_name, author_role, body, kind, is_ai, reply_to_id, attachment_json, created_at
       , updated_at
FROM panel_messages
WHERE id = ?
LIMIT 1`, id)
	msg, err := scanPanelMessage(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelMessage{}, fmt.Errorf("panel message %d not found", id)
		}
		return model.PanelMessage{}, err
	}
	items, err := s.hydratePanelMessages([]model.PanelMessage{msg}, viewerID)
	if err != nil {
		return model.PanelMessage{}, err
	}
	return items[0], nil
}

func (s *Store) ListPanelMessages(roomID int64, limit int) ([]model.PanelMessage, error) {
	return s.ListPanelMessagesForViewer(roomID, limit, 0)
}

func (s *Store) ListPanelMessagesForViewer(roomID int64, limit int, viewerID int64) ([]model.PanelMessage, error) {
	if roomID <= 0 {
		return nil, errors.New("room id invalido")
	}
	if limit <= 0 {
		limit = 60
	}
	rows, err := s.db.Query(`
SELECT id, room_id, author_id, author_name, author_role, body, kind, is_ai, reply_to_id, attachment_json, created_at
       , updated_at
FROM (
	SELECT id, room_id, author_id, author_name, author_role, body, kind, is_ai, reply_to_id, attachment_json, created_at, updated_at
	FROM panel_messages
	WHERE room_id = ?
	ORDER BY id DESC
	LIMIT ?
)
ORDER BY id ASC`, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("list panel messages: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelMessage, 0, limit)
	for rows.Next() {
		msg, err := scanPanelMessage(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel messages: %w", err)
	}
	return s.hydratePanelMessages(items, viewerID)
}

func (s *Store) UpdatePanelMessageBody(messageID int64, body string) error {
	if messageID <= 0 {
		return errors.New("message id invalido")
	}
	_, err := s.db.Exec(`
UPDATE panel_messages
SET body = ?, updated_at = ?
WHERE id = ?`,
		strings.TrimSpace(body),
		time.Now().UTC().Format(time.RFC3339Nano),
		messageID,
	)
	if err != nil {
		return fmt.Errorf("update panel message: %w", err)
	}
	return nil
}

func (s *Store) DeletePanelMessage(messageID int64) error {
	if messageID <= 0 {
		return errors.New("message id invalido")
	}
	if _, err := s.db.Exec(`DELETE FROM panel_message_reactions WHERE message_id = ?`, messageID); err != nil {
		return fmt.Errorf("delete panel reactions: %w", err)
	}
	if _, err := s.db.Exec(`DELETE FROM panel_message_pins WHERE message_id = ?`, messageID); err != nil {
		return fmt.Errorf("delete panel pins: %w", err)
	}
	if _, err := s.db.Exec(`DELETE FROM panel_message_favorites WHERE message_id = ?`, messageID); err != nil {
		return fmt.Errorf("delete panel favorites: %w", err)
	}
	if _, err := s.db.Exec(`DELETE FROM panel_messages WHERE id = ?`, messageID); err != nil {
		return fmt.Errorf("delete panel message: %w", err)
	}
	return nil
}

func (s *Store) TogglePanelMessagePin(messageID, roomID, userID int64) (bool, error) {
	if messageID <= 0 || roomID <= 0 || userID <= 0 {
		return false, errors.New("pin invalido")
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_message_pins WHERE message_id = ?`, messageID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check panel pin: %w", err)
	}
	if count > 0 {
		if _, err := s.db.Exec(`DELETE FROM panel_message_pins WHERE message_id = ?`, messageID); err != nil {
			return false, fmt.Errorf("delete panel pin: %w", err)
		}
		return false, nil
	}
	_, err := s.db.Exec(`
INSERT INTO panel_message_pins (message_id, room_id, pinned_by, created_at)
VALUES (?, ?, ?, ?)`,
		messageID,
		roomID,
		userID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return false, fmt.Errorf("insert panel pin: %w", err)
	}
	return true, nil
}

func (s *Store) TogglePanelMessageFavorite(messageID, userID int64) (bool, error) {
	if messageID <= 0 || userID <= 0 {
		return false, errors.New("favorito invalido")
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_message_favorites WHERE message_id = ? AND user_id = ?`, messageID, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check panel favorite: %w", err)
	}
	if count > 0 {
		if _, err := s.db.Exec(`DELETE FROM panel_message_favorites WHERE message_id = ? AND user_id = ?`, messageID, userID); err != nil {
			return false, fmt.Errorf("delete panel favorite: %w", err)
		}
		return false, nil
	}
	_, err := s.db.Exec(`
INSERT INTO panel_message_favorites (message_id, user_id, created_at)
VALUES (?, ?, ?)`,
		messageID,
		userID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return false, fmt.Errorf("insert panel favorite: %w", err)
	}
	return true, nil
}

func (s *Store) ListPinnedPanelMessagesForViewer(roomID int64, limit int, viewerID int64) ([]model.PanelMessage, error) {
	if roomID <= 0 {
		return nil, errors.New("room id invalido")
	}
	if limit <= 0 {
		limit = 6
	}
	rows, err := s.db.Query(`
SELECT pm.id, pm.room_id, pm.author_id, pm.author_name, pm.author_role, pm.body, pm.kind, pm.is_ai,
       pm.reply_to_id, pm.attachment_json, pm.created_at, pm.updated_at
FROM panel_message_pins pp
JOIN panel_messages pm ON pm.id = pp.message_id
WHERE pp.room_id = ?
ORDER BY pp.created_at DESC
LIMIT ?`, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("list panel pins: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelMessage, 0, limit)
	for rows.Next() {
		msg, err := scanPanelMessage(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel pins: %w", err)
	}
	return s.hydratePanelMessages(items, viewerID)
}

func (s *Store) ListLatestPanelMessagesForViewer(roomIDs []int64, viewerID int64) ([]model.PanelMessage, error) {
	roomIDs = uniqueInt64s(roomIDs)
	if len(roomIDs) == 0 {
		return []model.PanelMessage{}, nil
	}
	query, args := roomFilterQuery(`
SELECT pm.id, pm.room_id, pm.author_id, pm.author_name, pm.author_role, pm.body, pm.kind, pm.is_ai,
       pm.reply_to_id, pm.attachment_json, pm.created_at, pm.updated_at
FROM panel_messages pm
JOIN (
	SELECT room_id, MAX(id) AS max_id
	FROM panel_messages
	WHERE room_id IN (%s)
	GROUP BY room_id
) latest ON latest.max_id = pm.id
ORDER BY pm.room_id ASC`, roomIDs)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list latest panel messages: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelMessage, 0, len(roomIDs))
	for rows.Next() {
		msg, err := scanPanelMessage(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate latest panel messages: %w", err)
	}
	return s.hydratePanelMessages(items, viewerID)
}

func (s *Store) TogglePanelUserBlock(blockerID, blockedID int64) (bool, error) {
	if blockerID <= 0 || blockedID <= 0 || blockerID == blockedID {
		return false, errors.New("bloqueio invalido")
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_user_blocks WHERE blocker_id = ? AND blocked_id = ?`, blockerID, blockedID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check panel block: %w", err)
	}
	if count > 0 {
		if _, err := s.db.Exec(`DELETE FROM panel_user_blocks WHERE blocker_id = ? AND blocked_id = ?`, blockerID, blockedID); err != nil {
			return false, fmt.Errorf("delete panel block: %w", err)
		}
		return false, nil
	}
	_, err := s.db.Exec(`
INSERT INTO panel_user_blocks (blocker_id, blocked_id, created_at)
VALUES (?, ?, ?)`,
		blockerID,
		blockedID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return false, fmt.Errorf("insert panel block: %w", err)
	}
	return true, nil
}

func (s *Store) TogglePanelUserMute(muterID, mutedID int64) (bool, error) {
	if muterID <= 0 || mutedID <= 0 || muterID == mutedID {
		return false, errors.New("silencio invalido")
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_user_mutes WHERE muter_id = ? AND muted_id = ?`, muterID, mutedID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check panel mute: %w", err)
	}
	if count > 0 {
		if _, err := s.db.Exec(`DELETE FROM panel_user_mutes WHERE muter_id = ? AND muted_id = ?`, muterID, mutedID); err != nil {
			return false, fmt.Errorf("delete panel mute: %w", err)
		}
		return false, nil
	}
	_, err := s.db.Exec(`
INSERT INTO panel_user_mutes (muter_id, muted_id, created_at)
VALUES (?, ?, ?)`,
		muterID,
		mutedID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return false, fmt.Errorf("insert panel mute: %w", err)
	}
	return true, nil
}

func (s *Store) ListPanelBlockedIDs(blockerID int64) ([]int64, error) {
	if blockerID <= 0 {
		return []int64{}, nil
	}
	rows, err := s.db.Query(`SELECT blocked_id FROM panel_user_blocks WHERE blocker_id = ? ORDER BY blocked_id ASC`, blockerID)
	if err != nil {
		return nil, fmt.Errorf("list panel blocked ids: %w", err)
	}
	defer rows.Close()

	items := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan panel blocked id: %w", err)
		}
		items = append(items, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel blocked ids: %w", err)
	}
	return items, nil
}

func (s *Store) ListPanelMutedIDs(muterID int64) ([]int64, error) {
	if muterID <= 0 {
		return []int64{}, nil
	}
	rows, err := s.db.Query(`SELECT muted_id FROM panel_user_mutes WHERE muter_id = ? ORDER BY muted_id ASC`, muterID)
	if err != nil {
		return nil, fmt.Errorf("list panel muted ids: %w", err)
	}
	defer rows.Close()

	items := make([]int64, 0)
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan panel muted id: %w", err)
		}
		items = append(items, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel muted ids: %w", err)
	}
	return items, nil
}

func (s *Store) ListPanelBlockersForUser(userID int64) ([]int64, error) {
	if userID <= 0 {
		return []int64{}, nil
	}
	rows, err := s.db.Query(`SELECT blocker_id FROM panel_user_blocks WHERE blocked_id = ? ORDER BY blocker_id ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list panel blockers: %w", err)
	}
	defer rows.Close()

	items := make([]int64, 0)
	for rows.Next() {
		var blockerID int64
		if err := rows.Scan(&blockerID); err != nil {
			return nil, fmt.Errorf("scan panel blocker: %w", err)
		}
		items = append(items, blockerID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel blockers: %w", err)
	}
	return items, nil
}

func (s *Store) IsPanelUserBlocked(blockerID, blockedID int64) (bool, error) {
	if blockerID <= 0 || blockedID <= 0 {
		return false, errors.New("bloqueio invalido")
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_user_blocks WHERE blocker_id = ? AND blocked_id = ?`, blockerID, blockedID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check panel user blocked: %w", err)
	}
	return count > 0, nil
}

func (s *Store) TogglePanelReaction(messageID, userID int64, emoji string) error {
	if messageID <= 0 || userID <= 0 {
		return errors.New("reacao invalida")
	}
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return errors.New("emoji obrigatorio")
	}

	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_message_reactions WHERE message_id = ? AND user_id = ? AND emoji = ?`, messageID, userID, emoji)
	var count int
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("check panel reaction: %w", err)
	}
	if count > 0 {
		if _, err := s.db.Exec(`DELETE FROM panel_message_reactions WHERE message_id = ? AND user_id = ? AND emoji = ?`, messageID, userID, emoji); err != nil {
			return fmt.Errorf("delete panel reaction: %w", err)
		}
		return nil
	}

	_, err := s.db.Exec(`
INSERT INTO panel_message_reactions (message_id, user_id, emoji, created_at)
VALUES (?, ?, ?, ?)`,
		messageID,
		userID,
		emoji,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert panel reaction: %w", err)
	}
	return nil
}

func (s *Store) CreatePanelLog(item model.PanelLogItem) (model.PanelLogItem, error) {
	now := time.Now().UTC()
	result, err := s.db.Exec(`
INSERT INTO panel_logs (action, actor_id, actor_name, room_id, room_slug, detail, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		strings.TrimSpace(strings.ToLower(item.Action)),
		item.ActorID,
		strings.TrimSpace(item.ActorName),
		item.RoomID,
		strings.TrimSpace(strings.ToLower(item.RoomSlug)),
		strings.TrimSpace(item.Detail),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return model.PanelLogItem{}, fmt.Errorf("create panel log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.PanelLogItem{}, fmt.Errorf("fetch panel log id: %w", err)
	}
	item.ID = id
	item.CreatedAt = now
	return item, nil
}

func (s *Store) ListPanelLogs(limit int) ([]model.PanelLogItem, error) {
	if limit <= 0 {
		limit = 40
	}
	rows, err := s.db.Query(`
SELECT id, action, actor_id, actor_name, room_id, room_slug, detail, created_at
FROM panel_logs
ORDER BY id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list panel logs: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelLogItem, 0, limit)
	for rows.Next() {
		var item model.PanelLogItem
		var createdAt string
		if err := rows.Scan(
			&item.ID,
			&item.Action,
			&item.ActorID,
			&item.ActorName,
			&item.RoomID,
			&item.RoomSlug,
			&item.Detail,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan panel log: %w", err)
		}
		ts, err := time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse panel log created_at: %w", err)
		}
		item.CreatedAt = ts
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel logs: %w", err)
	}
	return items, nil
}

func (s *Store) CreatePanelEvent(input model.PanelEvent) (model.PanelEvent, error) {
	if strings.TrimSpace(input.Title) == "" {
		return model.PanelEvent{}, errors.New("titulo do evento obrigatorio")
	}
	now := time.Now().UTC()
	result, err := s.db.Exec(`
INSERT INTO panel_events (title, description, room_id, created_by, created_by_name, starts_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Description),
		input.RoomID,
		input.CreatedBy,
		strings.TrimSpace(input.CreatedByName),
		input.StartsAt.UTC().Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return model.PanelEvent{}, fmt.Errorf("create panel event: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.PanelEvent{}, fmt.Errorf("fetch panel event id: %w", err)
	}
	return s.GetPanelEventForViewer(id, input.CreatedBy)
}

func (s *Store) GetPanelEventForViewer(eventID, viewerID int64) (model.PanelEvent, error) {
	row := s.db.QueryRow(`
SELECT e.id, e.title, e.description, e.room_id, COALESCE(r.name, ''), e.created_by, e.created_by_name,
       e.starts_at, e.created_at, COUNT(er.user_id) AS rsvp_count,
       MAX(CASE WHEN er.user_id = ? THEN 1 ELSE 0 END) AS viewer_joined
FROM panel_events e
LEFT JOIN panel_rooms r ON r.id = e.room_id
LEFT JOIN panel_event_rsvps er ON er.event_id = e.id
WHERE e.id = ?
GROUP BY e.id, e.title, e.description, e.room_id, r.name, e.created_by, e.created_by_name, e.starts_at, e.created_at`,
		viewerID,
		eventID,
	)
	event, err := scanPanelEvent(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelEvent{}, fmt.Errorf("panel event %d not found", eventID)
		}
		return model.PanelEvent{}, err
	}
	return event, nil
}

func (s *Store) ListUpcomingPanelEventsForViewer(viewerID int64, since time.Time, limit int) ([]model.PanelEvent, error) {
	if limit <= 0 {
		limit = 8
	}
	rows, err := s.db.Query(`
SELECT e.id, e.title, e.description, e.room_id, COALESCE(r.name, ''), e.created_by, e.created_by_name,
       e.starts_at, e.created_at, COUNT(er.user_id) AS rsvp_count,
       MAX(CASE WHEN er.user_id = ? THEN 1 ELSE 0 END) AS viewer_joined
FROM panel_events e
LEFT JOIN panel_rooms r ON r.id = e.room_id
LEFT JOIN panel_event_rsvps er ON er.event_id = e.id
WHERE e.starts_at >= ?
GROUP BY e.id, e.title, e.description, e.room_id, r.name, e.created_by, e.created_by_name, e.starts_at, e.created_at
ORDER BY e.starts_at ASC, e.id DESC
LIMIT ?`,
		viewerID,
		since.UTC().Format(time.RFC3339Nano),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list panel events: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelEvent, 0, limit)
	for rows.Next() {
		event, err := scanPanelEvent(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel events: %w", err)
	}
	return items, nil
}

func (s *Store) TogglePanelEventRSVP(eventID, userID int64) (bool, error) {
	if eventID <= 0 || userID <= 0 {
		return false, errors.New("rsvp invalido")
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_event_rsvps WHERE event_id = ? AND user_id = ?`, eventID, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check panel event rsvp: %w", err)
	}
	if count > 0 {
		if _, err := s.db.Exec(`DELETE FROM panel_event_rsvps WHERE event_id = ? AND user_id = ?`, eventID, userID); err != nil {
			return false, fmt.Errorf("delete panel event rsvp: %w", err)
		}
		return false, nil
	}
	_, err := s.db.Exec(`
INSERT INTO panel_event_rsvps (event_id, user_id, created_at)
VALUES (?, ?, ?)`,
		eventID,
		userID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return false, fmt.Errorf("insert panel event rsvp: %w", err)
	}
	return true, nil
}

func (s *Store) DeletePanelEvent(eventID int64) error {
	if eventID <= 0 {
		return errors.New("evento invalido")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin delete panel event tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM panel_event_rsvps WHERE event_id = ?`, eventID); err != nil {
		return fmt.Errorf("delete panel event rsvps: %w", err)
	}
	result, err := tx.Exec(`DELETE FROM panel_events WHERE id = ?`, eventID)
	if err != nil {
		return fmt.Errorf("delete panel event: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("panel event rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("panel event %d not found", eventID)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete panel event: %w", err)
	}
	return nil
}

func (s *Store) CreatePanelPoll(input model.PanelPoll) (model.PanelPoll, error) {
	if input.RoomID <= 0 {
		return model.PanelPoll{}, errors.New("room id invalido")
	}
	if strings.TrimSpace(input.Question) == "" {
		return model.PanelPoll{}, errors.New("pergunta da enquete obrigatoria")
	}
	if len(input.Options) < 2 {
		return model.PanelPoll{}, errors.New("a enquete precisa de pelo menos duas opcoes")
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := s.db.Begin()
	if err != nil {
		return model.PanelPoll{}, fmt.Errorf("begin panel poll tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
INSERT INTO panel_polls (room_id, question, created_by, created_by_name, created_at)
VALUES (?, ?, ?, ?, ?)`,
		input.RoomID,
		strings.TrimSpace(input.Question),
		input.CreatedBy,
		strings.TrimSpace(input.CreatedByName),
		now,
	)
	if err != nil {
		return model.PanelPoll{}, fmt.Errorf("create panel poll: %w", err)
	}
	pollID, err := result.LastInsertId()
	if err != nil {
		return model.PanelPoll{}, fmt.Errorf("fetch panel poll id: %w", err)
	}

	for index, option := range input.Options {
		if _, err := tx.Exec(`
INSERT INTO panel_poll_options (poll_id, label, sort_order)
VALUES (?, ?, ?)`,
			pollID,
			strings.TrimSpace(option.Label),
			index,
		); err != nil {
			return model.PanelPoll{}, fmt.Errorf("create panel poll option: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return model.PanelPoll{}, fmt.Errorf("commit panel poll: %w", err)
	}
	return s.GetPanelPollForViewer(pollID, input.CreatedBy)
}

func (s *Store) GetPanelPollForViewer(pollID, viewerID int64) (model.PanelPoll, error) {
	rows, err := s.listPanelPollRowsByQuery(`
SELECT id, room_id, question, created_by, created_by_name, created_at
FROM panel_polls
WHERE id = ?
LIMIT 1`, []any{pollID}, viewerID)
	if err != nil {
		return model.PanelPoll{}, err
	}
	if len(rows) == 0 {
		return model.PanelPoll{}, fmt.Errorf("panel poll %d not found", pollID)
	}
	return rows[0], nil
}

func (s *Store) ListPanelPollsForViewer(roomID, viewerID int64, limit int) ([]model.PanelPoll, error) {
	if roomID <= 0 {
		return nil, errors.New("room id invalido")
	}
	if limit <= 0 {
		limit = 12
	}
	return s.listPanelPollRowsByQuery(`
SELECT id, room_id, question, created_by, created_by_name, created_at
FROM panel_polls
WHERE room_id = ?
ORDER BY id DESC
LIMIT ?`, []any{roomID, limit}, viewerID)
}

func (s *Store) listPanelPollRowsByQuery(query string, args []any, viewerID int64) ([]model.PanelPoll, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list panel polls: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelPoll, 0)
	for rows.Next() {
		poll, err := scanPanelPoll(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, poll)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel polls: %w", err)
	}
	return s.hydratePanelPolls(items, viewerID)
}

func (s *Store) TogglePanelPollVote(pollID, optionID, userID int64) (bool, error) {
	if pollID <= 0 || optionID <= 0 || userID <= 0 {
		return false, errors.New("voto invalido")
	}
	row := s.db.QueryRow(`SELECT COUNT(1) FROM panel_poll_options WHERE id = ? AND poll_id = ?`, optionID, pollID)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check panel poll option: %w", err)
	}
	if count == 0 {
		return false, errors.New("opcao da enquete nao encontrada")
	}

	current := s.db.QueryRow(`SELECT option_id FROM panel_poll_votes WHERE poll_id = ? AND user_id = ?`, pollID, userID)
	var currentOptionID int64
	switch err := current.Scan(&currentOptionID); {
	case errors.Is(err, sql.ErrNoRows):
		_, err = s.db.Exec(`
INSERT INTO panel_poll_votes (poll_id, option_id, user_id, created_at)
VALUES (?, ?, ?, ?)`,
			pollID,
			optionID,
			userID,
			time.Now().UTC().Format(time.RFC3339Nano),
		)
		if err != nil {
			return false, fmt.Errorf("insert panel poll vote: %w", err)
		}
		return true, nil
	case err != nil:
		return false, fmt.Errorf("check panel poll vote: %w", err)
	}

	if currentOptionID == optionID {
		if _, err := s.db.Exec(`DELETE FROM panel_poll_votes WHERE poll_id = ? AND user_id = ?`, pollID, userID); err != nil {
			return false, fmt.Errorf("delete panel poll vote: %w", err)
		}
		return false, nil
	}

	_, err := s.db.Exec(`
INSERT INTO panel_poll_votes (poll_id, option_id, user_id, created_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(poll_id, user_id) DO UPDATE SET
	option_id = excluded.option_id,
	created_at = excluded.created_at`,
		pollID,
		optionID,
		userID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return false, fmt.Errorf("upsert panel poll vote: %w", err)
	}
	return true, nil
}

func (s *Store) DeletePanelPoll(pollID int64) error {
	if pollID <= 0 {
		return errors.New("enquete invalida")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin delete panel poll tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM panel_poll_votes WHERE poll_id = ?`, pollID); err != nil {
		return fmt.Errorf("delete panel poll votes: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM panel_poll_options WHERE poll_id = ?`, pollID); err != nil {
		return fmt.Errorf("delete panel poll options: %w", err)
	}
	result, err := tx.Exec(`DELETE FROM panel_polls WHERE id = ?`, pollID)
	if err != nil {
		return fmt.Errorf("delete panel poll: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("panel poll rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("panel poll %d not found", pollID)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete panel poll: %w", err)
	}
	return nil
}

func (s *Store) SearchPanel(query string, roomIDs []int64, viewerID int64, limit int) (model.PanelSearchResult, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 {
		limit = 20
	}
	result := model.PanelSearchResult{
		Query:    query,
		Rooms:    []model.PanelRoom{},
		Users:    []model.PanelUser{},
		Messages: []model.PanelMessage{},
	}
	if query == "" {
		return result, nil
	}

	pattern := "%" + strings.ToLower(query) + "%"

	userRows, err := s.db.Query(`
SELECT id, username, email, display_name, role, theme, accent_color, avatar_url, bio, status, status_text,
       created_by, created_at, updated_at, last_login_at, password_hash
FROM panel_users
WHERE LOWER(username) LIKE ? OR LOWER(display_name) LIKE ? OR LOWER(email) LIKE ?
ORDER BY CASE role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 WHEN 'vip' THEN 2 ELSE 3 END, display_name ASC
LIMIT ?`, pattern, pattern, pattern, limit)
	if err != nil {
		return result, fmt.Errorf("search panel users: %w", err)
	}
	defer userRows.Close()
	for userRows.Next() {
		user, err := scanPanelUser(userRows.Scan)
		if err != nil {
			return result, err
		}
		user.PasswordHash = ""
		result.Users = append(result.Users, user)
	}
	if err := userRows.Err(); err != nil {
		return result, fmt.Errorf("iterate searched users: %w", err)
	}

	roomIDs = uniqueInt64s(roomIDs)
	if len(roomIDs) == 0 {
		return result, nil
	}

	roomQuery, roomArgs := roomFilterQuery(`
SELECT r.id, r.slug, r.name, r.description, r.icon, r.category, r.scope, r.sort_order, r.admin_only, r.vip_only,
       r.password_hash, r.created_at, r.updated_at,
       m.created_at, m.body
FROM panel_rooms r
LEFT JOIN panel_messages m ON m.id = (
	SELECT pm.id FROM panel_messages pm WHERE pm.room_id = r.id ORDER BY pm.id DESC LIMIT 1
)
WHERE r.id IN (%s) AND (LOWER(r.name) LIKE ? OR LOWER(r.description) LIKE ?)
ORDER BY r.sort_order ASC, r.name ASC
LIMIT ?`, roomIDs, pattern, pattern, limit)

	roomRows, err := s.db.Query(roomQuery, roomArgs...)
	if err != nil {
		return result, fmt.Errorf("search panel rooms: %w", err)
	}
	defer roomRows.Close()
	for roomRows.Next() {
		room, err := scanPanelRoom(roomRows.Scan)
		if err != nil {
			return result, err
		}
		result.Rooms = append(result.Rooms, room)
	}
	if err := roomRows.Err(); err != nil {
		return result, fmt.Errorf("iterate searched rooms: %w", err)
	}

	messageQuery, messageArgs := roomFilterQuery(`
SELECT id, room_id, author_id, author_name, author_role, body, kind, is_ai, reply_to_id, attachment_json, created_at
       , updated_at
FROM panel_messages
WHERE room_id IN (%s) AND (LOWER(body) LIKE ? OR LOWER(author_name) LIKE ? OR LOWER(attachment_json) LIKE ?)
ORDER BY id DESC
LIMIT ?`, roomIDs, pattern, pattern, pattern, limit)
	messageRows, err := s.db.Query(messageQuery, messageArgs...)
	if err != nil {
		return result, fmt.Errorf("search panel messages: %w", err)
	}
	defer messageRows.Close()
	foundMessages := make([]model.PanelMessage, 0, limit)
	for messageRows.Next() {
		msg, err := scanPanelMessage(messageRows.Scan)
		if err != nil {
			return result, err
		}
		foundMessages = append(foundMessages, msg)
	}
	if err := messageRows.Err(); err != nil {
		return result, fmt.Errorf("iterate searched messages: %w", err)
	}
	result.Messages, err = s.hydratePanelMessages(foundMessages, viewerID)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Store) hydratePanelPolls(items []model.PanelPoll, viewerID int64) ([]model.PanelPoll, error) {
	if len(items) == 0 {
		return items, nil
	}

	pollIDs := make([]int64, 0, len(items))
	indexByID := make(map[int64]int, len(items))
	for index, item := range items {
		pollIDs = append(pollIDs, item.ID)
		indexByID[item.ID] = index
		items[index].Options = []model.PanelPollOption{}
	}

	query, args := idsQuery(`
SELECT o.id, o.poll_id, o.label, o.sort_order, COUNT(v.user_id) AS vote_count,
       MAX(CASE WHEN v.user_id = ? THEN 1 ELSE 0 END) AS viewer_voted
FROM panel_poll_options o
LEFT JOIN panel_poll_votes v ON v.option_id = o.id
WHERE o.poll_id IN (%s)
GROUP BY o.id, o.poll_id, o.label, o.sort_order
ORDER BY o.poll_id ASC, o.sort_order ASC, o.id ASC`, pollIDs)
	args = append([]any{viewerID}, args...)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("load panel poll options: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			option      model.PanelPollOption
			pollID      int64
			sortOrder   int
			viewerVoted int
		)
		if err := rows.Scan(&option.ID, &pollID, &option.Label, &sortOrder, &option.Votes, &viewerVoted); err != nil {
			return nil, fmt.Errorf("scan panel poll option: %w", err)
		}
		option.ViewerVoted = viewerVoted != 0
		index, ok := indexByID[pollID]
		if !ok {
			continue
		}
		if option.ViewerVoted {
			items[index].ViewerOptionID = option.ID
		}
		items[index].TotalVotes += option.Votes
		items[index].Options = append(items[index].Options, option)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel poll options: %w", err)
	}
	return items, nil
}

func (s *Store) hydratePanelMessages(items []model.PanelMessage, viewerID int64) ([]model.PanelMessage, error) {
	if len(items) == 0 {
		return items, nil
	}
	replyMap, err := s.loadPanelReplyPreviews(items)
	if err != nil {
		return nil, err
	}
	reactionMap, err := s.loadPanelMessageReactions(items, viewerID)
	if err != nil {
		return nil, err
	}
	pinMap, err := s.loadPanelMessagePins(items)
	if err != nil {
		return nil, err
	}
	favoriteMap, err := s.loadPanelMessageFavorites(items, viewerID)
	if err != nil {
		return nil, err
	}

	for i := range items {
		if preview, ok := replyMap[items[i].ReplyToID]; ok {
			items[i].Reply = preview
		}
		if reactions, ok := reactionMap[items[i].ID]; ok {
			items[i].Reactions = reactions
		}
		items[i].IsPinned = pinMap[items[i].ID]
		items[i].ViewerFavorited = favoriteMap[items[i].ID]
	}
	return items, nil
}

func (s *Store) loadPanelReplyPreviews(items []model.PanelMessage) (map[int64]*model.PanelReplyPreview, error) {
	ids := make([]int64, 0)
	for _, item := range items {
		if item.ReplyToID > 0 {
			ids = append(ids, item.ReplyToID)
		}
	}
	ids = uniqueInt64s(ids)
	if len(ids) == 0 {
		return map[int64]*model.PanelReplyPreview{}, nil
	}

	query, args := idsQuery(`
SELECT id, author_name, body, attachment_json
FROM panel_messages
WHERE id IN (%s)`, ids)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("load panel reply previews: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]*model.PanelReplyPreview, len(ids))
	for rows.Next() {
		var (
			id             int64
			authorName     string
			body           string
			attachmentJSON string
		)
		if err := rows.Scan(&id, &authorName, &body, &attachmentJSON); err != nil {
			return nil, fmt.Errorf("scan panel reply preview: %w", err)
		}
		preview := previewBody(body)
		if preview == "" && strings.TrimSpace(attachmentJSON) != "" {
			var attachment model.PanelAttachment
			if err := json.Unmarshal([]byte(attachmentJSON), &attachment); err == nil {
				preview = "[anexo] " + attachment.Name
			}
		}
		result[id] = &model.PanelReplyPreview{
			MessageID:   id,
			AuthorName:  authorName,
			BodyPreview: preview,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel reply previews: %w", err)
	}
	return result, nil
}

func (s *Store) loadPanelMessageReactions(items []model.PanelMessage, viewerID int64) (map[int64][]model.PanelReaction, error) {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	ids = uniqueInt64s(ids)
	if len(ids) == 0 {
		return map[int64][]model.PanelReaction{}, nil
	}

	query, args := idsQuery(`
SELECT message_id, emoji, COUNT(1), MAX(CASE WHEN user_id = ? THEN 1 ELSE 0 END) AS viewer_reacted
FROM panel_message_reactions
WHERE message_id IN (%s)
GROUP BY message_id, emoji
ORDER BY message_id ASC, COUNT(1) DESC, emoji ASC`, ids)
	args = append([]any{viewerID}, args...)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("load panel reactions: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]model.PanelReaction, len(ids))
	for rows.Next() {
		var (
			messageID      int64
			emoji          string
			count          int
			viewerReactedI int
		)
		if err := rows.Scan(&messageID, &emoji, &count, &viewerReactedI); err != nil {
			return nil, fmt.Errorf("scan panel reaction: %w", err)
		}
		result[messageID] = append(result[messageID], model.PanelReaction{
			Emoji:         emoji,
			Count:         count,
			ViewerReacted: viewerReactedI != 0,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel reactions: %w", err)
	}
	return result, nil
}

func (s *Store) loadPanelMessagePins(items []model.PanelMessage) (map[int64]bool, error) {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	ids = uniqueInt64s(ids)
	if len(ids) == 0 {
		return map[int64]bool{}, nil
	}

	query, args := idsQuery(`
SELECT message_id
FROM panel_message_pins
WHERE message_id IN (%s)`, ids)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("load panel pins: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]bool, len(ids))
	for rows.Next() {
		var messageID int64
		if err := rows.Scan(&messageID); err != nil {
			return nil, fmt.Errorf("scan panel pin: %w", err)
		}
		result[messageID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel pins: %w", err)
	}
	return result, nil
}

func (s *Store) loadPanelMessageFavorites(items []model.PanelMessage, viewerID int64) (map[int64]bool, error) {
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	ids = uniqueInt64s(ids)
	if len(ids) == 0 || viewerID <= 0 {
		return map[int64]bool{}, nil
	}

	query, args := idsQuery(`
SELECT message_id
FROM panel_message_favorites
WHERE user_id = ? AND message_id IN (%s)`, ids)
	args = append([]any{viewerID}, args...)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("load panel favorites: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]bool, len(ids))
	for rows.Next() {
		var messageID int64
		if err := rows.Scan(&messageID); err != nil {
			return nil, fmt.Errorf("scan panel favorite: %w", err)
		}
		result[messageID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate panel favorites: %w", err)
	}
	return result, nil
}

func scanPanelUser(scan func(dest ...any) error) (model.PanelUser, error) {
	var (
		item        model.PanelUser
		createdAt   string
		updatedAt   string
		lastLoginAt sql.NullString
	)
	if err := scan(
		&item.ID,
		&item.Username,
		&item.Email,
		&item.DisplayName,
		&item.Role,
		&item.Theme,
		&item.AccentColor,
		&item.AvatarURL,
		&item.Bio,
		&item.Status,
		&item.StatusText,
		&item.CreatedBy,
		&createdAt,
		&updatedAt,
		&lastLoginAt,
		&item.PasswordHash,
	); err != nil {
		return model.PanelUser{}, err
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return model.PanelUser{}, fmt.Errorf("parse panel user created_at: %w", err)
	}
	updated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return model.PanelUser{}, fmt.Errorf("parse panel user updated_at: %w", err)
	}
	item.CreatedAt = created
	item.UpdatedAt = updated
	if lastLoginAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, lastLoginAt.String)
		if err != nil {
			return model.PanelUser{}, fmt.Errorf("parse panel user last_login_at: %w", err)
		}
		item.LastLoginAt = &parsed
	}
	return item, nil
}

func scanPanelRoom(scan func(dest ...any) error) (model.PanelRoom, error) {
	var (
		item            model.PanelRoom
		adminOnly       int
		vipOnly         int
		createdAt       string
		updatedAt       string
		lastMessageAt   sql.NullString
		lastMessageBody sql.NullString
	)
	if err := scan(
		&item.ID,
		&item.Slug,
		&item.Name,
		&item.Description,
		&item.Icon,
		&item.Category,
		&item.Scope,
		&item.SortOrder,
		&adminOnly,
		&vipOnly,
		&item.PasswordHash,
		&createdAt,
		&updatedAt,
		&lastMessageAt,
		&lastMessageBody,
	); err != nil {
		return model.PanelRoom{}, err
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return model.PanelRoom{}, fmt.Errorf("parse panel room created_at: %w", err)
	}
	updated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return model.PanelRoom{}, fmt.Errorf("parse panel room updated_at: %w", err)
	}
	item.CreatedAt = created
	item.UpdatedAt = updated
	item.Scope = strings.TrimSpace(strings.ToLower(item.Scope))
	if item.Scope == "" {
		item.Scope = "public"
	}
	item.AdminOnly = intToBool(adminOnly)
	item.VIPOnly = intToBool(vipOnly)
	item.PasswordProtected = strings.TrimSpace(item.PasswordHash) != ""
	if lastMessageAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, lastMessageAt.String)
		if err != nil {
			return model.PanelRoom{}, fmt.Errorf("parse panel room last_message_at: %w", err)
		}
		item.LastMessageAt = &parsed
	}
	if lastMessageBody.Valid {
		item.LastMessagePreview = lastMessageBody.String
	}
	return item, nil
}

func scanPanelMessage(scan func(dest ...any) error) (model.PanelMessage, error) {
	var (
		item           model.PanelMessage
		isAI           int
		replyToID      int64
		attachmentJSON string
		createdAt      string
		updatedAt      string
	)
	if err := scan(
		&item.ID,
		&item.RoomID,
		&item.AuthorID,
		&item.AuthorName,
		&item.AuthorRole,
		&item.Body,
		&item.Kind,
		&isAI,
		&replyToID,
		&attachmentJSON,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.PanelMessage{}, err
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return model.PanelMessage{}, fmt.Errorf("parse panel message created_at: %w", err)
	}
	item.CreatedAt = created
	item.IsAI = intToBool(isAI)
	item.ReplyToID = replyToID
	if strings.TrimSpace(updatedAt) != "" {
		parsed, err := time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return model.PanelMessage{}, fmt.Errorf("parse panel message updated_at: %w", err)
		}
		if !parsed.Equal(created) {
			item.UpdatedAt = &parsed
		}
	}
	if strings.TrimSpace(attachmentJSON) != "" {
		var attachment model.PanelAttachment
		if err := json.Unmarshal([]byte(attachmentJSON), &attachment); err != nil {
			return model.PanelMessage{}, fmt.Errorf("decode panel attachment: %w", err)
		}
		item.Attachment = &attachment
	}
	return item, nil
}

func scanPanelEvent(scan func(dest ...any) error) (model.PanelEvent, error) {
	var (
		item         model.PanelEvent
		roomName     string
		startsAtRaw  string
		createdAtRaw string
		viewerJoined int
	)
	if err := scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.RoomID,
		&roomName,
		&item.CreatedBy,
		&item.CreatedByName,
		&startsAtRaw,
		&createdAtRaw,
		&item.RSVPCount,
		&viewerJoined,
	); err != nil {
		return model.PanelEvent{}, err
	}
	startsAt, err := time.Parse(time.RFC3339Nano, startsAtRaw)
	if err != nil {
		return model.PanelEvent{}, fmt.Errorf("parse panel event starts_at: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtRaw)
	if err != nil {
		return model.PanelEvent{}, fmt.Errorf("parse panel event created_at: %w", err)
	}
	item.RoomName = roomName
	item.StartsAt = startsAt
	item.CreatedAt = createdAt
	item.ViewerJoined = viewerJoined != 0
	return item, nil
}

func scanPanelPoll(scan func(dest ...any) error) (model.PanelPoll, error) {
	var (
		item         model.PanelPoll
		createdAtRaw string
	)
	if err := scan(
		&item.ID,
		&item.RoomID,
		&item.Question,
		&item.CreatedBy,
		&item.CreatedByName,
		&createdAtRaw,
	); err != nil {
		return model.PanelPoll{}, err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtRaw)
	if err != nil {
		return model.PanelPoll{}, fmt.Errorf("parse panel poll created_at: %w", err)
	}
	item.CreatedAt = createdAt
	return item, nil
}

func previewBody(body string) string {
	body = strings.TrimSpace(body)
	if len(body) <= 120 {
		return body
	}
	return body[:117] + "..."
}

func roomFilterQuery(template string, roomIDs []int64, extraArgs ...any) (string, []any) {
	args := append(int64ToAny(roomIDs), extraArgs...)
	return fmt.Sprintf(template, placeholders(len(roomIDs))), args
}

func idsQuery(template string, ids []int64, extraArgs ...any) (string, []any) {
	args := append(int64ToAny(ids), extraArgs...)
	return fmt.Sprintf(template, placeholders(len(ids))), args
}

func uniqueInt64s(values []int64) []int64 {
	if len(values) == 0 {
		return values
	}
	seen := make(map[int64]struct{}, len(values))
	out := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func placeholders(count int) string {
	if count <= 0 {
		return "0"
	}
	parts := make([]string, count)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
}

func int64ToAny(values []int64) []any {
	args := make([]any, 0, len(values))
	for _, value := range values {
		args = append(args, value)
	}
	return args
}

func appendAny(base []any, extra ...any) []any {
	if len(extra) == 0 {
		return base
	}
	out := make([]any, 0, len(base)+len(extra))
	out = append(out, base...)
	out = append(out, extra...)
	return out
}
