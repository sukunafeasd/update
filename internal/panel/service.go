package panel

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"universald/internal/db"
	"universald/internal/model"
)

const (
	defaultOwnerUsername = "dief"
	defaultOwnerEmail    = "paineldief@local"
	defaultOwnerPassword = "PainelDief#2026"
	defaultPrivatePass   = "segredo123"
	defaultVIPPass       = "vip2026"
	sessionLifetime      = 7 * 24 * time.Hour
	presenceWindow       = 45 * time.Second
	passwordIterations   = 90000
	typingTTL            = 7 * time.Second
	floodWindow          = 12 * time.Second
	floodBurst           = 6
	loginWindow          = 90 * time.Second
	loginBurst           = 5
	loginCooldown        = 45 * time.Second
	recentLogLimit       = 14
	defaultUploadLimit   = 30 << 20
	vipUploadLimit       = 60 << 20
	adminUploadLimit     = 120 << 20
)

var ErrLoginRateLimited = errors.New("login temporariamente bloqueado")

type loginAttemptState struct {
	Hits         []time.Time
	BlockedUntil time.Time
}

type loginRateLimitError struct {
	until time.Time
}

func (e loginRateLimitError) Error() string {
	remaining := time.Until(e.until).Round(time.Second)
	if remaining < time.Second {
		remaining = time.Second
	}
	return fmt.Sprintf("bah, segura %s antes de tentar logar de novo", remaining.String())
}

func (e loginRateLimitError) Is(target error) bool {
	return target == ErrLoginRateLimited
}

func (e loginRateLimitError) RetryAfter() time.Duration {
	return time.Until(e.until)
}

type Service struct {
	store      *db.Store
	uploadsDir string

	mu            sync.Mutex
	unlocked      map[string]map[int64]time.Time
	typing        map[int64]map[int64]model.PanelTyping
	flood         map[int64][]time.Time
	loginAttempts map[string]loginAttemptState
	roomVer       map[int64]int64
	version       int64
}

func NewService(store *db.Store, uploadsDir string) *Service {
	return &Service{
		store:         store,
		uploadsDir:    uploadsDir,
		unlocked:      map[string]map[int64]time.Time{},
		typing:        map[int64]map[int64]model.PanelTyping{},
		flood:         map[int64][]time.Time{},
		loginAttempts: map[string]loginAttemptState{},
		roomVer:       map[int64]int64{},
		version:       1,
	}
}

func (s *Service) UploadsDir() string {
	return s.uploadsDir
}

func (s *Service) WriteBackupArchive(w io.Writer, source string) (string, error) {
	if s == nil || s.store == nil {
		return "", errors.New("panel service indisponivel")
	}
	if w == nil {
		return "", errors.New("writer de backup indisponivel")
	}

	now := time.Now().UTC()
	tempDir, err := os.MkdirTemp("", "painel-dief-export-*")
	if err != nil {
		return "", fmt.Errorf("create export temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	snapshotPath := filepath.Join(tempDir, "universald.snapshot.db")
	if err := s.store.SnapshotTo(snapshotPath); err != nil {
		return "", err
	}

	snapshotInfo, err := os.Stat(snapshotPath)
	if err != nil {
		return "", fmt.Errorf("stat db snapshot: %w", err)
	}

	uploadFiles, uploadBytes, err := dirUsage(s.uploadsDir)
	if err != nil {
		return "", err
	}

	manifest := map[string]any{
		"service":    "painel-dief",
		"version":    s.Version(),
		"source":     strings.TrimSpace(source),
		"createdAt":  now.Format(time.RFC3339Nano),
		"dbSnapshot": map[string]any{"name": "universald.snapshot.db", "bytes": snapshotInfo.Size()},
		"uploads":    map[string]any{"dir": "panel_uploads", "files": uploadFiles, "bytes": uploadBytes},
	}

	fileName := "painel-dief-export-" + now.Format("20060102-150405") + ".zip"
	archive := zip.NewWriter(w)
	if err := addZipJSON(archive, "backup-manifest.json", manifest, now); err != nil {
		_ = archive.Close()
		return "", err
	}
	if err := addZipFile(archive, "universald.snapshot.db", snapshotPath, now); err != nil {
		_ = archive.Close()
		return "", err
	}
	if err := addZipDir(archive, s.uploadsDir, "panel_uploads", now); err != nil {
		_ = archive.Close()
		return "", err
	}
	if err := archive.Close(); err != nil {
		return "", fmt.Errorf("close backup archive: %w", err)
	}
	return fileName, nil
}

func (s *Service) UploadLimitForRole(role string) int64 {
	return uploadLimitForRole(role)
}

func (s *Service) OpsSummary() (model.PanelOpsSummary, error) {
	summary, err := s.store.PanelOpsSummary(time.Now().UTC(), presenceWindow)
	if err != nil {
		return model.PanelOpsSummary{}, err
	}
	files, bytes, walkErr := dirUsage(s.uploadsDir)
	if walkErr != nil {
		return model.PanelOpsSummary{}, walkErr
	}
	summary.UploadFiles = files
	summary.UploadBytes = bytes
	summary.Version = s.Version()
	return summary, nil
}

func (s *Service) StartMaintenance(ctx context.Context, interval time.Duration) {
	if s == nil {
		return
	}
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.RunMaintenance(time.Now().UTC()); err != nil {
					fmt.Printf("painel maintenance: %v\n", err)
				}
			}
		}
	}()
}

func (s *Service) RunMaintenance(now time.Time) error {
	if s == nil || s.store == nil {
		return errors.New("panel service indisponivel")
	}
	if err := s.store.DeleteExpiredPanelSessions(now); err != nil {
		return err
	}
	if err := s.store.DeleteStalePanelPresence(now.Add(-2 * presenceWindow)); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for sessionID, rooms := range s.unlocked {
		kept := map[int64]time.Time{}
		for roomID, until := range rooms {
			if until.After(now) {
				kept[roomID] = until
			}
		}
		if len(kept) == 0 {
			delete(s.unlocked, sessionID)
			continue
		}
		s.unlocked[sessionID] = kept
	}

	for roomID, members := range s.typing {
		for userID, item := range members {
			if !item.ExpiresAt.After(now) {
				delete(members, userID)
			}
		}
		if len(members) == 0 {
			delete(s.typing, roomID)
		}
	}

	for userID, hits := range s.flood {
		kept := hits[:0]
		for _, at := range hits {
			if now.Sub(at) <= floodWindow {
				kept = append(kept, at)
			}
		}
		if len(kept) == 0 {
			delete(s.flood, userID)
			continue
		}
		s.flood[userID] = kept
	}

	for key, state := range s.loginAttempts {
		kept := state.Hits[:0]
		for _, at := range state.Hits {
			if now.Sub(at) <= loginWindow {
				kept = append(kept, at)
			}
		}
		state.Hits = kept
		if state.BlockedUntil.Before(now) {
			state.BlockedUntil = time.Time{}
		}
		if len(state.Hits) == 0 && state.BlockedUntil.IsZero() {
			delete(s.loginAttempts, key)
			continue
		}
		s.loginAttempts[key] = state
	}

	return nil
}

func (s *Service) EnsureBootstrapped() error {
	if s == nil || s.store == nil {
		return errors.New("panel service indisponivel")
	}

	owner, err := s.ensureOwner()
	if err != nil {
		return err
	}
	aiUser, err := s.ensureAIUser(owner.ID)
	if err != nil {
		return err
	}
	if err := s.ensureRooms(); err != nil {
		return err
	}
	if err := os.MkdirAll(s.uploadsDir, 0o755); err != nil {
		return fmt.Errorf("create uploads dir: %w", err)
	}
	return s.ensureWelcomeMessages(owner, aiUser)
}

func dirUsage(root string) (int, int64, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return 0, 0, nil
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("stat uploads dir: %w", err)
	}
	files := 0
	var bytes int64
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() {
			return nil
		}
		files++
		bytes += info.Size()
		return nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("walk uploads dir: %w", err)
	}
	return files, bytes, nil
}

func addZipJSON(archive *zip.Writer, name string, payload any, now time.Time) error {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal backup manifest: %w", err)
	}
	header := &zip.FileHeader{
		Name:     filepath.ToSlash(name),
		Method:   zip.Deflate,
		Modified: now,
	}
	entry, err := archive.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("create manifest entry: %w", err)
	}
	if _, err := entry.Write(raw); err != nil {
		return fmt.Errorf("write manifest entry: %w", err)
	}
	return nil
}

func addZipFile(archive *zip.Writer, name, path string, now time.Time) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open export file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat export file: %w", err)
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("build zip header: %w", err)
	}
	header.Name = filepath.ToSlash(name)
	header.Method = zip.Deflate
	header.Modified = now
	entry, err := archive.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("create zip entry: %w", err)
	}
	if _, err := io.Copy(entry, file); err != nil {
		return fmt.Errorf("copy zip entry: %w", err)
	}
	return nil
}

func addZipDir(archive *zip.Writer, root, prefix string, now time.Time) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat export dir: %w", err)
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		return addZipFile(archive, filepath.ToSlash(filepath.Join(prefix, relative)), path, now)
	})
}

