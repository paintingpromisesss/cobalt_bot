package service

import (
	"errors"
	"sync"
)

var ErrUserJobInProgress = errors.New("user already has active job")

type RequestQueue struct {
	mu     sync.Mutex
	active map[int64]struct{}
}

func NewRequestQueue() *RequestQueue {
	return &RequestQueue{
		active: make(map[int64]struct{}),
	}
}

func (q *RequestQueue) TryStart(userID int64) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.active[userID]; exists {
		return false
	}

	q.active[userID] = struct{}{}
	return true
}

func (q *RequestQueue) Finish(userID int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.active, userID)
}

func (q *RequestQueue) IsRunning(userID int64) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, exists := q.active[userID]
	return exists
}

func (q *RequestQueue) ActiveCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.active)
}

func (q *RequestQueue) Run(userID int64, fn func() error) error {
	if ok := q.TryStart(userID); !ok {
		return ErrUserJobInProgress
	}
	defer q.Finish(userID)

	return fn()
}
