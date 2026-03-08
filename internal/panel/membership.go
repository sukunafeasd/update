package panel

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"

	"universald/internal/model"
)

const joinAccessLifetime = 72 * time.Hour

func (s *Service) RequestJoinAccess(email, displayName, note string) (model.PanelJoinRequest, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := mail.ParseAddress(email); err != nil {
		return model.PanelJoinRequest{}, errors.New("bah, esse gmail veio mais torto que senha errada")
	}
	displayName = sanitizeDisplayName(displayName, strings.Split(email, "@")[0], strings.Split(email, "@")[0])
	note = sanitizeBio(note)

	if _, err := s.store.GetPanelUserByLogin(email); err == nil {
		return model.PanelJoinRequest{}, errors.New("esse email ja ta dentro da base, vivente")
	}

	latest, err := s.store.GetLatestPanelJoinRequestByEmail(email)
	if err == nil {
		switch latest.Status {
		case "pending":
			return model.PanelJoinRequest{}, errors.New("teu pedido ja ta na mesa do dono. segura a ansiedade, guri")
		case "approved":
			if latest.AccessCode != "" && latest.AccessCodeExpires != nil && latest.AccessCodeExpires.After(time.Now().UTC()) {
				return model.PanelJoinRequest{}, errors.New("bah, teu acesso ja foi aprovado. usa o codigo que te mandaram")
			}
		case "rejected":
			if strings.TrimSpace(latest.ReviewNote) != "" {
				return model.PanelJoinRequest{}, errors.New(strings.TrimSpace(latest.ReviewNote))
			}
			return model.PanelJoinRequest{}, errors.New("bah... teu pedido bateu na porteira e voltou")
		case "completed":
			return model.PanelJoinRequest{}, errors.New("esse email ja virou membro do painel")
		}
	}

	request, err := s.store.CreatePanelJoinRequest(model.PanelJoinRequest{
		Email:       email,
		DisplayName: displayName,
		Note:        note,
		Status:      "pending",
	})
	if err != nil {
		return model.PanelJoinRequest{}, err
	}
	s.logGuestAction("join_request", displayName, 0, "", "pediu entrada na base pelo login")
	s.bumpVersion()
	return request, nil
}

func (s *Service) ListJoinRequests(actor model.PanelUser) ([]model.PanelJoinRequest, error) {
	if strings.ToLower(strings.TrimSpace(actor.Role)) != "owner" {
		return nil, errors.New("so o dono pode revisar pedidos de entrada")
	}
	return s.store.ListPanelJoinRequests(120)
}

func (s *Service) ReviewJoinRequest(actor model.PanelUser, requestID int64, approve bool, reviewNote string) (model.PanelJoinRequest, error) {
	if strings.ToLower(strings.TrimSpace(actor.Role)) != "owner" {
		return model.PanelJoinRequest{}, errors.New("so o dono pode revisar pedidos de entrada")
	}
	item, err := s.store.GetPanelJoinRequestByID(requestID)
	if err != nil {
		return model.PanelJoinRequest{}, err
	}
	if item.Status == "completed" {
		return model.PanelJoinRequest{}, errors.New("esse pedido ja virou conta e nao pode mais ser revisado")
	}
	now := time.Now().UTC()
	item.ReviewedAt = &now
	item.ReviewedBy = actor.ID
	item.ReviewerName = actor.DisplayName
	item.ReviewNote = sanitizeReviewNote(reviewNote)
	item.EmailSent = false
	item.AccessCode = ""
	item.AccessCodeExpires = nil
	item.ApprovedUserID = 0

	if approve {
		code, codeErr := randomJoinAccessCode()
		if codeErr != nil {
			return model.PanelJoinRequest{}, codeErr
		}
		expiresAt := now.Add(joinAccessLifetime)
		item.Status = "approved"
		item.AccessCode = code
		item.AccessCodeExpires = &expiresAt
		if strings.TrimSpace(item.ReviewNote) == "" {
			item.ReviewNote = "Aprovado. Agora finaliza teu cadastro sem fazer fiasco."
		}
		sent, sendErr := sendJoinAccessEmail(item.Email, item.DisplayName, code, expiresAt)
		if sendErr != nil {
			item.EmailSent = false
			item.ReviewNote = strings.TrimSpace(item.ReviewNote + " Email falhou, entao o dono precisa te passar o codigo manualmente.")
		} else {
			item.EmailSent = sent
		}
		updated, updateErr := s.store.UpdatePanelJoinRequest(item)
		if updateErr != nil {
			return model.PanelJoinRequest{}, updateErr
		}
		s.logAction(actor, "join_request_approve", nil, fmt.Sprintf("aprovou o pedido de %s", updated.Email))
		s.bumpVersion()
		return updated, nil
	}

	item.Status = "rejected"
	if strings.TrimSpace(item.ReviewNote) == "" {
		item.ReviewNote = pickRejectLine(item.DisplayName)
	}
	updated, err := s.store.UpdatePanelJoinRequest(item)
	if err != nil {
		return model.PanelJoinRequest{}, err
	}
	s.logAction(actor, "join_request_reject", nil, fmt.Sprintf("recusou o pedido de %s", updated.Email))
	s.bumpVersion()
	return updated, nil
}

