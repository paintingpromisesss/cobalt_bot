package user_settings

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockRepo struct {
	getFn      func(ctx context.Context, userID int64) (Settings, bool, error)
	upsertFn   func(ctx context.Context, userID int64, settings Settings) error
	lastUpsert Settings
	lastUserID int64
}

func (m *mockRepo) GetByUserID(ctx context.Context, userID int64) (Settings, bool, error) {
	if m.getFn == nil {
		return Settings{}, false, nil
	}
	return m.getFn(ctx, userID)
}

func (m *mockRepo) UpsertByUserID(ctx context.Context, userID int64, settings Settings) error {
	m.lastUserID = userID
	m.lastUpsert = settings
	if m.upsertFn == nil {
		return nil
	}
	return m.upsertFn(ctx, userID, settings)
}

func TestNewServiceNilRepo(t *testing.T) {
	_, err := NewService(nil)
	if err == nil {
		t.Fatalf("expected error for nil repo")
	}
}

func TestServiceGetByUserIDReturnsDefaultsWhenNotFound(t *testing.T) {
	repo := &mockRepo{
		getFn: func(ctx context.Context, userID int64) (Settings, bool, error) {
			return Settings{}, false, nil
		},
	}
	svc, err := NewService(repo)
	if err != nil {
		t.Fatalf("init service: %v", err)
	}

	got, err := svc.GetByUserID(context.Background(), 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := DefaultSettings()
	if got != want {
		t.Fatalf("unexpected defaults: got=%+v want=%+v", got, want)
	}
}

func TestServiceGetByUserIDPropagatesRepoError(t *testing.T) {
	expectedErr := errors.New("db down")
	repo := &mockRepo{
		getFn: func(ctx context.Context, userID int64) (Settings, bool, error) {
			return Settings{}, false, expectedErr
		},
	}
	svc, _ := NewService(repo)

	_, err := svc.GetByUserID(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}

func TestServiceUpsertByUserIDNormalizesFields(t *testing.T) {
	repo := &mockRepo{}
	svc, _ := NewService(repo)

	err := svc.UpsertByUserID(context.Background(), 42, Settings{
		SubtitleLang: "  ru  ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.lastUserID != 42 {
		t.Fatalf("unexpected user id: %d", repo.lastUserID)
	}

	got := repo.lastUpsert
	if got.VideoQuality != "1080" ||
		got.DownloadMode != "auto" ||
		got.AudioFormat != "mp3" ||
		got.AudioBitrate != "128" ||
		got.FilenameStyle != "basic" ||
		got.YoutubeVideoCodec != "h264" ||
		got.YoutubeVideoContainer != "auto" ||
		got.SubtitleLang != "ru" {
		t.Fatalf("unexpected normalized settings: %+v", got)
	}
}

func TestServiceUpsertByUserIDRejectsInvalidFields(t *testing.T) {
	repo := &mockRepo{}
	svc, _ := NewService(repo)

	err := svc.UpsertByUserID(context.Background(), 1, Settings{
		VideoQuality: "9999",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid video quality") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceUpsertByUserIDRejectsLongSubtitle(t *testing.T) {
	repo := &mockRepo{}
	svc, _ := NewService(repo)

	err := svc.UpsertByUserID(context.Background(), 1, Settings{
		SubtitleLang: "123456789",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "subtitle lang too long") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceUpsertByUserIDPropagatesRepoError(t *testing.T) {
	expectedErr := errors.New("write failed")
	repo := &mockRepo{
		upsertFn: func(ctx context.Context, userID int64, settings Settings) error {
			return expectedErr
		},
	}
	svc, _ := NewService(repo)

	err := svc.UpsertByUserID(context.Background(), 1, Settings{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}