func (s *Service) Login(login, password string) (model.PanelUser, model.PanelSession, error) {
	key := s.loginThrottleKey(login)
	if err := s.guardLoginThrottle(key); err != nil {
		s.logGuestAction("login_throttle", strings.TrimSpace(login), 0, "", "tentou forcar o login rapido demais")
		return model.PanelUser{}, model.PanelSession{}, err
	}
	user, err := s.store.GetPanelUserByLogin(login)
	if err != nil {
		if throttleErr := s.registerLoginFailure(key); throttleErr != nil {
			s.logGuestAction("login_throttle", strings.TrimSpace(login), 0, "", "usuario inexistente entrou em cooldown de login")
			return model.PanelUser{}, model.PanelSession{}, throttleErr
		}
		s.logGuestAction("login_fail", strings.TrimSpace(login), 0, "", "tentativa com usuario ou email inexistente")
		return model.PanelUser{}, model.PanelSession{}, errors.New("usuario ou email nao encontrado")
	}
	if !verifyPassword(user.PasswordHash, password) {
		if throttleErr := s.registerLoginFailure(key); throttleErr != nil {
			s.logAction(user, "login_throttle", nil, "entrou em cooldown por excesso de tentativa")
			return model.PanelUser{}, model.PanelSession{}, throttleErr
		}
		s.logAction(user, "login_fail", nil, "errou a senha no login")
		return model.PanelUser{}, model.PanelSession{}, errors.New("sai daqui, crianca. essa senha passou longe")
	}
	s.clearLoginThrottle(key, user)

	now := time.Now().UTC()
	sessionID, err := randomToken(32)
	if err != nil {
		return model.PanelUser{}, model.PanelSession{}, fmt.Errorf("session token: %w", err)
	}
	session := model.PanelSession{
		ID:        sessionID,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionLifetime),
	}
	if err := s.store.DeleteExpiredPanelSessions(now); err != nil {
		return model.PanelUser{}, model.PanelSession{}, err
	}
	if err := s.store.SavePanelSession(session); err != nil {
		return model.PanelUser{}, model.PanelSession{}, err
	}
	if err := s.store.SetPanelUserLastLogin(user.ID, now); err != nil {
		return model.PanelUser{}, model.PanelSession{}, err
	}
	if err := s.store.UpsertPanelPresence(user.ID, 0, "online", now); err != nil {
		return model.PanelUser{}, model.PanelSession{}, err
	}

	s.logAction(user, "login_ok", nil, "entrou no Painel Dief")
	s.bumpVersion()

	user.PasswordHash = ""
	user.LastLoginAt = &now
	return user, session, nil
}

func (s *Service) Authenticate(sessionID string) (model.PanelUser, model.PanelSession, error) {
	session, err := s.store.GetPanelSession(strings.TrimSpace(sessionID))
	if err != nil {
		return model.PanelUser{}, model.PanelSession{}, errors.New("sessao expirada")
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		_ = s.store.DeletePanelSession(session.ID)
		return model.PanelUser{}, model.PanelSession{}, errors.New("sessao expirada")
	}
	user, err := s.store.GetPanelUserByID(session.UserID)
	if err != nil {
		return model.PanelUser{}, model.PanelSession{}, errors.New("usuario da sessao nao encontrado")
	}
	user.PasswordHash = ""
	return user, session, nil
}

func (s *Service) Logout(sessionID string) error {
	session, _ := s.store.GetPanelSession(strings.TrimSpace(sessionID))
	if session.UserID > 0 {
		if user, err := s.store.GetPanelUserByID(session.UserID); err == nil {
			s.logAction(user, "logout", nil, "encerrou a sessao")
			s.clearUserTyping(user.ID)
		}
	}

	s.mu.Lock()
	delete(s.unlocked, sessionID)
	s.mu.Unlock()

	if err := s.store.DeletePanelSession(sessionID); err != nil {
		return err
	}
	s.bumpVersion()
	return nil
}

func (s *Service) Bootstrap(user model.PanelUser, sessionID string) (model.PanelBootstrap, error) {
	rooms, access, err := s.visibleRooms(user, sessionID)
	if err != nil {
		return model.PanelBootstrap{}, err
	}
	online, err := s.store.ListPanelPresence(time.Now().UTC(), presenceWindow)
	if err != nil {
		return model.PanelBootstrap{}, err
	}
	blockedIDs, mutedIDs, blockerIDs, err := s.viewerSocialState(user.ID)
	if err != nil {
		return model.PanelBootstrap{}, err
	}
	online = applyPresenceRelations(online, blockedIDs, mutedIDs, blockerIDs)
	logs, err := s.recentLogsForViewer(user)
	if err != nil {
		return model.PanelBootstrap{}, err
	}
	roomIDs := roomIDsByAccess(rooms, access, true)
	latestMessages, err := s.store.ListLatestPanelMessagesForViewer(roomIDs, user.ID)
	if err != nil {
		return model.PanelBootstrap{}, err
	}
	latestMessages = applyMessageBlockState(latestMessages, blockedIDs, user.ID)
	events, err := s.listEventsForViewer(user, rooms, 8)
	if err != nil {
		return model.PanelBootstrap{}, err
	}
	return model.PanelBootstrap{
		Viewer:         user,
		Rooms:          rooms,
		Online:         online,
		Typing:         s.activeTyping(roomIDs, user.ID),
		RecentLogs:     logs,
		LatestMessages: latestMessages,
		Events:         events,
		RoomVersions:   s.roomVersions(roomIDs),
		BlockedUserIDs: setToIDs(blockedIDs),
		MutedUserIDs:   setToIDs(mutedIDs),
		RoomAccess:     access,
		ServerTime:     time.Now().UTC(),
		Version:        s.Version(),
	}, nil
}

func (s *Service) UpdatePresence(user model.PanelUser, roomID int64, status string) error {
	if strings.TrimSpace(status) == "" {
		status = "online"
	}
	if roomID < 0 {
		roomID = 0
	}
	if err := s.store.UpsertPanelPresence(user.ID, roomID, sanitizePresence(status), time.Now().UTC()); err != nil {
		return err
	}
	s.bumpVersion()
	return nil
}

func (s *Service) ListMessages(user model.PanelUser, sessionID string, roomID int64, limit int) ([]model.PanelMessage, model.PanelRoom, error) {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return nil, model.PanelRoom{}, err
	}
	if allowed, _, err := s.checkRoomAccess(user, sessionID, room); err != nil {
		return nil, model.PanelRoom{}, err
	} else if !allowed {
		return nil, model.PanelRoom{}, errors.New("acesso negado nessa sala")
	}
	items, err := s.store.ListPanelMessagesForViewer(roomID, limit, user.ID)
	if err != nil {
		return nil, model.PanelRoom{}, err
	}
	blockedIDs, _, _, err := s.viewerSocialState(user.ID)
	if err != nil {
		return nil, model.PanelRoom{}, err
	}
	items = applyMessageBlockState(items, blockedIDs, user.ID)
	return items, room, nil
}

func (s *Service) PostMessage(user model.PanelUser, sessionID string, roomID int64, body, kind string, attachment *model.PanelAttachment, replyToID int64) (model.PanelMessage, error) {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return model.PanelMessage{}, err
	}
	if allowed, _, err := s.checkRoomAccess(user, sessionID, room); err != nil {
		return model.PanelMessage{}, err
	} else if !allowed {
		return model.PanelMessage{}, errors.New("acesso negado nessa sala")
	}
	if !isPrivileged(user.Role) {
		if err := s.guardFlood(user.ID); err != nil {
			return model.PanelMessage{}, err
		}
	}
	body = strings.TrimSpace(body)
	kind = sanitizeMessageKind(kind, attachment)

	if replyToID > 0 {
		replyTo, err := s.store.GetPanelMessageByID(replyToID)
		if err != nil {
			return model.PanelMessage{}, errors.New("a mensagem respondida sumiu do radar")
		}
		if replyTo.RoomID != roomID {
			return model.PanelMessage{}, errors.New("nao da pra responder mensagem de outra sala")
		}
	}
	if room.Scope == "dm" {
		if err := s.guardDirectMessageRelation(user.ID, room.ID); err != nil {
			return model.PanelMessage{}, err
		}
	}

	msg, err := s.store.CreatePanelMessage(model.PanelMessage{
		RoomID:     roomID,
		AuthorID:   user.ID,
		AuthorName: user.DisplayName,
		AuthorRole: user.Role,
		Body:       body,
		Kind:       kind,
		Attachment: attachment,
		ReplyToID:  replyToID,
	})
	if err != nil {
		return model.PanelMessage{}, err
	}

	s.clearTypingForRoom(roomID, user.ID)
	s.logAction(user, "message_post", &room, describeMessageLog(msg))
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return msg, nil
}

func (s *Service) EditMessage(user model.PanelUser, sessionID string, roomID, messageID int64, body string) (model.PanelMessage, error) {
	room, msg, err := s.loadMessageContext(user, sessionID, roomID, messageID)
	if err != nil {
		return model.PanelMessage{}, err
	}
	if !canManagePanelMessage(user, msg) {
		return model.PanelMessage{}, errors.New("so autor, admin ou owner podem editar essa mensagem")
	}
	body = strings.TrimSpace(body)
	if body == "" && msg.Attachment == nil {
		return model.PanelMessage{}, errors.New("nao da pra salvar mensagem vazia")
	}
	if err := s.store.UpdatePanelMessageBody(msg.ID, body); err != nil {
		return model.PanelMessage{}, err
	}
	updated, err := s.store.GetPanelMessageForViewer(msg.ID, user.ID)
	if err != nil {
		return model.PanelMessage{}, err
	}
	s.logAction(user, "message_edit", &room, fmt.Sprintf("editou a mensagem %d", msg.ID))
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return updated, nil
}

func (s *Service) DeleteMessage(user model.PanelUser, sessionID string, roomID, messageID int64) error {
	room, msg, err := s.loadMessageContext(user, sessionID, roomID, messageID)
	if err != nil {
		return err
	}
	if !canManagePanelMessage(user, msg) {
		return errors.New("so autor, admin ou owner podem apagar essa mensagem")
	}
	if err := s.store.DeletePanelMessage(msg.ID); err != nil {
		return err
	}
	s.logAction(user, "message_delete", &room, fmt.Sprintf("apagou a mensagem %d", msg.ID))
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return nil
}

func (s *Service) TogglePin(user model.PanelUser, sessionID string, roomID, messageID int64) (model.PanelMessage, bool, error) {
	room, msg, err := s.loadMessageContext(user, sessionID, roomID, messageID)
	if err != nil {
		return model.PanelMessage{}, false, err
	}
	if !canPinPanelMessage(user, msg) {
		return model.PanelMessage{}, false, errors.New("so autor, admin ou owner podem fixar essa mensagem")
	}
	pinned, err := s.store.TogglePanelMessagePin(msg.ID, room.ID, user.ID)
	if err != nil {
		return model.PanelMessage{}, false, err
	}
	updated, err := s.store.GetPanelMessageForViewer(msg.ID, user.ID)
	if err != nil {
		return model.PanelMessage{}, false, err
	}
	action := "message_unpin"
	if pinned {
		action = "message_pin"
	}
	s.logAction(user, action, &room, fmt.Sprintf("mexeu no destaque da mensagem %d", msg.ID))
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return updated, pinned, nil
}

