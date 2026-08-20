package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/podcctv/detective-chicken/internal/model"
)

var (
	ErrConflict           = errors.New("conflict")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRegistrationClosed = errors.New("registration is closed")
)

type UserAccount struct {
	User         model.User `json:"user"`
	TenantID     string     `json:"tenant_id"`
	PasswordHash []byte     `json:"password_hash"`
}

type Session struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PasswordReset struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func normalizeUsername(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func validUsername(value string) bool {
	if len(value) < 3 || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("._-", r) {
			continue
		}
		return false
	}
	return true
}

func passwordHash(password string) ([]byte, error) {
	if len(password) < 10 || len(password) > 128 {
		return nil, errors.New("password must be 10 to 128 characters")
	}
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func tokenKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (m *Memory) PublicSettings() model.PublicSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return model.PublicSettings{RegistrationEnabled: m.registrationEnabled, Bootstrapped: len(m.users) > 0}
}

func (m *Memory) RegisterUser(username, displayName, password string) (model.User, error) {
	username = normalizeUsername(username)
	if !validUsername(username) {
		return model.User{}, errors.New("username must be 3 to 64 letters, numbers, dots, dashes or underscores")
	}
	hash, err := passwordHash(password)
	if err != nil {
		return model.User{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.usernames[username]; exists {
		return model.User{}, ErrConflict
	}
	first := len(m.users) == 0
	if !first && !m.registrationEnabled {
		return model.User{}, ErrRegistrationClosed
	}
	now := time.Now().UTC()
	role := "user"
	if first {
		role = "admin"
	}
	if strings.TrimSpace(displayName) == "" {
		displayName = username
	}
	u := model.User{ID: randomID("usr"), Username: username, DisplayName: strings.TrimSpace(displayName), Role: role, CreatedAt: now, UpdatedAt: now}
	m.users[u.ID] = UserAccount{User: u, TenantID: "tenant_default", PasswordHash: hash}
	m.usernames[username] = u.ID
	if first {
		m.registrationEnabled = false
	}
	m.persistLocked()
	return u, nil
}

func (m *Memory) CreateSession(username, password string) (model.User, string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.usernames[normalizeUsername(username)]
	if !ok {
		return model.User{}, "", time.Time{}, ErrInvalidCredentials
	}
	account := m.users[id]
	if bcrypt.CompareHashAndPassword(account.PasswordHash, []byte(password)) != nil {
		return model.User{}, "", time.Time{}, ErrInvalidCredentials
	}
	token := randomID("sess") + randomID("")
	expires := time.Now().UTC().Add(30 * 24 * time.Hour)
	m.sessions[tokenKey(token)] = Session{UserID: id, ExpiresAt: expires}
	m.persistLocked()
	return account.User, token, expires, nil
}

func (m *Memory) UserBySession(token string) (model.User, string, error) {
	if token == "" {
		return model.User{}, "", ErrInvalidCredentials
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := tokenKey(token)
	session, ok := m.sessions[key]
	if !ok || time.Now().UTC().After(session.ExpiresAt) {
		if ok {
			delete(m.sessions, key)
			m.persistLocked()
		}
		return model.User{}, "", ErrInvalidCredentials
	}
	account, ok := m.users[session.UserID]
	if !ok {
		return model.User{}, "", ErrInvalidCredentials
	}
	return account.User, account.TenantID, nil
}

func (m *Memory) DeleteSession(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, tokenKey(token))
	m.persistLocked()
}

func (m *Memory) ChangePassword(userID, currentPassword, newPassword string) error {
	hash, err := passwordHash(newPassword)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	account, ok := m.users[userID]
	if !ok || bcrypt.CompareHashAndPassword(account.PasswordHash, []byte(currentPassword)) != nil {
		return ErrInvalidCredentials
	}
	account.PasswordHash = hash
	account.User.UpdatedAt = time.Now().UTC()
	m.users[userID] = account
	m.deleteUserSessionsLocked(userID)
	m.persistLocked()
	return nil
}

func (m *Memory) CreatePasswordReset(userID string) (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[userID]; !ok {
		return "", time.Time{}, ErrNotFound
	}
	token := randomID("reset") + randomID("")
	expires := time.Now().UTC().Add(30 * time.Minute)
	m.passwordResets[tokenKey(token)] = PasswordReset{UserID: userID, ExpiresAt: expires}
	m.persistLocked()
	return token, expires, nil
}

func (m *Memory) CompletePasswordReset(token, newPassword string) error {
	hash, err := passwordHash(newPassword)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := tokenKey(token)
	reset, ok := m.passwordResets[key]
	if !ok || time.Now().UTC().After(reset.ExpiresAt) {
		delete(m.passwordResets, key)
		return ErrInvalidCredentials
	}
	account, ok := m.users[reset.UserID]
	if !ok {
		return ErrNotFound
	}
	account.PasswordHash = hash
	account.User.UpdatedAt = time.Now().UTC()
	m.users[reset.UserID] = account
	delete(m.passwordResets, key)
	m.deleteUserSessionsLocked(reset.UserID)
	m.persistLocked()
	return nil
}

func (m *Memory) deleteUserSessionsLocked(userID string) {
	for key, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, key)
		}
	}
}

func (m *Memory) Users() []model.User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	users := make([]model.User, 0, len(m.users))
	for _, account := range m.users {
		users = append(users, account.User)
	}
	sort.Slice(users, func(i, j int) bool { return users[i].CreatedAt.Before(users[j].CreatedAt) })
	return users
}

func (m *Memory) UpdateUserRole(actorID, userID, role string) (model.User, error) {
	if role != "admin" && role != "user" {
		return model.User{}, errors.New("role must be admin or user")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	account, ok := m.users[userID]
	if !ok {
		return model.User{}, ErrNotFound
	}
	if account.User.Role == "admin" && role == "user" {
		admins := 0
		for _, candidate := range m.users {
			if candidate.User.Role == "admin" {
				admins++
			}
		}
		if admins == 1 || actorID == userID {
			return model.User{}, ErrForbidden
		}
	}
	account.User.Role = role
	account.User.UpdatedAt = time.Now().UTC()
	m.users[userID] = account
	m.persistLocked()
	return account.User, nil
}

func (m *Memory) SetRegistrationEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registrationEnabled = enabled
	m.persistLocked()
}
