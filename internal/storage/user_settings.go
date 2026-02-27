package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/user_settings"
)

type UserSettingsRepository struct {
	db *sql.DB
}

func NewUserSettingsRepository(db *sql.DB) (*UserSettingsRepository, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	return &UserSettingsRepository{db: db}, nil
}

func (r *UserSettingsRepository) GetByUserID(ctx context.Context, userID int64) (user_settings.Settings, bool, error) {
	const query = `
SELECT
  video_quality,
  download_mode,
  audio_format,
  audio_bitrate,
  filename_style,
  youtube_video_codec,
  youtube_video_container,
  youtube_better_audio,
  subtitle_lang
FROM user_settings
WHERE user_id = ?;`

	var out user_settings.Settings
	var betterAudio int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&out.VideoQuality,
		&out.DownloadMode,
		&out.AudioFormat,
		&out.AudioBitrate,
		&out.FilenameStyle,
		&out.YoutubeVideoCodec,
		&out.YoutubeVideoContainer,
		&betterAudio,
		&out.SubtitleLang,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return user_settings.Settings{}, false, nil
	}
	if err != nil {
		return user_settings.Settings{}, false, fmt.Errorf("select user settings: %w", err)
	}

	out.YoutubeBetterAudio = betterAudio == 1
	return out, true, nil
}

func (r *UserSettingsRepository) UpsertByUserID(ctx context.Context, userID int64, settings user_settings.Settings) error {
	const query = `
INSERT INTO user_settings (
  user_id,
  video_quality,
  download_mode,
  audio_format,
  audio_bitrate,
  filename_style,
  youtube_video_codec,
  youtube_video_container,
  youtube_better_audio,
  subtitle_lang,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
  video_quality = excluded.video_quality,
  download_mode = excluded.download_mode,
  audio_format = excluded.audio_format,
  audio_bitrate = excluded.audio_bitrate,
  filename_style = excluded.filename_style,
  youtube_video_codec = excluded.youtube_video_codec,
  youtube_video_container = excluded.youtube_video_container,
  youtube_better_audio = excluded.youtube_better_audio,
  subtitle_lang = excluded.subtitle_lang,
  updated_at = excluded.updated_at;`

	betterAudio := 0
	if settings.YoutubeBetterAudio {
		betterAudio = 1
	}

	updatedAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := r.db.ExecContext(
		ctx,
		query,
		userID,
		settings.VideoQuality,
		settings.DownloadMode,
		settings.AudioFormat,
		settings.AudioBitrate,
		settings.FilenameStyle,
		settings.YoutubeVideoCodec,
		settings.YoutubeVideoContainer,
		betterAudio,
		settings.SubtitleLang,
		updatedAt,
	); err != nil {
		return fmt.Errorf("upsert user settings: %w", err)
	}

	return nil
}