func (s *Service) ToggleFavorite(user model.PanelUser, sessionID string, roomID, messageID int64) (model.PanelMessage, bool, error) {
	room, msg, err := s.loadMessageContext(user, sessionID, roomID, messageID)
	if err != nil {
		return model.PanelMessage{}, false, err
	}
	favorited, err := s.store.TogglePanelMessageFavorite(msg.ID, user.ID)
	if err != nil {
		return model.PanelMessage{}, false, err
	}
	updated, err := s.store.GetPanelMessageForViewer(msg.ID, user.ID)
	if err != nil {
		return model.PanelMessage{}, false, err
	}
	action := "message_unfavorite"
	if favorited {
		action = "message_favorite"
	}
	s.logAction(user, action, &room, fmt.Sprintf("mexeu no favorito da mensagem %d", msg.ID))
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return updated, favorited, nil
}

func (s *Service) ListPinnedMessages(user model.PanelUser, sessionID string, roomID int64, limit int) ([]model.PanelMessage, error) {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return nil, err
	}
	if allowed, _, err := s.checkRoomAccess(user, sessionID, room); err != nil {
		return nil, err
	} else if !allowed {
		return nil, errors.New("acesso negado nessa sala")
	}
	items, err := s.store.ListPinnedPanelMessagesForViewer(room.ID, limit, user.ID)
	if err != nil {
		return nil, err
	}
	blockedIDs, _, _, err := s.viewerSocialState(user.ID)
	if err != nil {
		return nil, err
	}
	return applyMessageBlockState(items, blockedIDs, user.ID), nil
}

func (s *Service) UnlockRoom(user model.PanelUser, sessionID string, roomID int64, password string) error {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return err
	}
	if isPrivileged(user.Role) {
		s.markUnlocked(sessionID, roomID)
		s.logAction(user, "room_unlock", &room, "abriu a sala com privilegio elevado")
		return nil
	}
	if !room.PasswordProtected {
		return errors.New("essa sala nao precisa de senha")
	}
	if !verifyPassword(room.PasswordHash, password) {
		s.logAction(user, "room_unlock_fail", &room, "errou a senha da sala")
		return errors.New("senha errada. sai daqui, crianca")
	}
	s.markUnlocked(sessionID, roomID)
	s.logAction(user, "room_unlock", &room, "destravou a sala com a senha correta")
	s.bumpVersion()
	return nil
}

func (s *Service) CreateUser(actor model.PanelUser, username, email, password, displayName, role string) (model.PanelUser, error) {
	if strings.ToLower(actor.Role) != "owner" {
		return model.PanelUser{}, errors.New("so o dono pode cadastrar gente nova")
	}
	role = normalizeRole(role)
	if role == "owner" {
		return model.PanelUser{}, errors.New("o cargo owner fica reservado")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return model.PanelUser{}, err
	}
	user, err := s.store.CreatePanelUser(model.PanelUser{
		Username:     strings.TrimSpace(strings.ToLower(username)),
		Email:        strings.TrimSpace(strings.ToLower(email)),
		DisplayName:  strings.TrimSpace(displayName),
		Role:         role,
		Theme:        "matrix",
		AccentColor:  defaultAccentByRole(role),
		Status:       "online",
		PasswordHash: hash,
		CreatedBy:    actor.ID,
	})
	if err != nil {
		return model.PanelUser{}, err
	}
	user.PasswordHash = ""
	s.logAction(actor, "user_create", nil, fmt.Sprintf("cadastrou %s como %s", user.DisplayName, user.Role))
	s.bumpVersion()
	return user, nil
}

func (s *Service) OpenDirectRoom(viewer model.PanelUser, targetUserID int64) (model.PanelRoom, error) {
	if targetUserID <= 0 {
		return model.PanelRoom{}, errors.New("usuario alvo invalido")
	}
	if viewer.ID == targetUserID {
		return model.PanelRoom{}, errors.New("bah, falar contigo mesmo ja e demais")
	}

	target, err := s.store.GetPanelUserByID(targetUserID)
	if err != nil {
		return model.PanelRoom{}, errors.New("o usuario alvo nao apareceu no radar")
	}
	blocked, err := s.store.IsPanelUserBlocked(viewer.ID, target.ID)
	if err != nil {
		return model.PanelRoom{}, err
	}
	blockedByTarget, err := s.store.IsPanelUserBlocked(target.ID, viewer.ID)
	if err != nil {
		return model.PanelRoom{}, err
	}
	if blocked || blockedByTarget {
		return model.PanelRoom{}, errors.New("essa DM ficou bloqueada por relacao entre usuarios")
	}

	leftUser, rightUser := viewer, target
	if target.ID < viewer.ID {
		leftUser, rightUser = target, viewer
	}

	room, err := s.store.UpsertPanelRoom(model.PanelRoom{
		Slug:        directRoomSlug(leftUser.ID, rightUser.ID),
		Name:        directRoomName(leftUser.DisplayName, rightUser.DisplayName),
		Description: "Conversa direta privada e persistente.",
		Icon:        "DM",
		Category:    "dm",
		Scope:       "dm",
		SortOrder:   90,
	})
	if err != nil {
		return model.PanelRoom{}, err
	}
	if err := s.store.AddPanelRoomMembers(room.ID, viewer.ID, target.ID); err != nil {
		return model.PanelRoom{}, err
	}
	room.PeerUserID = target.ID
	s.logAction(viewer, "dm_open", &room, "abriu ou retomou DM com "+target.DisplayName)
	s.bumpVersion()
	return room, nil
}

func (s *Service) ToggleBlock(viewer model.PanelUser, targetUserID int64) (bool, error) {
	if targetUserID <= 0 || targetUserID == viewer.ID {
		return false, errors.New("nao da pra bloquear esse usuario")
	}
	target, err := s.store.GetPanelUserByID(targetUserID)
	if err != nil {
		return false, errors.New("usuario alvo nao encontrado")
	}
	if strings.ToLower(target.Role) == "owner" {
		return false, errors.New("o dono nao entra nessa brincadeira de bloqueio")
	}
	blocked, err := s.store.TogglePanelUserBlock(viewer.ID, targetUserID)
	if err != nil {
		return false, err
	}
	action := "user_unblock"
	detail := "removeu o bloqueio de " + target.DisplayName
	if blocked {
		action = "user_block"
		detail = "bloqueou " + target.DisplayName
	}
	s.logAction(viewer, action, nil, detail)
	s.bumpVersion()
	return blocked, nil
}

func (s *Service) ToggleMute(viewer model.PanelUser, targetUserID int64) (bool, error) {
	if targetUserID <= 0 || targetUserID == viewer.ID {
		return false, errors.New("nao da pra silenciar esse usuario")
	}
	target, err := s.store.GetPanelUserByID(targetUserID)
	if err != nil {
		return false, errors.New("usuario alvo nao encontrado")
	}
	muted, err := s.store.TogglePanelUserMute(viewer.ID, targetUserID)
	if err != nil {
		return false, err
	}
	action := "user_unmute"
	detail := "tirou o silencio de " + target.DisplayName
	if muted {
		action = "user_mute"
		detail = "silenciou " + target.DisplayName
	}
	s.logAction(viewer, action, nil, detail)
	s.bumpVersion()
	return muted, nil
}

func (s *Service) GetSocialProfile(viewer model.PanelUser, targetUserID int64) (model.PanelSocialProfile, error) {
	if targetUserID <= 0 {
		return model.PanelSocialProfile{}, errors.New("usuario alvo invalido")
	}
	presence, err := s.store.GetPanelPresenceByUserID(targetUserID, time.Now().UTC(), presenceWindow)
	if err != nil {
		return model.PanelSocialProfile{}, err
	}
	blockedIDs, mutedIDs, blockerIDs, err := s.viewerSocialState(viewer.ID)
	if err != nil {
		return model.PanelSocialProfile{}, err
	}
	items := applyPresenceRelations([]model.PanelPresence{presence}, blockedIDs, mutedIDs, blockerIDs)
	presence = items[0]
	return model.PanelSocialProfile{
		User:      presence,
		CanDM:     targetUserID != viewer.ID && !presence.BlockedByViewer && !presence.HasBlockedViewer && strings.ToLower(presence.Role) != "ai",
		CanManage: targetUserID == viewer.ID,
	}, nil
}

func (s *Service) UpdateProfile(user model.PanelUser, displayName, bio, theme, accentColor, avatarURL, status, statusText string) (model.PanelUser, error) {
	cleanAvatarURL, err := sanitizeAvatarURL(avatarURL)
	if err != nil {
		return model.PanelUser{}, err
	}
	updated, err := s.store.UpdatePanelUserProfile(model.PanelUser{
		ID:          user.ID,
		DisplayName: sanitizeDisplayName(displayName, user.DisplayName, user.Username),
		Bio:         sanitizeBio(bio),
		Theme:       sanitizeTheme(theme),
		AccentColor: sanitizeAccent(accentColor),
		AvatarURL:   cleanAvatarURL,
		Status:      sanitizePresence(status),
		StatusText:  sanitizeStatusText(statusText),
	})
	if err != nil {
		return model.PanelUser{}, err
	}
	updated.PasswordHash = ""
	s.logAction(updated, "profile_update", nil, "atualizou o proprio perfil")
	s.bumpVersion()
	return updated, nil
}

func (s *Service) ListUsers(viewer model.PanelUser) ([]model.PanelUser, error) {
	if !isPrivileged(viewer.Role) && strings.ToLower(viewer.Role) != "owner" {
		return nil, errors.New("acesso restrito")
	}
	items, err := s.store.ListPanelUsers()
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].PasswordHash = ""
	}
	return items, nil
}

