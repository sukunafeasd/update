package db

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"universald/internal/model"
)

func (s *Store) UpdatePanelUserRole(userID int64, role string) (model.PanelUser, error) {
	if userID <= 0 {
		return model.PanelUser{}, errors.New("user id invalido")
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		role = "member"
	}
	_, err := s.db.Exec(`
UPDATE panel_users
SET role = ?, updated_at = ?
WHERE id = ?`,
		role,
		time.Now().UTC().Format(time.RFC3339Nano),
		userID,
	)
	if err != nil {
		return model.PanelUser{}, fmt.Errorf("update panel user role: %w", err)
	}
	return s.GetPanelUserByID(userID)
}

func (s *Store) DeletePanelUserCascade(userID int64) error {
	if userID <= 0 {
		return errors.New("user id invalido")
	}
	steps := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM panel_sessions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM panel_presence WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM panel_room_members WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM panel_user_blocks WHERE blocker_id = ? OR blocked_id = ?`, []any{userID, userID}},
		{`DELETE FROM panel_user_mutes WHERE muter_id = ? OR muted_id = ?`, []any{userID, userID}},
		{`DELETE FROM panel_message_reactions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM panel_message_favorites WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM panel_event_rsvps WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM panel_poll_votes WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM panel_join_requests WHERE approved_user_id = ?`, []any{userID}},
		{`DELETE FROM panel_users WHERE id = ?`, []any{userID}},
	}
	for _, step := range steps {
		if _, err := s.db.Exec(step.query, step.args...); err != nil {
			return fmt.Errorf("delete panel user cascade: %w", err)
		}
	}
	return nil
}
