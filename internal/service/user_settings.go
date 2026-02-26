package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/repository"
)

type UserSettingsService struct {
	repo repository.UserSettingsRepository
}

func NewUserSettingsService(repo repository.UserSettingsRepository) (*UserSettingsService, error) {
	if repo == nil {
		return nil, fmt.Errorf("settings repository is nil")
	}
	return &UserSettingsService{repo: repo}, nil
}

func DefaultUserSettings() repository.UserSettings {
	return repository.UserSettings{
		VideoQuality:          "1080",
		DownloadMode:          "auto",
		AudioFormat:           "mp3",
		AudioBitrate:          "128",
		FilenameStyle:         "basic",
		YoutubeVideoCodec:     "h264",
		YoutubeVideoContainer: "auto",
		YoutubeBetterAudio:    false,
		SubtitleLang:          "",
	}
}

func (s *UserSettingsService) GetByUserID(ctx context.Context, userID int64) (repository.UserSettings, error) {
	out, found, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return repository.UserSettings{}, fmt.Errorf("get user settings: %w", err)
	}
	if !found {
		return DefaultUserSettings(), nil
	}
	return out, nil
}

func (s *UserSettingsService) UpsertByUserID(ctx context.Context, userID int64, settings repository.UserSettings) error {
	settings, err := normalizeAndValidate(settings)
	if err != nil {
		return err
	}

	if err := s.repo.UpsertByUserID(ctx, userID, settings); err != nil {
		return fmt.Errorf("upsert user settings: %w", err)
	}

	return nil
}

func normalizeAndValidate(settings repository.UserSettings) (repository.UserSettings, error) {
	s := settings
	if s.VideoQuality == "" {
		s.VideoQuality = "1080"
	}
	if s.DownloadMode == "" {
		s.DownloadMode = "auto"
	}
	if s.AudioFormat == "" {
		s.AudioFormat = "mp3"
	}
	if s.AudioBitrate == "" {
		s.AudioBitrate = "128"
	}
	if s.FilenameStyle == "" {
		s.FilenameStyle = "basic"
	}
	if s.YoutubeVideoCodec == "" {
		s.YoutubeVideoCodec = "h264"
	}
	if s.YoutubeVideoContainer == "" {
		s.YoutubeVideoContainer = "auto"
	}
	s.SubtitleLang = strings.TrimSpace(s.SubtitleLang)

	if !slices.Contains([]string{"best", "4320", "2160", "1440", "1080", "720", "480", "360", "240", "144"}, s.VideoQuality) {
		return repository.UserSettings{}, fmt.Errorf("invalid video quality: %q", s.VideoQuality)
	}
	if !slices.Contains([]string{"auto", "audio", "mute"}, s.DownloadMode) {
		return repository.UserSettings{}, fmt.Errorf("invalid download mode: %q", s.DownloadMode)
	}
	if !slices.Contains([]string{"best", "mp3", "ogg", "wav", "opus"}, s.AudioFormat) {
		return repository.UserSettings{}, fmt.Errorf("invalid audio format: %q", s.AudioFormat)
	}
	if !slices.Contains([]string{"best", "320", "256", "128", "96", "64", "8"}, s.AudioBitrate) {
		return repository.UserSettings{}, fmt.Errorf("invalid audio bitrate: %q", s.AudioBitrate)
	}
	if !slices.Contains([]string{"classic", "pretty", "basic", "nerdy"}, s.FilenameStyle) {
		return repository.UserSettings{}, fmt.Errorf("invalid filename style: %q", s.FilenameStyle)
	}
	if !slices.Contains([]string{"h264", "av1", "vp9"}, s.YoutubeVideoCodec) {
		return repository.UserSettings{}, fmt.Errorf("invalid youtube video codec: %q", s.YoutubeVideoCodec)
	}
	if !slices.Contains([]string{"auto", "mp4", "webm", "mkv"}, s.YoutubeVideoContainer) {
		return repository.UserSettings{}, fmt.Errorf("invalid youtube video container: %q", s.YoutubeVideoContainer)
	}
	if len(s.SubtitleLang) > 8 {
		return repository.UserSettings{}, fmt.Errorf("subtitle lang too long: %q", s.SubtitleLang)
	}

	return s, nil
}