func (s *Service) SaveUpload(user model.PanelUser, filename, contentType string, src io.Reader) (model.PanelAttachment, error) {
	if err := os.MkdirAll(s.uploadsDir, 0o755); err != nil {
		return model.PanelAttachment{}, fmt.Errorf("create uploads dir: %w", err)
	}
	limit := uploadLimitForRole(user.Role)
	raw, err := io.ReadAll(io.LimitReader(src, limit+1))
	if err != nil {
		return model.PanelAttachment{}, fmt.Errorf("read upload: %w", err)
	}
	if len(raw) == 0 {
		return model.PanelAttachment{}, errors.New("arquivo vazio nao sobe pro painel")
	}
	if int64(len(raw)) > limit {
		return model.PanelAttachment{}, fmt.Errorf("arquivo acima do limite de %s pra esse cargo", uploadLimitLabel(limit))
	}
	token, err := randomToken(8)
	if err != nil {
		return model.PanelAttachment{}, err
	}
	cleanName := sanitizeFilename(filename)
	if cleanName == "" {
		cleanName = "arquivo.bin"
	}
	contentType, kind, err := normalizeUpload(cleanName, contentType, raw)
	if err != nil {
		return model.PanelAttachment{}, err
	}
	ext := strings.ToLower(filepath.Ext(cleanName))
	finalName := token + "-" + cleanName
	fullPath := filepath.Join(s.uploadsDir, finalName)
	file, err := os.Create(fullPath)
	if err != nil {
		return model.PanelAttachment{}, fmt.Errorf("create upload file: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(raw); err != nil {
		return model.PanelAttachment{}, fmt.Errorf("save upload: %w", err)
	}
	width := 0
	height := 0
	if kind == "image" {
		if cfg, _, cfgErr := image.DecodeConfig(bytes.NewReader(raw)); cfgErr == nil {
			width = cfg.Width
			height = cfg.Height
		}
	}
	return model.PanelAttachment{
		Name:        cleanName,
		URL:         "/uploads/" + finalName,
		Kind:        kind,
		ContentType: contentType,
		SizeBytes:   int64(len(raw)),
		Extension:   ext,
		Width:       width,
		Height:      height,
	}, nil
}

func (s *Service) RunTerminal(actor model.PanelUser, command string) (model.PanelTerminalResult, error) {
	if !isPrivileged(actor.Role) {
		return model.PanelTerminalResult{}, errors.New("terminal liberado so para admin/owner")
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return model.PanelTerminalResult{}, errors.New("comando vazio")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-lc", command)
	}
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		exitCode = 1
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
		if ctx.Err() == context.DeadlineExceeded {
			output = append(output, []byte("\n[tempo esgotado no terminal local]")...)
			exitCode = 124
		}
	}
	out := strings.TrimSpace(string(output))
	if len(out) > 16000 {
		out = out[:16000] + "\n...[saida truncada]"
	}

	s.logAction(actor, "terminal_run", nil, fmt.Sprintf("rodou comando [%d] %s", exitCode, summarizeText(command, 120)))
	s.bumpVersion()
	return model.PanelTerminalResult{
		Command:  command,
		Output:   out,
		ExitCode: exitCode,
		RanAt:    time.Now().UTC(),
	}, nil
}

func (s *Service) ToggleReaction(user model.PanelUser, sessionID string, roomID, messageID int64, emoji string) (model.PanelMessage, error) {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return model.PanelMessage{}, err
	}
	if allowed, _, err := s.checkRoomAccess(user, sessionID, room); err != nil {
		return model.PanelMessage{}, err
	} else if !allowed {
		return model.PanelMessage{}, errors.New("acesso negado nessa sala")
	}

	msg, err := s.store.GetPanelMessageForViewer(messageID, user.ID)
	if err != nil {
		return model.PanelMessage{}, err
	}
	if msg.RoomID != roomID {
		return model.PanelMessage{}, errors.New("essa mensagem nao mora nessa sala")
	}

	emoji = sanitizeReaction(emoji)
	if emoji == "" {
		return model.PanelMessage{}, errors.New("reacao invalida")
	}
	if err := s.store.TogglePanelReaction(messageID, user.ID, emoji); err != nil {
		return model.PanelMessage{}, err
	}
	updated, err := s.store.GetPanelMessageForViewer(messageID, user.ID)
	if err != nil {
		return model.PanelMessage{}, err
	}
	s.logAction(user, "reaction_toggle", &room, fmt.Sprintf("mexeu na reacao %s da mensagem %d", emoji, messageID))
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return updated, nil
}

func (s *Service) UpdateTyping(user model.PanelUser, sessionID string, roomID int64, active bool) error {
	if roomID <= 0 {
		s.clearUserTyping(user.ID)
		return nil
	}
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return err
	}
	if allowed, _, err := s.checkRoomAccess(user, sessionID, room); err != nil {
		return err
	} else if !allowed {
		return errors.New("acesso negado nessa sala")
	}

	if !active {
		s.clearTypingForRoom(roomID, user.ID)
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.typing[roomID] == nil {
		s.typing[roomID] = map[int64]model.PanelTyping{}
	}
	s.typing[roomID][user.ID] = model.PanelTyping{
		RoomID:      roomID,
		UserID:      user.ID,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		ExpiresAt:   time.Now().UTC().Add(typingTTL),
	}
	return nil
}

func (s *Service) ListLogs(viewer model.PanelUser, limit int) ([]model.PanelLogItem, error) {
	if !isPrivileged(viewer.Role) {
		return nil, errors.New("logs restritos para admin/owner")
	}
	return s.store.ListPanelLogs(limit)
}

func (s *Service) ListEvents(viewer model.PanelUser, sessionID string, limit int) ([]model.PanelEvent, error) {
	rooms, _, err := s.visibleRooms(viewer, sessionID)
	if err != nil {
		return nil, err
	}
	return s.listEventsForViewer(viewer, rooms, limit)
}

func (s *Service) CreateEvent(viewer model.PanelUser, sessionID string, title, description string, roomID int64, startsAt time.Time) (model.PanelEvent, error) {
	if strings.ToLower(strings.TrimSpace(viewer.Role)) == "ai" {
		return model.PanelEvent{}, errors.New("ia nao cria evento no painel")
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if len(title) < 3 {
		return model.PanelEvent{}, errors.New("titulo do evento ficou curto demais")
	}
	if len(description) > 220 {
		return model.PanelEvent{}, errors.New("descricao do evento passou do limite")
	}
	if startsAt.IsZero() {
		return model.PanelEvent{}, errors.New("data do evento invalida")
	}
	if startsAt.Before(time.Now().UTC().Add(-6 * time.Hour)) {
		return model.PanelEvent{}, errors.New("esse horario ficou velho demais pra entrar na agenda")
	}
	var room model.PanelRoom
	if roomID > 0 {
		var err error
		room, err = s.store.GetPanelRoomByID(roomID)
		if err != nil {
			return model.PanelEvent{}, errors.New("sala do evento nao apareceu no radar")
		}
		if allowed, _, err := s.checkRoomAccess(viewer, sessionID, room); err != nil {
			return model.PanelEvent{}, err
		} else if !allowed {
			return model.PanelEvent{}, errors.New("tu nao pode atrelar evento a essa sala")
		}
	}
	event, err := s.store.CreatePanelEvent(model.PanelEvent{
		Title:         title,
		Description:   description,
		RoomID:        roomID,
		CreatedBy:     viewer.ID,
		CreatedByName: viewer.DisplayName,
		StartsAt:      startsAt.UTC(),
	})
	if err != nil {
		return model.PanelEvent{}, err
	}
	if roomID > 0 {
		event.RoomName = room.Name
	}
	s.logAction(viewer, "event_create", nil, "criou evento "+title)
	s.bumpVersion()
	return event, nil
}

func (s *Service) ToggleEventRSVP(viewer model.PanelUser, sessionID string, eventID int64) (model.PanelEvent, bool, error) {
	if eventID <= 0 {
		return model.PanelEvent{}, false, errors.New("evento invalido")
	}
	events, err := s.ListEvents(viewer, sessionID, 32)
	if err != nil {
		return model.PanelEvent{}, false, err
	}
	var found model.PanelEvent
	for _, item := range events {
		if item.ID == eventID {
			found = item
			break
		}
	}
	if found.ID == 0 {
		return model.PanelEvent{}, false, errors.New("evento nao encontrado no teu alcance")
	}
	joined, err := s.store.TogglePanelEventRSVP(eventID, viewer.ID)
	if err != nil {
		return model.PanelEvent{}, false, err
	}
	event, err := s.store.GetPanelEventForViewer(eventID, viewer.ID)
	if err != nil {
		return model.PanelEvent{}, false, err
	}
	action := "event_leave"
	detail := "saiu do evento " + event.Title
	if joined {
		action = "event_join"
		detail = "confirmou presenca em " + event.Title
	}
	s.logAction(viewer, action, nil, detail)
	s.bumpVersion()
	return event, joined, nil
}

func (s *Service) DeleteEvent(viewer model.PanelUser, sessionID string, eventID int64) error {
	if eventID <= 0 {
		return errors.New("evento invalido")
	}
	event, err := s.store.GetPanelEventForViewer(eventID, viewer.ID)
	if err != nil {
		return errors.New("evento nao encontrado")
	}
	var roomRef *model.PanelRoom
	if event.RoomID > 0 {
		room, err := s.store.GetPanelRoomByID(event.RoomID)
		if err != nil {
			return err
		}
		if allowed, _, err := s.checkRoomAccess(viewer, sessionID, room); err != nil {
			return err
		} else if !allowed {
			return errors.New("tu nao pode apagar evento dessa sala")
		}
		roomRef = &room
	}
	if !canManageOwnedRecord(viewer, event.CreatedBy) {
		return errors.New("so criador, admin ou owner podem apagar esse evento")
	}
	if err := s.store.DeletePanelEvent(eventID); err != nil {
		return err
	}
	s.logAction(viewer, "event_delete", roomRef, "apagou evento "+event.Title)
	s.bumpVersion()
	return nil
}

func (s *Service) ListPolls(viewer model.PanelUser, sessionID string, roomID int64, limit int) ([]model.PanelPoll, error) {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return nil, err
	}
	if allowed, _, err := s.checkRoomAccess(viewer, sessionID, room); err != nil {
		return nil, err
	} else if !allowed {
		return nil, errors.New("tu nao pode listar enquete dessa sala")
	}
	return s.store.ListPanelPollsForViewer(roomID, viewer.ID, limit)
}

