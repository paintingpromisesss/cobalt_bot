package pickersession

import (
	"context"
	"sync"
	"time"
)

type PickerSessionManager struct {
	sessions map[string]*pickerSession
	mu       sync.Mutex
	seq      uint64
	ttl      time.Duration
}

func NewPickerSessionManager(ctx context.Context, ttl time.Duration, cleanupInterval time.Duration) *PickerSessionManager {
	if ttl <= 0 {
		panic("ttl must be positive")
	}

	if cleanupInterval <= 0 {
		panic("cleanupInterval must be positive")
	}

	m := &PickerSessionManager{
		sessions: make(map[string]*pickerSession),
		ttl:      ttl,
	}

	go m.startCleanup(ctx, cleanupInterval)

	return m
}

func (m *PickerSessionManager) DeleteSession(sessionID string, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.validateSessionLocked(sessionID, userID)
	if err != nil {
		return err
	}

	delete(m.sessions, sessionID)
	return nil
}

func (m *PickerSessionManager) startCleanup(ctx context.Context, cleanupInterval time.Duration) {
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

// validateSession func must be called with m.mu locked
func (m *PickerSessionManager) validateSessionLocked(sessionID string, userID int64) (*pickerSession, error) {
	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if session.userID != userID {
		return nil, ErrSessionForbidden
	}

	if time.Now().After(session.expiresAt) {
		delete(m.sessions, sessionID)
		return nil, ErrSessionExpired
	}

	return session, nil
}
