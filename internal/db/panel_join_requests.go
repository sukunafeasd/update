package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"universald/internal/model"
)

func (s *Store) CreatePanelJoinRequest(input model.PanelJoinRequest) (model.PanelJoinRequest, error) {
	now := time.Now().UTC()
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return model.PanelJoinRequest{}, errors.New("email obrigatorio")
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = "Novo membro"
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = "pending"
	}
	result, err := s.db.Exec(`
INSERT INTO panel_join_requests (
	email, display_name, note, status, review_note, requested_at, reviewed_at,
	reviewed_by, reviewer_name, access_code, access_code_expires, approved_user_id, email_sent
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		email,
		displayName,
		strings.TrimSpace(input.Note),
		status,
		strings.TrimSpace(input.ReviewNote),
		now.Format(time.RFC3339Nano),
		timeValueString(input.ReviewedAt),
		input.ReviewedBy,
		strings.TrimSpace(input.ReviewerName),
		strings.TrimSpace(input.AccessCode),
		timeValueString(input.AccessCodeExpires),
		input.ApprovedUserID,
		boolToInt(input.EmailSent),
	)
	if err != nil {
		return model.PanelJoinRequest{}, fmt.Errorf("create join request: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return model.PanelJoinRequest{}, fmt.Errorf("fetch join request id: %w", err)
	}
	return s.GetPanelJoinRequestByID(id)
}

func (s *Store) GetPanelJoinRequestByID(id int64) (model.PanelJoinRequest, error) {
	row := s.db.QueryRow(`
SELECT id, email, display_name, note, status, review_note, requested_at, reviewed_at,
       reviewed_by, reviewer_name, access_code, access_code_expires, approved_user_id, email_sent
FROM panel_join_requests
WHERE id = ?
LIMIT 1`, id)
	item, err := scanPanelJoinRequest(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelJoinRequest{}, fmt.Errorf("join request %d not found", id)
		}
		return model.PanelJoinRequest{}, err
	}
	return item, nil
}

func (s *Store) GetLatestPanelJoinRequestByEmail(email string) (model.PanelJoinRequest, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	row := s.db.QueryRow(`
SELECT id, email, display_name, note, status, review_note, requested_at, reviewed_at,
       reviewed_by, reviewer_name, access_code, access_code_expires, approved_user_id, email_sent
FROM panel_join_requests
WHERE email = ?
ORDER BY id DESC
LIMIT 1`, email)
	item, err := scanPanelJoinRequest(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PanelJoinRequest{}, fmt.Errorf("join request %q not found", email)
		}
		return model.PanelJoinRequest{}, err
	}
	return item, nil
}

func (s *Store) ListPanelJoinRequests(limit int) ([]model.PanelJoinRequest, error) {
	if limit <= 0 {
		limit = 80
	}
	rows, err := s.db.Query(`
SELECT id, email, display_name, note, status, review_note, requested_at, reviewed_at,
       reviewed_by, reviewer_name, access_code, access_code_expires, approved_user_id, email_sent
FROM panel_join_requests
ORDER BY CASE status WHEN 'pending' THEN 0 WHEN 'approved' THEN 1 WHEN 'rejected' THEN 2 ELSE 3 END, id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	defer rows.Close()

	items := make([]model.PanelJoinRequest, 0)
	for rows.Next() {
		item, scanErr := scanPanelJoinRequest(rows.Scan)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate join requests: %w", err)
	}
	return items, nil
}

func (s *Store) UpdatePanelJoinRequest(input model.PanelJoinRequest) (model.PanelJoinRequest, error) {
	if input.ID <= 0 {
		return model.PanelJoinRequest{}, errors.New("join request invalido")
	}
	_, err := s.db.Exec(`
UPDATE panel_join_requests
SET email = ?, display_name = ?, note = ?, status = ?, review_note = ?, reviewed_at = ?,
    reviewed_by = ?, reviewer_name = ?, access_code = ?, access_code_expires = ?, approved_user_id = ?, email_sent = ?
WHERE id = ?`,
		strings.ToLower(strings.TrimSpace(input.Email)),
		strings.TrimSpace(input.DisplayName),
		strings.TrimSpace(input.Note),
		strings.ToLower(strings.TrimSpace(input.Status)),
		strings.TrimSpace(input.ReviewNote),
		timeValueString(input.ReviewedAt),
		input.ReviewedBy,
		strings.TrimSpace(input.ReviewerName),
		strings.TrimSpace(input.AccessCode),
		timeValueString(input.AccessCodeExpires),
		input.ApprovedUserID,
		boolToInt(input.EmailSent),
		input.ID,
	)
	if err != nil {
		return model.PanelJoinRequest{}, fmt.Errorf("update join request: %w", err)
	}
	return s.GetPanelJoinRequestByID(input.ID)
}

func scanPanelJoinRequest(scan func(dest ...any) error) (model.PanelJoinRequest, error) {
	var item model.PanelJoinRequest
	var requestedAt string
	var reviewedAt string
	var accessExpires string
	var emailSent int
	if err := scan(
		&item.ID,
		&item.Email,
		&item.DisplayName,
		&item.Note,
		&item.Status,
		&item.ReviewNote,
		&requestedAt,
		&reviewedAt,
		&item.ReviewedBy,
		&item.ReviewerName,
		&item.AccessCode,
		&accessExpires,
		&item.ApprovedUserID,
		&emailSent,
	); err != nil {
		return model.PanelJoinRequest{}, err
	}
	parsedRequested, err := time.Parse(time.RFC3339Nano, requestedAt)
	if err != nil {
		return model.PanelJoinRequest{}, fmt.Errorf("parse join request requested_at: %w", err)
	}
	item.RequestedAt = parsedRequested
	if reviewedAt = strings.TrimSpace(reviewedAt); reviewedAt != "" {
		parsedReviewed, parseErr := time.Parse(time.RFC3339Nano, reviewedAt)
		if parseErr != nil {
			return model.PanelJoinRequest{}, fmt.Errorf("parse join request reviewed_at: %w", parseErr)
		}
		item.ReviewedAt = &parsedReviewed
	}
	if accessExpires = strings.TrimSpace(accessExpires); accessExpires != "" {
		parsedExpires, parseErr := time.Parse(time.RFC3339Nano, accessExpires)
		if parseErr != nil {
			return model.PanelJoinRequest{}, fmt.Errorf("parse join request access_code_expires: %w", parseErr)
		}
		item.AccessCodeExpires = &parsedExpires
	}
	item.EmailSent = intToBool(emailSent)
	return item, nil
}

func timeValueString(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