func (s *Service) CreatePoll(viewer model.PanelUser, sessionID string, roomID int64, question string, options []string) (model.PanelPoll, error) {
	if strings.ToLower(strings.TrimSpace(viewer.Role)) == "ai" {
		return model.PanelPoll{}, errors.New("ia nao abre enquete no painel")
	}
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return model.PanelPoll{}, err
	}
	if allowed, _, err := s.checkRoomAccess(viewer, sessionID, room); err != nil {
		return model.PanelPoll{}, err
	} else if !allowed {
		return model.PanelPoll{}, errors.New("tu nao pode criar enquete nessa sala")
	}
	question = collapseWhitespace(question)
	if len([]rune(question)) < 6 {
		return model.PanelPoll{}, errors.New("pergunta da enquete ficou curta demais")
	}
	if len([]rune(question)) > 140 {
		question = string([]rune(question)[:140])
	}
	cleanOptions := make([]model.PanelPollOption, 0, len(options))
	seen := map[string]bool{}
	for _, option := range options {
		option = collapseWhitespace(option)
		if option == "" {
			continue
		}
		if len([]rune(option)) > 60 {
			option = string([]rune(option)[:60])
		}
		key := strings.ToLower(option)
		if seen[key] {
			continue
		}
		seen[key] = true
		cleanOptions = append(cleanOptions, model.PanelPollOption{Label: option})
	}
	if len(cleanOptions) < 2 {
		return model.PanelPoll{}, errors.New("a enquete precisa de no minimo duas opcoes diferentes")
	}
	if len(cleanOptions) > 5 {
		cleanOptions = cleanOptions[:5]
	}
	poll, err := s.store.CreatePanelPoll(model.PanelPoll{
		RoomID:        roomID,
		Question:      question,
		CreatedBy:     viewer.ID,
		CreatedByName: viewer.DisplayName,
		Options:       cleanOptions,
	})
	if err != nil {
		return model.PanelPoll{}, err
	}
	s.logAction(viewer, "poll_create", &room, "criou enquete "+question)
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return poll, nil
}

func (s *Service) VotePoll(viewer model.PanelUser, sessionID string, roomID, pollID, optionID int64) (model.PanelPoll, bool, error) {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return model.PanelPoll{}, false, err
	}
	if allowed, _, err := s.checkRoomAccess(viewer, sessionID, room); err != nil {
		return model.PanelPoll{}, false, err
	} else if !allowed {
		return model.PanelPoll{}, false, errors.New("tu nao pode votar nessa enquete")
	}
	polls, err := s.store.ListPanelPollsForViewer(roomID, viewer.ID, 24)
	if err != nil {
		return model.PanelPoll{}, false, err
	}
	var found model.PanelPoll
	for _, item := range polls {
		if item.ID == pollID {
			found = item
			break
		}
	}
	if found.ID == 0 {
		return model.PanelPoll{}, false, errors.New("enquete nao encontrada nessa sala")
	}
	var optionLabel string
	for _, option := range found.Options {
		if option.ID == optionID {
			optionLabel = option.Label
			break
		}
	}
	if optionLabel == "" {
		return model.PanelPoll{}, false, errors.New("opcao da enquete nao apareceu no radar")
	}
	voted, err := s.store.TogglePanelPollVote(pollID, optionID, viewer.ID)
	if err != nil {
		return model.PanelPoll{}, false, err
	}
	poll, err := s.store.GetPanelPollForViewer(pollID, viewer.ID)
	if err != nil {
		return model.PanelPoll{}, false, err
	}
	detail := "tirou voto da enquete " + poll.Question
	action := "poll_unvote"
	if voted {
		action = "poll_vote"
		detail = "votou em " + optionLabel + " na enquete " + poll.Question
	}
	s.logAction(viewer, action, &room, detail)
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return poll, voted, nil
}

func (s *Service) DeletePoll(viewer model.PanelUser, sessionID string, roomID, pollID int64) error {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return err
	}
	if allowed, _, err := s.checkRoomAccess(viewer, sessionID, room); err != nil {
		return err
	} else if !allowed {
		return errors.New("tu nao pode apagar enquete dessa sala")
	}
	poll, err := s.store.GetPanelPollForViewer(pollID, viewer.ID)
	if err != nil || poll.RoomID != roomID {
		return errors.New("enquete nao encontrada nessa sala")
	}
	if !canManageOwnedRecord(viewer, poll.CreatedBy) {
		return errors.New("so criador, admin ou owner podem apagar essa enquete")
	}
	if err := s.store.DeletePanelPoll(pollID); err != nil {
		return err
	}
	s.logAction(viewer, "poll_delete", &room, "apagou enquete "+poll.Question)
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return nil
}

func (s *Service) Search(viewer model.PanelUser, sessionID, query string, limit int) (model.PanelSearchResult, error) {
	rooms, access, err := s.visibleRooms(viewer, sessionID)
	if err != nil {
		return model.PanelSearchResult{}, err
	}
	accessible := roomIDsByAccess(rooms, access, false)
	result, err := s.store.SearchPanel(query, accessible, viewer.ID, limit)
	if err != nil {
		return model.PanelSearchResult{}, err
	}
	blockedIDs, _, _, err := s.viewerSocialState(viewer.ID)
	if err != nil {
		return model.PanelSearchResult{}, err
	}
	result.Messages = applyMessageBlockState(result.Messages, blockedIDs, viewer.ID)
	for i := range result.Users {
		result.Users[i].PasswordHash = ""
	}
	return result, nil
}

func (s *Service) AskAI(viewer model.PanelUser, prompt string) string {
	prompt = strings.TrimSpace(prompt)
	lower := strings.ToLower(prompt)

	base := "Bah tche, "
	switch {
	case lower == "":
		return base + "manda a pergunta inteira que eu te explico esse painel sem lero-lero."
	case strings.Contains(lower, "login") || strings.Contains(lower, "senha"):
		return base + "tu entra com usuario ou email cadastrado pelo owner. Se errar a senha, o painel te gasta sem pena."
	case strings.Contains(lower, "admin") || strings.Contains(lower, "owner"):
		return base + "owner cadastra usuarios, admin governa sala oculta, logs, terminal e liberacao de acesso. O owner segue acima do resto."
	case strings.Contains(lower, "vip"):
		return base + "o lounge VIP fica reservado. Quem eh VIP, admin ou owner entra; o resto bate na porta."
	case strings.Contains(lower, "tema") || strings.Contains(lower, "matrix"):
		return base + "cada perfil escolhe tema, cor, avatar e bio. Tem matrix, obsidian, ember, cobalt e neon pra deixar a cabine com mais personalidade."
	case strings.Contains(lower, "fotos") || strings.Contains(lower, "arquivos") || strings.Contains(lower, "upload"):
		return base + "foto, video, audio e arquivo sobem pro storage local e podem ir direto pras salas certas com preview."
	case strings.Contains(lower, "tempo real") || strings.Contains(lower, "online") || strings.Contains(lower, "typing"):
		return base + "presenca, digitando, reacoes e atualizacao do chat rodam em tempo real pelo stream do servidor."
	case strings.Contains(lower, "apps") || strings.Contains(lower, "terminal") || strings.Contains(lower, "codigo"):
		return base + "no Apps Lab tu tem chat tecnico, painel de busca e terminal local. Mas terminal fica preso em admin e owner pra nao virar bagunca."
	case strings.Contains(lower, "evento") || strings.Contains(lower, "agenda"):
		return base + "na visao rapida tu encontra a agenda da base. Da pra criar evento, atrelar sala e confirmar presenca sem sair do painel."
	case strings.Contains(lower, "enquete") || strings.Contains(lower, "poll"):
		return base + "agora cada sala pode abrir enquete rapida pra decidir jogo, horario ou qualquer treta do grupo sem sair do chat."
	case strings.Contains(lower, "busca") || strings.Contains(lower, "search"):
		return base + "tem busca global por sala, usuario e mensagem dentro do que tu pode acessar. Nada de meter o nariz onde nao foi liberado."
	case strings.Contains(lower, "dm") || strings.Contains(lower, "direta") || strings.Contains(lower, "privado"):
		return base + "da pra abrir conversa direta com a galera pela lista de membros ou pela busca. Fica persistente e separada dos canais."
	case strings.Contains(lower, "reacao") || strings.Contains(lower, "emoji") || strings.Contains(lower, "fix") || strings.Contains(lower, "favorit") || strings.Contains(lower, "editar"):
		return base + "agora da pra reagir, editar, apagar, fixar e favoritar mensagem. O chat ficou bem mais facil de organizar sem perder o fio."
	default:
		return base + "o Painel Dief virou uma central gamer com salas por categoria, busca, respostas, reacoes, fixados, favoritos, logs admin, uploads e perfil customizado. Se quiser, eu destrincho qualquer modulo."
	}
}

func (s *Service) PostAIExchange(viewer model.PanelUser, sessionID string, roomID int64, prompt string) (model.PanelMessage, model.PanelMessage, error) {
	userMsg, err := s.PostMessage(viewer, sessionID, roomID, prompt, "text", nil, 0)
	if err != nil {
		return model.PanelMessage{}, model.PanelMessage{}, err
	}
	aiUser, err := s.store.GetPanelUserByLogin("nego.dramias")
	if err != nil {
		return model.PanelMessage{}, model.PanelMessage{}, err
	}
	reply, err := s.store.CreatePanelMessage(model.PanelMessage{
		RoomID:     roomID,
		AuthorID:   aiUser.ID,
		AuthorName: "Nego Dramias",
		AuthorRole: "ai",
		Body:       s.AskAI(viewer, prompt),
		Kind:       "text",
		IsAI:       true,
		ReplyToID:  userMsg.ID,
	})
	if err != nil {
		return model.PanelMessage{}, model.PanelMessage{}, err
	}
	if room, roomErr := s.store.GetPanelRoomByID(roomID); roomErr == nil {
		s.logAction(aiUser, "ai_reply", &room, fmt.Sprintf("respondeu a %s", viewer.DisplayName))
	}
	s.bumpRoomVersion(roomID)
	s.bumpVersion()
	return userMsg, reply, nil
}

