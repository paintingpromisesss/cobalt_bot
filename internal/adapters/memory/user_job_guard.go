package memory

import (
	"errors"
	"sync"
)

var ErrUserJobInProgress = errors.New("user already has active job")

type UserJobGuard struct {
	mu     sync.Mutex
	active map[int64]struct{}
}

func NewUserJobGuard() *UserJobGuard {
	return &UserJobGuard{
		active: make(map[int64]struct{}),
	}
}

func (q *UserJobGuard) TryStart(userID int64) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.active[userID]; exists {
		return false
	}

	q.active[userID] = struct{}{}
	return true
}

func (q *UserJobGuard) Finish(userID int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.active, userID)
}

func (q *UserJobGuard) IsRunning(userID int64) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, exists := q.active[userID]
	return exists
}

func (q *UserJobGuard) ActiveCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.active)
}

func (q *UserJobGuard) Run(userID int64, fn func() error) error {
	if ok := q.TryStart(userID); !ok {
		return ErrUserJobInProgress
	}
	defer q.Finish(userID)

	return fn()
}
