package panel

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"testing"
	"time"

	"universald/internal/db"
	"universald/internal/model"
)

func TestMessageLifecycleFeatures(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	svc := NewService(store, filepath.Join(t.TempDir(), "uploads"))
	if err := svc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	viewer, session, err := svc.Login(ownerUsername(), ownerPassword())
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}

	bootstrap, err := svc.Bootstrap(viewer, session.ID)
	if err != nil {
		t.Fatalf("bootstrap session: %v", err)
	}

	room := findRoomBySlug(bootstrap.Rooms, "chat-geral")
	if room.ID == 0 {
		t.Fatalf("chat-geral not found in bootstrap")
	}

	posted, err := svc.PostMessage(viewer, session.ID, room.ID, "bah @todos mensagem inicial", "text", nil, 0)
	if err != nil {
		t.Fatalf("post message: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	edited, err := svc.EditMessage(viewer, session.ID, room.ID, posted.ID, "mensagem editada e mais alinhada")
	if err != nil {
		t.Fatalf("edit message: %v", err)
	}
	if edited.UpdatedAt == nil {
		t.Fatalf("expected edited message to include updatedAt")
	}

	favorited, favoriteState, err := svc.ToggleFavorite(viewer, session.ID, room.ID, posted.ID)
	if err != nil {
		t.Fatalf("favorite message: %v", err)
	}
	if !favoriteState || !favorited.ViewerFavorited {
		t.Fatalf("expected message to be favorited after toggle")
	}

	pinned, pinState, err := svc.TogglePin(viewer, session.ID, room.ID, posted.ID)
	if err != nil {
		t.Fatalf("pin message: %v", err)
	}
	if !pinState || !pinned.IsPinned {
		t.Fatalf("expected message to be pinned after toggle")
	}

	pins, err := svc.ListPinnedMessages(viewer, session.ID, room.ID, 10)
	if err != nil {
		t.Fatalf("list pinned messages: %v", err)
	}
	if len(pins) != 1 || pins[0].ID != posted.ID {
		t.Fatalf("expected pinned list to contain the edited message")
	}

	messages, _, err := svc.ListMessages(viewer, session.ID, room.ID, 30)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	current := findMessageByID(messages, posted.ID)
	if current.ID == 0 {
		t.Fatalf("message not found after list")
	}
	if !current.IsPinned || !current.ViewerFavorited || current.UpdatedAt == nil {
		t.Fatalf("expected hydrated flags on message, got pinned=%v favorite=%v edited=%v", current.IsPinned, current.ViewerFavorited, current.UpdatedAt != nil)
	}

	if err := svc.DeleteMessage(viewer, session.ID, room.ID, posted.ID); err != nil {
		t.Fatalf("delete message: %v", err)
	}

	messages, _, err = svc.ListMessages(viewer, session.ID, room.ID, 30)
	if err != nil {
		t.Fatalf("list messages after delete: %v", err)
	}
	if findMessageByID(messages, posted.ID).ID != 0 {
		t.Fatalf("expected deleted message to disappear from room")
	}
}

func TestSocialProfileAndBlockFlow(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "panel-social.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	svc := NewService(store, filepath.Join(t.TempDir(), "uploads"))
	if err := svc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	owner, ownerSession, err := svc.Login(ownerUsername(), ownerPassword())
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}

	target, err := svc.CreateUser(owner, "lucas", "lucas@paineldief.local", "Lucas#2026", "Lucas", "member")
	if err != nil {
		t.Fatalf("create target user: %v", err)
	}
	targetLogin, targetSession, err := svc.Login(target.Username, "Lucas#2026")
	if err != nil {
		t.Fatalf("login target: %v", err)
	}

	bootstrap, err := svc.Bootstrap(owner, ownerSession.ID)
	if err != nil {
		t.Fatalf("bootstrap owner: %v", err)
	}
	room := findRoomBySlug(bootstrap.Rooms, "chat-geral")
	if room.ID == 0 {
		t.Fatalf("chat-geral not found")
	}

	if _, err := svc.PostMessage(targetLogin, targetSession.ID, room.ID, "mensagem do lucas no geral", "text", nil, 0); err != nil {
		t.Fatalf("post target message: %v", err)
	}

	muted, err := svc.ToggleMute(owner, target.ID)
	if err != nil {
		t.Fatalf("mute target: %v", err)
	}
	if !muted {
		t.Fatalf("expected muted state true")
	}

	blocked, err := svc.ToggleBlock(owner, target.ID)
	if err != nil {
		t.Fatalf("block target: %v", err)
	}
	if !blocked {
		t.Fatalf("expected blocked state true")
	}

	profile, err := svc.GetSocialProfile(owner, target.ID)
	if err != nil {
		t.Fatalf("get social profile: %v", err)
	}
	if !profile.User.BlockedByViewer || !profile.User.MutedByViewer {
		t.Fatalf("expected profile flags for blocked/muted user")
	}
	if profile.CanDM {
		t.Fatalf("expected blocked profile to disable DM")
	}

	items, _, err := svc.ListMessages(owner, ownerSession.ID, room.ID, 20)
	if err != nil {
		t.Fatalf("list room messages: %v", err)
	}
	var blockedSeen bool
	for _, item := range items {
		if item.AuthorID == target.ID && item.BlockedByViewer {
			blockedSeen = true
			break
		}
	}
	if !blockedSeen {
		t.Fatalf("expected blocked user's message to be flagged in room")
	}

	if _, err := svc.OpenDirectRoom(owner, target.ID); err == nil {
		t.Fatalf("expected direct room open to fail while blocked")
	}
}