func (s *Service) StreamSnapshot(user model.PanelUser, sessionID string) (map[string]any, error) {
	rooms, access, err := s.visibleRooms(user, sessionID)
	if err != nil {
		return nil, err
	}
	online, err := s.store.ListPanelPresence(time.Now().UTC(), presenceWindow)
	if err != nil {
		return nil, err
	}
	blockedIDs, mutedIDs, blockerIDs, err := s.viewerSocialState(user.ID)
	if err != nil {
		return nil, err
	}
	online = applyPresenceRelations(online, blockedIDs, mutedIDs, blockerIDs)
	logs, err := s.recentLogsForViewer(user)
	if err != nil {
		return nil, err
	}
	roomIDs := roomIDsByAccess(rooms, access, true)
	latestMessages, err := s.store.ListLatestPanelMessagesForViewer(roomIDs, user.ID)
	if err != nil {
		return nil, err
	}
	latestMessages = applyMessageBlockState(latestMessages, blockedIDs, user.ID)
	events, err := s.listEventsForViewer(user, rooms, 8)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"version":        s.Version(),
		"serverTime":     time.Now().UTC(),
		"online":         online,
		"rooms":          rooms,
		"typing":         s.activeTyping(roomIDs, user.ID),
		"recentLogs":     logs,
		"latestMessages": latestMessages,
		"events":         events,
		"roomVersions":   s.roomVersions(roomIDs),
		"blockedUserIds": setToIDs(blockedIDs),
		"mutedUserIds":   setToIDs(mutedIDs),
		"roomAccess":     access,
	}, nil
}

func (s *Service) loadMessageContext(user model.PanelUser, sessionID string, roomID, messageID int64) (model.PanelRoom, model.PanelMessage, error) {
	room, err := s.store.GetPanelRoomByID(roomID)
	if err != nil {
		return model.PanelRoom{}, model.PanelMessage{}, err
	}
	if allowed, _, err := s.checkRoomAccess(user, sessionID, room); err != nil {
		return model.PanelRoom{}, model.PanelMessage{}, err
	} else if !allowed {
		return model.PanelRoom{}, model.PanelMessage{}, errors.New("acesso negado nessa sala")
	}
	msg, err := s.store.GetPanelMessageForViewer(messageID, user.ID)
	if err != nil {
		return model.PanelRoom{}, model.PanelMessage{}, err
	}
	if msg.RoomID != room.ID {
		return model.PanelRoom{}, model.PanelMessage{}, errors.New("essa mensagem nao mora nessa sala")
	}
	return room, msg, nil
}

func (s *Service) Version() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.version
}

func (s *Service) roomVersions(roomIDs []int64) map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make(map[string]int64, len(roomIDs))
	for _, roomID := range roomIDs {
		value := s.roomVer[roomID]
		if value <= 0 {
			value = 1
		}
		items[strconv.FormatInt(roomID, 10)] = value
	}
	return items
}

func (s *Service) viewerSocialState(viewerID int64) (map[int64]bool, map[int64]bool, map[int64]bool, error) {
	blockedIDs, err := s.store.ListPanelBlockedIDs(viewerID)
	if err != nil {
		return nil, nil, nil, err
	}
	mutedIDs, err := s.store.ListPanelMutedIDs(viewerID)
	if err != nil {
		return nil, nil, nil, err
	}
	blockerIDs, err := s.store.ListPanelBlockersForUser(viewerID)
	if err != nil {
		return nil, nil, nil, err
	}
	return idsToSet(blockedIDs), idsToSet(mutedIDs), idsToSet(blockerIDs), nil
}

func (s *Service) listEventsForViewer(viewer model.PanelUser, rooms []model.PanelRoom, limit int) ([]model.PanelEvent, error) {
	items, err := s.store.ListUpcomingPanelEventsForViewer(viewer.ID, time.Now().UTC().Add(-4*time.Hour), limit)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return items, nil
	}
	allowedRooms := map[int64]string{}
	for _, room := range rooms {
		allowedRooms[room.ID] = room.Name
	}
	out := make([]model.PanelEvent, 0, len(items))
	for _, item := range items {
		if item.RoomID > 0 {
			name, ok := allowedRooms[item.RoomID]
			if !ok {
				continue
			}
			if strings.TrimSpace(item.RoomName) == "" {
				item.RoomName = name
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *Service) attachRoomPeer(room *model.PanelRoom, viewerID int64) {
	if room == nil || room.Scope != "dm" || viewerID <= 0 {
		return
	}
	memberIDs, err := s.store.ListPanelRoomMemberIDs(room.ID)
	if err != nil {
		return
	}
	for _, memberID := range memberIDs {
		if memberID != viewerID {
			room.PeerUserID = memberID
			return
		}
	}
}

func (s *Service) guardDirectMessageRelation(viewerID, roomID int64) error {
	if viewerID <= 0 || roomID <= 0 {
		return nil
	}
	memberIDs, err := s.store.ListPanelRoomMemberIDs(roomID)
	if err != nil {
		return err
	}
	for _, memberID := range memberIDs {
		if memberID == viewerID {
			continue
		}
		blocked, err := s.store.IsPanelUserBlocked(viewerID, memberID)
		if err != nil {
			return err
		}
		blockedByPeer, err := s.store.IsPanelUserBlocked(memberID, viewerID)
		if err != nil {
			return err
		}
		if blocked || blockedByPeer {
			return errors.New("essa DM nao aceita mensagem enquanto houver bloqueio entre voces")
		}
	}
	return nil
}

func applyPresenceRelations(items []model.PanelPresence, blockedIDs, mutedIDs, blockerIDs map[int64]bool) []model.PanelPresence {
	for i := range items {
		items[i].BlockedByViewer = blockedIDs[items[i].UserID]
		items[i].MutedByViewer = mutedIDs[items[i].UserID]
		items[i].HasBlockedViewer = blockerIDs[items[i].UserID]
	}
	return items
}

func applyMessageBlockState(items []model.PanelMessage, blockedIDs map[int64]bool, viewerID int64) []model.PanelMessage {
	for i := range items {
		if items[i].AuthorID == viewerID {
			continue
		}
		items[i].BlockedByViewer = blockedIDs[items[i].AuthorID]
		if items[i].BlockedByViewer {
			items[i].Reply = nil
			items[i].Reactions = nil
			items[i].Attachment = nil
		}
	}
	return items
}

func (s *Service) visibleRooms(user model.PanelUser, sessionID string) ([]model.PanelRoom, map[string]string, error) {
	rooms, err := s.store.ListPanelRooms()
	if err != nil {
		return nil, nil, err
	}
	out := make([]model.PanelRoom, 0, len(rooms))
	access := map[string]string{}
	for _, room := range rooms {
		allowed, accessType, err := s.checkRoomAccess(user, sessionID, room)
		if err != nil {
			return nil, nil, err
		}
		if room.Scope == "dm" && !allowed {
			continue
		}
		if room.AdminOnly && !allowed {
			continue
		}
		if room.VIPOnly && !allowed && !isPrivileged(user.Role) {
			s.attachRoomPeer(&room, user.ID)
			out = append(out, room)
			access[room.Slug] = "vip"
			continue
		}
		s.attachRoomPeer(&room, user.ID)
		out = append(out, room)
		access[room.Slug] = accessType
	}
	return out, access, nil
}

func (s *Service) checkRoomAccess(user model.PanelUser, sessionID string, room model.PanelRoom) (bool, string, error) {
	role := strings.ToLower(strings.TrimSpace(user.Role))
	if room.Scope == "dm" {
		if isPrivileged(role) {
			return true, "dm", nil
		}
		member, err := s.store.IsPanelRoomMember(room.ID, user.ID)
		if err != nil {
			return false, "hidden", err
		}
		if member {
			return true, "dm", nil
		}
		return false, "hidden", nil
	}
	if room.AdminOnly {
		if isPrivileged(role) {
			return true, "admin", nil
		}
		return false, "hidden", nil
	}
	if room.VIPOnly && role != "vip" && !isPrivileged(role) {
		return false, "vip", nil
	}
	if room.PasswordProtected && !isPrivileged(role) && !s.isUnlocked(sessionID, room.ID) {
		return false, "locked", nil
	}
	return true, "open", nil
}

func (s *Service) markUnlocked(sessionID string, roomID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.unlocked[sessionID] == nil {
		s.unlocked[sessionID] = map[int64]time.Time{}
	}
	s.unlocked[sessionID][roomID] = time.Now().UTC()
}

func (s *Service) isUnlocked(sessionID string, roomID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	rooms := s.unlocked[sessionID]
	if rooms == nil {
		return false
	}
	_, ok := rooms[roomID]
	return ok
}

func (s *Service) loginThrottleKey(login string) string {
	key := strings.ToLower(strings.TrimSpace(login))
	if key == "" {
		return "guest"
	}
	return key
}

func (s *Service) trimLoginHits(now time.Time, hits []time.Time) []time.Time {
	recent := hits[:0]
	for _, at := range hits {
		if now.Sub(at) <= loginWindow {
			recent = append(recent, at)
		}
	}
	return recent
}

func (s *Service) guardLoginThrottle(key string) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.loginAttempts[key]
	state.Hits = s.trimLoginHits(now, state.Hits)
	if state.BlockedUntil.After(now) {
		s.loginAttempts[key] = state
		return loginRateLimitError{until: state.BlockedUntil}
	}
	if state.BlockedUntil.IsZero() && len(state.Hits) == 0 {
		delete(s.loginAttempts, key)
		return nil
	}
	state.BlockedUntil = time.Time{}
	s.loginAttempts[key] = state
	return nil
}

func (s *Service) registerLoginFailure(key string) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.loginAttempts[key]
	state.Hits = s.trimLoginHits(now, state.Hits)
	if state.BlockedUntil.After(now) {
		s.loginAttempts[key] = state
		return loginRateLimitError{until: state.BlockedUntil}
	}
	state.Hits = append(state.Hits, now)
	if len(state.Hits) >= loginBurst {
		state.Hits = nil
		state.BlockedUntil = now.Add(loginCooldown)
		s.loginAttempts[key] = state
		return loginRateLimitError{until: state.BlockedUntil}
	}
	s.loginAttempts[key] = state
	return nil
}

