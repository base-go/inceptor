// Package auth provides authentication for the Inceptor API.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

const DefaultPassword = "inceptor"

// Session represents an authenticated session
type Session struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Manager handles authentication and sessions
type Manager struct {
	passwordHash    string
	isDefaultPassword bool
	sessions        map[string]*Session
	mu              sync.RWMutex
	onPasswordChange func(hash string) // callback to persist password
}

// NewManager creates a new auth manager
func NewManager(passwordHash string, onPasswordChange func(hash string)) *Manager {
	m := &Manager{
		sessions:         make(map[string]*Session),
		onPasswordChange: onPasswordChange,
	}

	if passwordHash == "" {
		// No password set, use default
		m.passwordHash = HashPassword(DefaultPassword)
		m.isDefaultPassword = true
	} else {
		m.passwordHash = passwordHash
		m.isDefaultPassword = passwordHash == HashPassword(DefaultPassword)
	}

	return m
}

// HashPassword hashes a password using SHA256
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// ValidatePassword checks if the password matches the stored hash
func (m *Manager) ValidatePassword(password string) bool {
	return HashPassword(password) == m.passwordHash
}

// NeedsPasswordChange returns true if using default password
func (m *Manager) NeedsPasswordChange() bool {
	return m.isDefaultPassword
}

// CreateSession creates a new session for authenticated user
func (m *Manager) CreateSession() (*Session, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}

	session := &Session{
		Token:     hex.EncodeToString(token),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour sessions
	}

	m.mu.Lock()
	m.sessions[session.Token] = session
	m.mu.Unlock()

	return session, nil
}

// ValidateSession checks if a session token is valid
func (m *Manager) ValidateSession(token string) bool {
	if token == "" {
		return false
	}

	m.mu.RLock()
	session, exists := m.sessions[token]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		m.DeleteSession(token)
		return false
	}

	return true
}

// ChangePassword updates the password
func (m *Manager) ChangePassword(oldPassword, newPassword string) bool {
	if !m.ValidatePassword(oldPassword) {
		return false
	}
	if newPassword == "" || len(newPassword) < 4 {
		return false
	}

	m.passwordHash = HashPassword(newPassword)
	m.isDefaultPassword = false

	// Persist the new password hash
	if m.onPasswordChange != nil {
		m.onPasswordChange(m.passwordHash)
	}

	return true
}

// DeleteSession removes a session
func (m *Manager) DeleteSession(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

// GetPasswordHash returns the current password hash
func (m *Manager) GetPasswordHash() string {
	return m.passwordHash
}

// CleanupExpiredSessions removes expired sessions
func (m *Manager) CleanupExpiredSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for token, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, token)
		}
	}
}