func TestUploadValidationAndAttachmentSearch(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "panel-upload.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	svc := NewService(store, filepath.Join(t.TempDir(), "uploads"))
	if err := svc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	viewer, session, err := svc.Login(ownerUsername(), ownerPassword())
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}

	bootstrap, err := svc.Bootstrap(viewer, session.ID)
	if err != nil {
		t.Fatalf("bootstrap owner: %v", err)
	}
	room := findRoomBySlug(bootstrap.Rooms, "fotos")
	if room.ID == 0 {
		t.Fatalf("fotos not found")
	}

	var pngBuf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 123, G: 255, B: 0, A: 255})
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	attachment, err := svc.SaveUpload(viewer, "mapa-do-painel.png", "image/png", bytes.NewReader(pngBuf.Bytes()))
	if err != nil {
		t.Fatalf("save upload: %v", err)
	}
	if attachment.Kind != "image" || attachment.Width != 2 || attachment.Height != 2 {
		t.Fatalf("unexpected attachment metadata: %+v", attachment)
	}

	msg, err := svc.PostMessage(viewer, session.ID, room.ID, "print de teste", attachment.Kind, &attachment, 0)
	if err != nil {
		t.Fatalf("post upload message: %v", err)
	}
	if msg.Attachment == nil || msg.Attachment.Name != attachment.Name {
		t.Fatalf("expected attachment to be persisted in message")
	}

	search, err := svc.Search(viewer, session.ID, "mapa-do-painel", 10)
	if err != nil {
		t.Fatalf("search attachment by name: %v", err)
	}
	if findMessageByID(search.Messages, msg.ID).ID == 0 {
		t.Fatalf("expected attachment search to find message")
	}

	if _, err := svc.SaveUpload(viewer, "script-malvado.exe", "application/octet-stream", bytes.NewReader([]byte("nope"))); err == nil {
		t.Fatalf("expected unsupported upload to fail")
	}
	if _, err := svc.SaveUpload(viewer, "payload.html", "text/html", bytes.NewReader([]byte("<html>x</html>"))); err == nil {
		t.Fatalf("expected active html upload to fail")
	}
}