func (s *Service) clearLoginThrottle(login string, user model.PanelUser) {
	keys := []string{
		s.loginThrottleKey(login),
		s.loginThrottleKey(user.Username),
		s.loginThrottleKey(user.Email),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range keys {
		delete(s.loginAttempts, key)
	}
}

func (s *Service) guardFlood(userID int64) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	recent := s.flood[userID][:0]
	for _, at := range s.flood[userID] {
		if now.Sub(at) <= floodWindow {
			recent = append(recent, at)
		}
	}
	if len(recent) >= floodBurst {
		s.flood[userID] = recent
		return errors.New("bah tche, segura o spam e respira uns segundos")
	}
	recent = append(recent, now)
	s.flood[userID] = recent
	return nil
}

func (s *Service) activeTyping(roomIDs []int64, excludeUserID int64) []model.PanelTyping {
	now := time.Now().UTC()
	roomSet := make(map[int64]struct{}, len(roomIDs))
	for _, roomID := range roomIDs {
		roomSet[roomID] = struct{}{}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]model.PanelTyping, 0)
	for roomID, typers := range s.typing {
		if _, ok := roomSet[roomID]; !ok {
			continue
		}
		for userID, item := range typers {
			if now.After(item.ExpiresAt) {
				delete(typers, userID)
				continue
			}
			if userID == excludeUserID {
				continue
			}
			items = append(items, item)
		}
		if len(typers) == 0 {
			delete(s.typing, roomID)
		}
	}
	return items
}

func (s *Service) clearTypingForRoom(roomID, userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.typing[roomID] == nil {
		return
	}
	delete(s.typing[roomID], userID)
	if len(s.typing[roomID]) == 0 {
		delete(s.typing, roomID)
	}
}

func (s *Service) clearUserTyping(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for roomID := range s.typing {
		delete(s.typing[roomID], userID)
		if len(s.typing[roomID]) == 0 {
			delete(s.typing, roomID)
		}
	}
}

func (s *Service) recentLogsForViewer(user model.PanelUser) ([]model.PanelLogItem, error) {
	if !isPrivileged(user.Role) {
		return []model.PanelLogItem{}, nil
	}
	return s.store.ListPanelLogs(recentLogLimit)
}

func (s *Service) bumpVersion() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.version++
}

func (s *Service) bumpRoomVersion(roomID int64) {
	if roomID <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roomVer[roomID]++
	if s.roomVer[roomID] <= 0 {
		s.roomVer[roomID] = 1
	}
}

func (s *Service) ensureOwner() (model.PanelUser, error) {
	if user, err := s.store.GetPanelUserByLogin(ownerUsername()); err == nil {
		return user, nil
	}
	count, err := s.store.CountPanelUsers()
	if err != nil {
		return model.PanelUser{}, err
	}
	if count > 0 {
		users, err := s.store.ListPanelUsers()
		if err != nil {
			return model.PanelUser{}, err
		}
		if len(users) == 0 {
			return model.PanelUser{}, errors.New("nao foi possivel carregar owner")
		}
		return users[0], nil
	}
	hash, err := hashPassword(ownerPassword())
	if err != nil {
		return model.PanelUser{}, err
	}
	return s.store.CreatePanelUser(model.PanelUser{
		Username:     ownerUsername(),
		Email:        ownerEmail(),
		DisplayName:  "Dono do Painel",
		Role:         "owner",
		Theme:        "matrix",
		AccentColor:  "#7bff00",
		Bio:          "Owner do Painel Dief.",
		Status:       "online",
		PasswordHash: hash,
	})
}

func (s *Service) ensureAIUser(ownerID int64) (model.PanelUser, error) {
	user, err := s.store.GetPanelUserByLogin("nego.dramias")
	if err == nil {
		return user, nil
	}
	hash, err := hashPassword("ia-local-never-login")
	if err != nil {
		return model.PanelUser{}, err
	}
	return s.store.CreatePanelUser(model.PanelUser{
		Username:     "nego.dramias",
		Email:        "ia@paineldief.local",
		DisplayName:  "Nego Dramias",
		Role:         "ai",
		Theme:        "matrix",
		AccentColor:  "#00f7ff",
		Bio:          "Assistente local do Painel Dief.",
		Status:       "away",
		PasswordHash: hash,
		CreatedBy:    ownerID,
	})
}

func (s *Service) ensureRooms() error {
	privateHash, err := hashPassword(defaultPrivatePass)
	if err != nil {
		return err
	}
	vipHash, err := hashPassword(defaultVIPPass)
	if err != nil {
		return err
	}
	rooms := []model.PanelRoom{
		{Slug: "chat-geral", Name: "Chat Geral", Description: "Canal principal da tropa.", Icon: "GL", Category: "chat", Scope: "public", SortOrder: 10},
		{Slug: "fotos", Name: "Fotos", Description: "Prints, memes, setup e drops visuais.", Icon: "PX", Category: "media", Scope: "public", SortOrder: 20},
		{Slug: "arquivos", Name: "Arquivos", Description: "Arquivos, packs, audios e docs.", Icon: "AR", Category: "media", Scope: "public", SortOrder: 30},
		{Slug: "chat-priv", Name: "Chat Priv", Description: "Sala com senha para papo reservado.", Icon: "PR", Category: "chat", Scope: "public", SortOrder: 40, PasswordHash: privateHash},
		{Slug: "nego-dramias-ia", Name: "Nego Dramias IA", Description: "Assistente do painel, sarcatico e gaucho.", Icon: "AI", Category: "ia", Scope: "public", SortOrder: 50},
		{Slug: "apps-lab", Name: "Apps Lab", Description: "Sala tecnica para codigo, snippets e terminal.", Icon: "DEV", Category: "dev", Scope: "public", SortOrder: 60},
		{Slug: "lounge-vip", Name: "Lounge VIP", Description: "Canal reservado para VIPs e admins.", Icon: "VIP", Category: "vip", Scope: "public", SortOrder: 70, VIPOnly: true, PasswordHash: vipHash},
		{Slug: "cofre-admin", Name: "Cofre Admin", Description: "Sala oculta para moderacao e controle total.", Icon: "ADM", Category: "admin", Scope: "public", SortOrder: 80, AdminOnly: true},
	}
	for _, room := range rooms {
		if _, err := s.store.UpsertPanelRoom(room); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ensureWelcomeMessages(owner, aiUser model.PanelUser) error {
	rooms, err := s.store.ListPanelRooms()
	if err != nil {
		return err
	}
	for _, room := range rooms {
		items, err := s.store.ListPanelMessages(room.ID, 1)
		if err != nil {
			return err
		}
		if len(items) > 0 {
			continue
		}
		msg := model.PanelMessage{
			RoomID: room.ID,
			Kind:   "text",
		}
		switch room.Slug {
		case "nego-dramias-ia":
			msg.AuthorID = aiUser.ID
			msg.AuthorName = "Nego Dramias"
			msg.AuthorRole = "ai"
			msg.Body = "Bah tche, me chama que eu te explico o painel, as salas, as permissoes e ate o terminal sem rodeio."
			msg.IsAI = true
		case "cofre-admin":
			msg.AuthorID = owner.ID
			msg.AuthorName = owner.DisplayName
			msg.AuthorRole = owner.Role
			msg.Body = "Area oculta liberada. Aqui ficam logs, moderacao e as conversas que nao saem pra rua."
		default:
			msg.AuthorID = owner.ID
			msg.AuthorName = owner.DisplayName
			msg.AuthorRole = owner.Role
			msg.Body = welcomeText(room.Slug)
		}
		if _, err := s.store.CreatePanelMessage(msg); err != nil {
			return err
		}
	}
	s.bumpVersion()
	return nil
}

func (s *Service) logAction(actor model.PanelUser, action string, room *model.PanelRoom, detail string) {
	item := model.PanelLogItem{
		Action:    action,
		ActorID:   actor.ID,
		ActorName: actor.DisplayName,
		Detail:    detail,
	}
	if item.ActorName == "" {
		item.ActorName = actor.Username
	}
	if room != nil {
		item.RoomID = room.ID
		item.RoomSlug = room.Slug
	}
	_, _ = s.store.CreatePanelLog(item)
}

func (s *Service) logGuestAction(action, actorName string, roomID int64, roomSlug, detail string) {
	_, _ = s.store.CreatePanelLog(model.PanelLogItem{
		Action:    action,
		ActorID:   0,
		ActorName: strings.TrimSpace(actorName),
		RoomID:    roomID,
		RoomSlug:  roomSlug,
		Detail:    detail,
	})
}

func hashPassword(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 4 {
		return "", errors.New("senha curta demais")
	}
	salt, err := randomToken(18)
	if err != nil {
		return "", err
	}
	encoded := derivePassword(raw, salt, passwordIterations)
	return "sha256$" + strconv.Itoa(passwordIterations) + "$" + salt + "$" + encoded, nil
}

func verifyPassword(stored, raw string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter < 1000 {
		return false
	}
	expected := derivePassword(strings.TrimSpace(raw), parts[2], iter)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(parts[3])) == 1
}

func derivePassword(raw, salt string, iterations int) string {
	buf := []byte(raw + "|" + salt)
	for i := 0; i < iterations; i++ {
		sum := sha256.Sum256(buf)
		buf = append(sum[:], salt...)
	}
	sum := sha256.Sum256(buf)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func sanitizeTheme(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "matrix", "obsidian", "ember", "cobalt", "neon":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "matrix"
	}
}

func sanitizePresence(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "online", "busy", "away", "offline":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "online"
	}
}

func sanitizeAccent(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) != 7 || !strings.HasPrefix(raw, "#") {
		return "#7bff00"
	}
	return raw
}

func sanitizeDisplayName(raw string, fallbacks ...string) string {
	raw = collapseWhitespace(raw)
	if raw == "" {
		for _, fallback := range fallbacks {
			fallback = collapseWhitespace(fallback)
			if fallback != "" {
				raw = fallback
				break
			}
		}
	}
	if raw == "" {
		raw = "Membro"
	}
	if len([]rune(raw)) > 48 {
		raw = string([]rune(raw)[:48])
	}
	return raw
}

func sanitizeBio(raw string) string {
	raw = collapseWhitespace(raw)
	if len([]rune(raw)) > 320 {
		raw = string([]rune(raw)[:320])
	}
	return raw
}

