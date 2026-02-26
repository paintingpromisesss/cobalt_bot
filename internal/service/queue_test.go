package service

import (
	"errors"
	"testing"
)

func TestNewRequestQueue(t *testing.T) {
	q := NewRequestQueue()
	if q == nil {
		t.Fatalf("expected queue instance")
	}
	if got := q.ActiveCount(); got != 0 {
		t.Fatalf("expected empty queue, got %d active", got)
	}
}

func TestRequestQueueTryStartAndFinishByUserID(t *testing.T) {
	q := NewRequestQueue()

	if !q.TryStart(42) {
		t.Fatalf("expected first start to succeed")
	}

	if q.TryStart(42) {
		t.Fatalf("expected second start for same user to fail")
	}

	if !q.IsRunning(42) {
		t.Fatalf("expected user to be running")
	}

	q.Finish(42)

	if q.IsRunning(42) {
		t.Fatalf("expected user to be released")
	}

	if !q.TryStart(42) {
		t.Fatalf("expected start after finish to succeed")
	}
}

func TestRequestQueueTryStartAllowsDifferentUsers(t *testing.T) {
	q := NewRequestQueue()

	if !q.TryStart(1) {
		t.Fatalf("expected user 1 to start")
	}
	if !q.TryStart(2) {
		t.Fatalf("expected user 2 to start")
	}

	if got := q.ActiveCount(); got != 2 {
		t.Fatalf("expected 2 active users, got %d", got)
	}
}

func TestRequestQueueRunSuccess(t *testing.T) {
	q := NewRequestQueue()

	called := false
	err := q.Run(7, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatalf("expected fn to be called")
	}
	if q.IsRunning(7) {
		t.Fatalf("expected user to be released after run")
	}
}

func TestRequestQueueRunRejectsParallelForSameUser(t *testing.T) {
	q := NewRequestQueue()
	if !q.TryStart(99) {
		t.Fatalf("expected lock to be taken")
	}

	err := q.Run(99, func() error { return nil })
	if !errors.Is(err, ErrUserJobInProgress) {
		t.Fatalf("expected ErrUserJobInProgress, got %v", err)
	}
}

func TestRequestQueueRunReleasesLockOnError(t *testing.T) {
	q := NewRequestQueue()
	expectedErr := errors.New("boom")

	err := q.Run(3, func() error { return expectedErr })
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped function error, got %v", err)
	}

	if q.IsRunning(3) {
		t.Fatalf("expected user to be released after error")
	}
}