func TestLoginThrottleBlocksBurstAndRecovers(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "panel-login.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	svc := NewService(store, filepath.Join(t.TempDir(), "uploads"))
	if err := svc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	var lastErr error
	for attempt := 0; attempt < loginBurst; attempt++ {
		_, _, lastErr = svc.Login(ownerUsername(), "senha-ruim")
	}
	if !errors.Is(lastErr, ErrLoginRateLimited) {
		t.Fatalf("expected rate limit after burst, got %v", lastErr)
	}

	if _, _, err := svc.Login(ownerUsername(), ownerPassword()); !errors.Is(err, ErrLoginRateLimited) {
		t.Fatalf("expected valid login to remain blocked during cooldown, got %v", err)
	}

	key := svc.loginThrottleKey(ownerUsername())
	state := svc.loginAttempts[key]
	state.BlockedUntil = time.Now().UTC().Add(-time.Second)
	svc.loginAttempts[key] = state

	if _, _, err := svc.Login(ownerUsername(), ownerPassword()); err != nil {
		t.Fatalf("expected login to recover after cooldown, got %v", err)
	}
}

func TestRunMaintenancePrunesRuntimeStateAndExpiredData(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "panel-maintenance.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	svc := NewService(store, filepath.Join(t.TempDir(), "uploads"))
	if err := svc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	viewer, _, err := svc.Login(ownerUsername(), ownerPassword())
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}

	now := time.Now().UTC()
	expiredSession := model.PanelSession{
		ID:        "expired-session",
		UserID:    viewer.ID,
		CreatedAt: now.Add(-8 * time.Hour),
		ExpiresAt: now.Add(-time.Hour),
	}
	if err := store.SavePanelSession(expiredSession); err != nil {
		t.Fatalf("save expired session: %v", err)
	}
	if err := store.UpsertPanelPresence(viewer.ID, 0, "online", now.Add(-72*time.Hour)); err != nil {
		t.Fatalf("save stale presence: %v", err)
	}

	svc.unlocked["expired-session"] = map[int64]time.Time{99: now.Add(-time.Minute)}
	svc.typing[77] = map[int64]model.PanelTyping{
		viewer.ID: {RoomID: 77, UserID: viewer.ID, DisplayName: viewer.DisplayName, ExpiresAt: now.Add(-time.Second)},
	}
	svc.flood[viewer.ID] = []time.Time{now.Add(-2 * floodWindow)}
	svc.loginAttempts["owner"] = loginAttemptState{
		Hits:         []time.Time{now.Add(-2 * loginWindow)},
		BlockedUntil: now.Add(-time.Second),
	}

	if err := svc.RunMaintenance(now); err != nil {
		t.Fatalf("run maintenance: %v", err)
	}

	if _, _, err := svc.Authenticate(expiredSession.ID); err == nil {
		t.Fatalf("expected expired session to be gone after maintenance")
	}
	if len(svc.unlocked) != 0 {
		t.Fatalf("expected unlocked cache to be pruned, got %+v", svc.unlocked)
	}
	if len(svc.typing) != 0 {
		t.Fatalf("expected typing cache to be pruned, got %+v", svc.typing)
	}
	if len(svc.flood) != 0 {
		t.Fatalf("expected flood cache to be pruned, got %+v", svc.flood)
	}
	if len(svc.loginAttempts) != 0 {
		t.Fatalf("expected login throttle cache to be pruned, got %+v", svc.loginAttempts)
	}

	summary, err := store.PanelOpsSummary(now, 365*24*time.Hour)
	if err != nil {
		t.Fatalf("ops summary: %v", err)
	}
	if summary.OnlineUsers != 0 {
		t.Fatalf("expected stale presence to be deleted, got onlineUsers=%d", summary.OnlineUsers)
	}
}

