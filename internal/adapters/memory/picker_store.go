package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	domainpicker "github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
)

type PickerStore struct {
	sessions map[string]*pickerSession
	mu       sync.Mutex
	ttl      time.Duration
}

func NewPickerStore(ctx context.Context, ttl time.Duration, cleanupInterval time.Duration) *PickerStore {
	if ttl <= 0 {
		panic("ttl must be positive")
	}

	if cleanupInterval <= 0 {
		panic("cleanupInterval must be positive")
	}

	m := &PickerStore{
		sessions: make(map[string]*pickerSession),
		ttl:      ttl,
	}

	go m.startCleanup(ctx, cleanupInterval)

	return m
}

func (m *PickerStore) deleteSession(sessionID string, userID int64, sessionType PickerSessionType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.validateSessionLocked(sessionID, userID, sessionType)
	if err != nil {
		return err
	}

	delete(m.sessions, sessionID)
	return nil
}

func (m *PickerStore) startCleanup(ctx context.Context, cleanupInterval time.Duration) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			m.mu.Lock()
			for id, session := range m.sessions {
				if now.After(session.expiresAt) {
					delete(m.sessions, id)
				}
			}
			m.mu.Unlock()
		}
	}
}

func (m *PickerStore) validateSessionLocked(sessionID string, userID int64, sessionType PickerSessionType) (*pickerSession, error) {
	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, domainpicker.ErrSessionNotFound
	}

	if session.userID != userID {
		return nil, domainpicker.ErrSessionForbidden
	}

	if time.Now().After(session.expiresAt) {
		delete(m.sessions, sessionID)
		return nil, domainpicker.ErrSessionExpired
	}

	if session.sessionType != sessionType {
		return nil, domainpicker.ErrWrongSessionType
	}

	return session, nil
}

func (m *PickerStore) newUniqueSessionIDLocked() (string, error) {
	for {
		id, err := newSessionID()
		if err != nil {
			return "", err
		}
		if _, exists := m.sessions[id]; !exists {
			return id, nil
		}
	}
}

func newSessionID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