func (s *Service) CompleteJoinAccess(email, accessCode, username, displayName, password string) (model.PanelUser, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	accessCode = strings.ToUpper(strings.TrimSpace(accessCode))
	if _, err := mail.ParseAddress(email); err != nil {
		return model.PanelUser{}, errors.New("email invalido pra completar entrada")
	}
	if accessCode == "" {
		return model.PanelUser{}, errors.New("sem codigo de acesso nao passa da guarita")
	}

	item, err := s.store.GetLatestPanelJoinRequestByEmail(email)
	if err != nil {
		return model.PanelUser{}, errors.New("nenhum pedido encontrado pra esse email")
	}
	if item.Status == "rejected" {
		return model.PanelUser{}, errors.New(strings.TrimSpace(item.ReviewNote))
	}
	if item.Status == "completed" {
		return model.PanelUser{}, errors.New("esse pedido ja virou conta. tenta logar em vez de inventar moda")
	}
	if item.Status != "approved" {
		return model.PanelUser{}, errors.New("teu pedido ainda nao foi aprovado pelo dono")
	}
	if item.AccessCodeExpires == nil || !item.AccessCodeExpires.After(time.Now().UTC()) {
		return model.PanelUser{}, errors.New("o codigo venceu. pede uma aprovacao nova pro dono")
	}
	if subtleCompare(item.AccessCode, accessCode) != 1 {
		return model.PanelUser{}, errors.New("codigo errado, vivente. ate a porteira desconfiou")
	}

	if _, err := s.store.GetPanelUserByLogin(email); err == nil {
		return model.PanelUser{}, errors.New("esse email ja foi usado por uma conta real")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return model.PanelUser{}, err
	}
	reviewerID := item.ReviewedBy
	if reviewerID <= 0 {
		if owner, ownerErr := s.store.GetPanelUserByLogin(defaultOwnerUsername); ownerErr == nil {
			reviewerID = owner.ID
		}
	}
	user, err := s.store.CreatePanelUser(model.PanelUser{
		Username:     strings.ToLower(strings.TrimSpace(username)),
		Email:        email,
		DisplayName:  sanitizeDisplayName(displayName, item.DisplayName, username),
		Role:         "member",
		Theme:        "matrix",
		AccentColor:  defaultAccentByRole("member"),
		Status:       "online",
		PasswordHash: hash,
		CreatedBy:    reviewerID,
	})
	if err != nil {
		return model.PanelUser{}, err
	}
	item.Status = "completed"
	item.ApprovedUserID = user.ID
	item.AccessCode = ""
	item.AccessCodeExpires = nil
	item.EmailSent = false
	updated, err := s.store.UpdatePanelJoinRequest(item)
	if err != nil {
		return model.PanelUser{}, err
	}
	user.PasswordHash = ""
	s.logGuestAction("join_request_complete", updated.DisplayName, 0, "", "concluiu o cadastro aprovado")
	s.bumpVersion()
	return user, nil
}