func TestPanelEventsCreateAndRSVPFlow(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "panel-events.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	svc := NewService(store, filepath.Join(t.TempDir(), "uploads"))
	if err := svc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	owner, ownerSession, err := svc.Login(ownerUsername(), ownerPassword())
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}
	member, err := svc.CreateUser(owner, "nina", "nina@paineldief.local", "Nina#2026", "Nina", "member")
	if err != nil {
		t.Fatalf("create member: %v", err)
	}
	member, memberSession, err := svc.Login(member.Username, "Nina#2026")
	if err != nil {
		t.Fatalf("login member: %v", err)
	}

	ownerBootstrap, err := svc.Bootstrap(owner, ownerSession.ID)
	if err != nil {
		t.Fatalf("bootstrap owner: %v", err)
	}
	room := findRoomBySlug(ownerBootstrap.Rooms, "chat-geral")
	if room.ID == 0 {
		t.Fatalf("chat-geral not found")
	}

	event, err := svc.CreateEvent(owner, ownerSession.ID, "Ranked da madrugada", "Fechar squad e subir ranking.", room.ID, time.Now().UTC().Add(2*time.Hour))
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if event.ID == 0 || event.RoomID != room.ID {
		t.Fatalf("expected created event with room, got %+v", event)
	}

	joinedEvent, joined, err := svc.ToggleEventRSVP(member, memberSession.ID, event.ID)
	if err != nil {
		t.Fatalf("toggle rsvp: %v", err)
	}
	if !joined || !joinedEvent.ViewerJoined || joinedEvent.RSVPCount != 1 {
		t.Fatalf("expected member to join event, got joined=%v event=%+v", joined, joinedEvent)
	}

	memberBootstrap, err := svc.Bootstrap(member, memberSession.ID)
	if err != nil {
		t.Fatalf("bootstrap member: %v", err)
	}
	if len(memberBootstrap.Events) != 1 {
		t.Fatalf("expected one event in bootstrap, got %d", len(memberBootstrap.Events))
	}
	if !memberBootstrap.Events[0].ViewerJoined || memberBootstrap.Events[0].RSVPCount != 1 {
		t.Fatalf("expected joined event in bootstrap, got %+v", memberBootstrap.Events[0])
	}

	if err := svc.DeleteEvent(member, memberSession.ID, event.ID); err == nil {
		t.Fatalf("expected non-creator member delete to be rejected")
	}
	if err := svc.DeleteEvent(owner, ownerSession.ID, event.ID); err != nil {
		t.Fatalf("delete event: %v", err)
	}
	memberBootstrap, err = svc.Bootstrap(member, memberSession.ID)
	if err != nil {
		t.Fatalf("bootstrap member after delete: %v", err)
	}
	if len(memberBootstrap.Events) != 0 {
		t.Fatalf("expected event list empty after delete, got %d", len(memberBootstrap.Events))
	}
}

func TestUpdateProfileSanitizesAvatarAndDisplayName(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "panel-profile.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	svc := NewService(store, filepath.Join(t.TempDir(), "uploads"))
	if err := svc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	viewer, _, err := svc.Login(ownerUsername(), ownerPassword())
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}

	if _, err := svc.UpdateProfile(viewer, "", "bio enxuta", "matrix", "#7bff00", "javascript:alert(1)", "online", "status curto"); err == nil {
		t.Fatalf("expected invalid avatar url to be rejected")
	}

	updated, err := svc.UpdateProfile(viewer, "   ", "bio alinhada", "cobalt", "#54b8ff", "/uploads/avatar.png", "away", "jogando e montando a base")
	if err != nil {
		t.Fatalf("update profile with local upload avatar: %v", err)
	}
	if updated.DisplayName == "" {
		t.Fatalf("expected display name fallback to be preserved")
	}
	if updated.AvatarURL != "/uploads/avatar.png" {
		t.Fatalf("expected sanitized local avatar url, got %q", updated.AvatarURL)
	}
	if updated.Status != "away" {
		t.Fatalf("expected sanitized presence status, got %q", updated.Status)
	}
	if updated.StatusText != "jogando e montando a base" {
		t.Fatalf("expected status text to persist, got %q", updated.StatusText)
	}
}

