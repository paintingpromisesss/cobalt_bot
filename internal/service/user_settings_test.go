package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/paintingpromisesss/cobalt_bot/internal/repository"
)

type mockUserSettingsRepo struct {
	getFn      func(ctx context.Context, userID int64) (repository.UserSettings, bool, error)
	upsertFn   func(ctx context.Context, userID int64, settings repository.UserSettings) error
	lastUpsert repository.UserSettings
	lastUserID int64
}

func (m *mockUserSettingsRepo) GetByUserID(ctx context.Context, userID int64) (repository.UserSettings, bool, error) {
	if m.getFn == nil {
		return repository.UserSettings{}, false, nil
	}
	return m.getFn(ctx, userID)
}

func (m *mockUserSettingsRepo) UpsertByUserID(ctx context.Context, userID int64, settings repository.UserSettings) error {
	m.lastUserID = userID
	m.lastUpsert = settings
	if m.upsertFn == nil {
		return nil
	}
	return m.upsertFn(ctx, userID, settings)
}

func TestNewUserSettingsServiceNilRepo(t *testing.T) {
	_, err := NewUserSettingsService(nil)
	if err == nil {
		t.Fatalf("expected error for nil repo")
	}
}

func TestUserSettingsServiceGetByUserIDReturnsDefaultsWhenNotFound(t *testing.T) {
	repo := &mockUserSettingsRepo{
		getFn: func(ctx context.Context, userID int64) (repository.UserSettings, bool, error) {
			return repository.UserSettings{}, false, nil
		},
	}
	svc, err := NewUserSettingsService(repo)
	if err != nil {
		t.Fatalf("init service: %v", err)
	}

	got, err := svc.GetByUserID(context.Background(), 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := DefaultUserSettings()
	if got != want {
		t.Fatalf("unexpected defaults: got=%+v want=%+v", got, want)
	}
}

func TestUserSettingsServiceGetByUserIDPropagatesRepoError(t *testing.T) {
	expectedErr := errors.New("db down")
	repo := &mockUserSettingsRepo{
		getFn: func(ctx context.Context, userID int64) (repository.UserSettings, bool, error) {
			return repository.UserSettings{}, false, expectedErr
		},
	}
	svc, _ := NewUserSettingsService(repo)

	_, err := svc.GetByUserID(context.Background(), 1)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}

func TestUserSettingsServiceUpsertByUserIDNormalizesFields(t *testing.T) {
	repo := &mockUserSettingsRepo{}
	svc, _ := NewUserSettingsService(repo)

	err := svc.UpsertByUserID(context.Background(), 42, repository.UserSettings{
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

func TestUserSettingsServiceUpsertByUserIDRejectsInvalidFields(t *testing.T) {
	repo := &mockUserSettingsRepo{}
	svc, _ := NewUserSettingsService(repo)

	err := svc.UpsertByUserID(context.Background(), 1, repository.UserSettings{
		VideoQuality: "9999",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid video quality") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserSettingsServiceUpsertByUserIDRejectsLongSubtitle(t *testing.T) {
	repo := &mockUserSettingsRepo{}
	svc, _ := NewUserSettingsService(repo)

	err := svc.UpsertByUserID(context.Background(), 1, repository.UserSettings{
		SubtitleLang: "123456789",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "subtitle lang too long") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserSettingsServiceUpsertByUserIDPropagatesRepoError(t *testing.T) {
	expectedErr := errors.New("write failed")
	repo := &mockUserSettingsRepo{
		upsertFn: func(ctx context.Context, userID int64, settings repository.UserSettings) error {
			return expectedErr
		},
	}
	svc, _ := NewUserSettingsService(repo)

	err := svc.UpsertByUserID(context.Background(), 1, repository.UserSettings{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}