func sanitizeStatusText(raw string) string {
	raw = collapseWhitespace(raw)
	if len([]rune(raw)) > 80 {
		raw = string([]rune(raw)[:80])
	}
	return raw
}

func sanitizeAvatarURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if len(raw) > 512 {
		return "", errors.New("avatar grande demais. usa um link menor ou upload local")
	}
	if strings.HasPrefix(raw, "/uploads/") {
		clean := filepath.ToSlash(filepath.Clean(raw))
		if !strings.HasPrefix(clean, "/uploads/") || clean == "/uploads" || strings.Contains(clean, "..") {
			return "", errors.New("avatar invalido. usa upload local do painel ou https")
		}
		if !isAllowedAvatarExt(filepath.Ext(clean)) {
			return "", errors.New("avatar precisa ser imagem valida do painel")
		}
		return clean, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil {
		return "", errors.New("link de avatar invalido")
	}
	if strings.ToLower(parsed.Scheme) != "https" || strings.TrimSpace(parsed.Host) == "" || parsed.User != nil {
		return "", errors.New("avatar externo precisa usar https valido")
	}
	return parsed.String(), nil
}

func sanitizeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	replacer := strings.NewReplacer("..", "", "/", "-", "\\", "-", ":", "-", "*", "-", "\"", "", "<", "", ">", "", "|", "-")
	return replacer.Replace(name)
}

func isAllowedAvatarExt(ext string) bool {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func collapseWhitespace(raw string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
}

func uploadLimitForRole(role string) int64 {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "admin":
		return adminUploadLimit
	case "vip":
		return vipUploadLimit
	default:
		return defaultUploadLimit
	}
}

func uploadLimitLabel(size int64) string {
	if size%(1<<20) == 0 {
		return fmt.Sprintf("%dMB", size>>20)
	}
	return fmt.Sprintf("%d bytes", size)
}

func normalizeUpload(filename, declaredType string, raw []byte) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	sniffed := http.DetectContentType(raw)
	declared := strings.ToLower(strings.TrimSpace(strings.Split(declaredType, ";")[0]))
	kind := inferAttachmentKind(sniffed, filename)
	if kind == "file" {
		kind = inferAttachmentKind(declared, filename)
	}
	if !isAllowedUpload(ext, declared, sniffed) {
		return "", "", errors.New("formato nao permitido. usa imagem, video, audio ou arquivo comum suportado")
	}
	contentType := sniffed
	if strings.TrimSpace(contentType) == "" || contentType == "application/octet-stream" {
		contentType = declared
	}
	if strings.TrimSpace(contentType) == "" || contentType == "application/octet-stream" {
		contentType = contentTypeByExtension(ext)
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	return contentType, kind, nil
}

func isAllowedUpload(ext, declaredType, sniffedType string) bool {
	if isBlockedUploadExt(ext) || isBlockedUploadType(declaredType) || isBlockedUploadType(sniffedType) {
		return false
	}
	if strings.TrimSpace(ext) != "" {
		return isAllowedUploadExt(ext)
	}
	return isAllowedUploadType(declaredType) || isAllowedUploadType(sniffedType)
}

func isBlockedUploadExt(ext string) bool {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".html", ".htm", ".css", ".scss", ".svg", ".xml", ".xhtml", ".mjs":
		return true
	default:
		return false
	}
}

func isAllowedUploadExt(ext string) bool {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp",
		".mp4", ".webm", ".mov", ".m4v",
		".mp3", ".wav", ".ogg", ".m4a", ".aac",
		".pdf", ".zip", ".rar", ".7z",
		".txt", ".md", ".json", ".csv", ".log",
		".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".js", ".ts", ".tsx", ".jsx", ".go", ".py", ".sql", ".yaml", ".yml", ".toml", ".ini":
		return true
	default:
		return false
	}
}

func isBlockedUploadType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	switch value {
	case "text/html", "application/xhtml+xml", "text/css", "image/svg+xml", "application/xml", "text/xml":
		return true
	default:
		return false
	}
}

func isAllowedUploadType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	switch {
	case strings.HasPrefix(value, "image/"):
		return true
	case strings.HasPrefix(value, "video/"):
		return true
	case strings.HasPrefix(value, "audio/"):
		return true
	case value == "application/pdf",
		value == "application/zip",
		value == "application/x-zip-compressed",
		value == "application/x-7z-compressed",
		value == "application/x-rar-compressed",
		value == "application/json",
		value == "text/plain",
		value == "text/csv",
		value == "text/markdown",
		value == "application/javascript",
		value == "text/javascript",
		value == "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		value == "application/msword",
		value == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		value == "application/vnd.ms-excel",
		value == "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		value == "application/vnd.ms-powerpoint":
		return true
	default:
		return false
	}
}

func contentTypeByExtension(ext string) string {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4", ".m4v":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".m4a", ".aac":
		return "audio/mp4"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".7z":
		return "application/x-7z-compressed"
	case ".json":
		return "application/json"
	case ".csv":
		return "text/csv"
	case ".md":
		return "text/markdown"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".txt", ".log", ".go", ".py", ".ts", ".tsx", ".jsx", ".sql", ".yaml", ".yml", ".toml", ".ini":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

func inferAttachmentKind(contentType, filename string) string {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.HasPrefix(ct, "image/"):
		return "image"
	case strings.HasPrefix(ct, "video/"):
		return "video"
	case strings.HasPrefix(ct, "audio/"):
		return "audio"
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return "image"
	case ".mp4", ".webm", ".mov", ".mkv":
		return "video"
	case ".mp3", ".wav", ".ogg", ".m4a":
		return "audio"
	default:
		return "file"
	}
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "vip", "ai":
		return strings.ToLower(strings.TrimSpace(role))
	default:
		return "member"
	}
}

func defaultAccentByRole(role string) string {
	switch normalizeRole(role) {
	case "admin":
		return "#ff5f7a"
	case "vip":
		return "#ffd15c"
	case "ai":
		return "#00f7ff"
	default:
		return "#7bff00"
	}
}

func isPrivileged(role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	return role == "owner" || role == "admin"
}

func canManageOwnedRecord(user model.PanelUser, createdBy int64) bool {
	return isPrivileged(user.Role) || user.ID == createdBy
}

func canManagePanelMessage(user model.PanelUser, msg model.PanelMessage) bool {
	return canManageOwnedRecord(user, msg.AuthorID)
}

func canPinPanelMessage(user model.PanelUser, msg model.PanelMessage) bool {
	return canManagePanelMessage(user, msg)
}

func idsToSet(values []int64) map[int64]bool {
	out := make(map[int64]bool, len(values))
	for _, value := range values {
		if value > 0 {
			out[value] = true
		}
	}
	return out
}

func setToIDs(values map[int64]bool) []int64 {
	if len(values) == 0 {
		return []int64{}
	}
	out := make([]int64, 0, len(values))
	for value, enabled := range values {
		if enabled && value > 0 {
			out = append(out, value)
		}
	}
	return out
}

func ownerUsername() string {
	return strings.ToLower(strings.TrimSpace(envOr("PAINEL_DIEF_OWNER_USER", defaultOwnerUsername)))
}

func ownerEmail() string {
	return strings.ToLower(strings.TrimSpace(envOr("PAINEL_DIEF_OWNER_EMAIL", defaultOwnerEmail)))
}

func ownerPassword() string {
	return strings.TrimSpace(envOr("PAINEL_DIEF_OWNER_PASSWORD", defaultOwnerPassword))
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func welcomeText(slug string) string {
	switch slug {
	case "chat-geral":
		return "Bem-vindo ao Painel Dief. Joga a resenha aqui e segura o caos."
	case "fotos":
		return "Posta print, arte, setup ou meme. Se vier horroroso, pelo menos vem com estilo."
	case "arquivos":
		return "Area de drop local pra arquivos, docs e packs do servidor."
	case "chat-priv":
		return "Sala privada pronta. Quem nao tiver senha vai bater na porta e voltar triste."
	case "apps-lab":
		return "Espaco tecnico pra apps, codigos, automacoes e testes no terminal."
	case "lounge-vip":
		return "Lounge VIP no ar. Aqui a conversa eh mais fina e o acesso eh seletivo."
	default:
		return "Sala pronta no Painel Dief."
	}
}

func roomIDsByAccess(rooms []model.PanelRoom, access map[string]string, includeLocked bool) []int64 {
	ids := make([]int64, 0, len(rooms))
	for _, room := range rooms {
		accessType := access[room.Slug]
		if accessType == "open" || accessType == "admin" || accessType == "dm" || (includeLocked && accessType == "locked") || (includeLocked && accessType == "vip") {
			ids = append(ids, room.ID)
		}
	}
	return ids
}

func describeMessageLog(msg model.PanelMessage) string {
	if msg.Attachment != nil && strings.TrimSpace(msg.Body) == "" {
		return fmt.Sprintf("enviou anexo %s", msg.Attachment.Name)
	}
	if msg.Attachment != nil {
		return fmt.Sprintf("mandou mensagem com anexo %s", msg.Attachment.Name)
	}
	return fmt.Sprintf("disparou mensagem: %s", summarizeText(msg.Body, 96))
}

func summarizeText(raw string, limit int) string {
	raw = strings.TrimSpace(raw)
	if limit <= 0 || len(raw) <= limit {
		return raw
	}
	return raw[:limit-3] + "..."
}

func sanitizeMessageKind(kind string, attachment *model.PanelAttachment) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	switch kind {
	case "text", "image", "video", "audio", "file":
		return kind
	}
	if attachment != nil {
		return sanitizeMessageKind(attachment.Kind, nil)
	}
	return "text"
}

func sanitizeReaction(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if len([]rune(raw)) > 8 {
		return ""
	}
	return raw
}

func directRoomSlug(leftID, rightID int64) string {
	return fmt.Sprintf("dm-%d-%d", leftID, rightID)
}

func directRoomName(leftName, rightName string) string {
	leftName = strings.TrimSpace(leftName)
	rightName = strings.TrimSpace(rightName)
	if leftName == "" {
		leftName = "Usuario"
	}
	if rightName == "" {
		rightName = "Usuario"
	}
	return "DM // " + leftName + " + " + rightName
}