func TestPanelPollCreateAndVoteFlow(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "panel-polls.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	svc := NewService(store, filepath.Join(t.TempDir(), "uploads"))
	if err := svc.EnsureBootstrapped(); err != nil {
		t.Fatalf("bootstrap panel: %v", err)
	}

	owner, ownerSession, err := svc.Login(ownerUsername(), ownerPassword())
	if err != nil {
		t.Fatalf("login owner: %v", err)
	}
	member, err := svc.CreateUser(owner, "bia", "bia@paineldief.local", "Bia#2026", "Bia", "member")
	if err != nil {
		t.Fatalf("create member: %v", err)
	}
	member, memberSession, err := svc.Login(member.Username, "Bia#2026")
	if err != nil {
		t.Fatalf("login member: %v", err)
	}

	bootstrap, err := svc.Bootstrap(owner, ownerSession.ID)
	if err != nil {
		t.Fatalf("bootstrap owner: %v", err)
	}
	room := findRoomBySlug(bootstrap.Rooms, "chat-geral")
	if room.ID == 0 {
		t.Fatalf("chat-geral not found")
	}

	poll, err := svc.CreatePoll(owner, ownerSession.ID, room.ID, "Qual squad sobe hoje?", []string{"Valorant", "Fortnite", "Valorant", "CS2"})
	if err != nil {
		t.Fatalf("create poll: %v", err)
	}
	if poll.ID == 0 || len(poll.Options) != 3 {
		t.Fatalf("expected persisted poll with deduplicated options, got %+v", poll)
	}

	votedPoll, voted, err := svc.VotePoll(member, memberSession.ID, room.ID, poll.ID, poll.Options[1].ID)
	if err != nil {
		t.Fatalf("vote poll: %v", err)
	}
	if !voted || votedPoll.ViewerOptionID != poll.Options[1].ID || votedPoll.TotalVotes != 1 {
		t.Fatalf("expected member vote to persist, got voted=%v poll=%+v", voted, votedPoll)
	}

	listed, err := svc.ListPolls(member, memberSession.ID, room.ID, 10)
	if err != nil {
		t.Fatalf("list polls: %v", err)
	}
	if len(listed) != 1 || listed[0].ViewerOptionID != poll.Options[1].ID {
		t.Fatalf("expected listed polls to include viewer vote, got %+v", listed)
	}

	clearedPoll, voted, err := svc.VotePoll(member, memberSession.ID, room.ID, poll.ID, poll.Options[1].ID)
	if err != nil {
		t.Fatalf("toggle vote off: %v", err)
	}
	if voted || clearedPoll.TotalVotes != 0 || clearedPoll.ViewerOptionID != 0 {
		t.Fatalf("expected second vote to clear selection, got voted=%v poll=%+v", voted, clearedPoll)
	}

	if err := svc.DeletePoll(member, memberSession.ID, room.ID, poll.ID); err == nil {
		t.Fatalf("expected non-creator member delete to be rejected")
	}
	if err := svc.DeletePoll(owner, ownerSession.ID, room.ID, poll.ID); err != nil {
		t.Fatalf("delete poll: %v", err)
	}
	listed, err = svc.ListPolls(member, memberSession.ID, room.ID, 10)
	if err != nil {
		t.Fatalf("list polls after delete: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected no polls after delete, got %+v", listed)
	}
}

func findRoomBySlug(items []model.PanelRoom, slug string) model.PanelRoom {
	for _, item := range items {
		if item.Slug == slug {
			return item
		}
	}
	return model.PanelRoom{}
}

func findMessageByID(items []model.PanelMessage, id int64) model.PanelMessage {
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	return model.PanelMessage{}
}