func (s *Service) UpdateUserRole(actor model.PanelUser, targetUserID int64, role string) (model.PanelUser, error) {
	if strings.ToLower(strings.TrimSpace(actor.Role)) != "owner" {
		return model.PanelUser{}, errors.New("so o dono mexe no cargo dos outros")
	}
	target, err := s.store.GetPanelUserByID(targetUserID)
	if err != nil {
		return model.PanelUser{}, err
	}
	if target.ID == actor.ID || strings.ToLower(strings.TrimSpace(target.Role)) == "owner" {
		return model.PanelUser{}, errors.New("o dono nao mexe no proprio cargo por aqui")
	}
	role = normalizeRole(role)
	if role == "ai" {
		return model.PanelUser{}, errors.New("esse cargo fica reservado pro nego dramias")
	}
	updated, err := s.store.UpdatePanelUserRole(target.ID, role)
	if err != nil {
		return model.PanelUser{}, err
	}
	updated.PasswordHash = ""
	s.logAction(actor, "user_role_update", nil, fmt.Sprintf("ajustou %s para %s", updated.DisplayName, updated.Role))
	s.bumpVersion()
	return updated, nil
}

func (s *Service) ExpelUser(actor model.PanelUser, targetUserID int64) error {
	if strings.ToLower(strings.TrimSpace(actor.Role)) != "owner" {
		return errors.New("so o dono pode expulsar usuario")
	}
	target, err := s.store.GetPanelUserByID(targetUserID)
	if err != nil {
		return err
	}
	if target.ID == actor.ID || strings.ToLower(strings.TrimSpace(target.Role)) == "owner" || strings.ToLower(strings.TrimSpace(target.Role)) == "ai" {
		return errors.New("esse usuario nao pode ser expulso por esse fluxo")
	}
	if err := s.store.DeletePanelUserCascade(target.ID); err != nil {
		return err
	}
	s.logAction(actor, "user_expel", nil, fmt.Sprintf("expulsou %s da base", target.DisplayName))
	s.bumpVersion()
	return nil
}

func pickRejectLine(displayName string) string {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "guri"
	}
	lines := []string{
		"Bah... o pedido do " + name + " bateu na cerca e voltou.",
		"Sai daqui, crianca. Hoje nao rolou tua entrada, " + name + ".",
		"O dono olhou teu pedido e disse: capaz, vivente.",
		"O outro lado da porteira nao abriu pra ti dessa vez, " + name + ".",
	}
	return lines[int(time.Now().UnixNano())%len(lines)]
}

func sanitizeReviewNote(note string) string {
	note = strings.TrimSpace(note)
	if len(note) > 180 {
		note = note[:180]
	}
	return note
}

func randomJoinAccessCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("random join code: %w", err)
	}
	for i := range buf {
		buf[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(buf), nil
}

func subtleCompare(left, right string) int {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	if len(left) != len(right) {
		return 0
	}
	var result byte
	for i := 0; i < len(left); i++ {
		result |= left[i] ^ right[i]
	}
	if result == 0 {
		return 1
	}
	return 0
}

func sendJoinAccessEmail(to, displayName, accessCode string, expiresAt time.Time) (bool, error) {
	host := strings.TrimSpace(os.Getenv("PAINEL_DIEF_SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("PAINEL_DIEF_SMTP_PORT"))
	from := strings.TrimSpace(os.Getenv("PAINEL_DIEF_SMTP_FROM"))
	user := strings.TrimSpace(os.Getenv("PAINEL_DIEF_SMTP_USER"))
	pass := strings.TrimSpace(os.Getenv("PAINEL_DIEF_SMTP_PASS"))
	if host == "" || port == "" || from == "" {
		return false, nil
	}
	subject := "Painel Dief // codigo de entrada"
	body := fmt.Sprintf("Bah %s,\n\nteu pedido foi aprovado no Painel Dief.\n\nCodigo: %s\nValido ate: %s\n\nSe nao foi tu, ignora essa bomba.\n", strings.TrimSpace(displayName), accessCode, expiresAt.Format(time.RFC1123))
	msg := []byte("To: " + to + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body)
	addr := host + ":" + port
	var auth smtp.Auth
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
		return false, err
	}
	return true, nil
}
