package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/user"
)

var ErrUserSettingsNotFound = errors.New("user settings not found")

type userSettingsRow struct {
	UserID         int64          `db:"user_id"`
	AudioBitrate   string         `db:"audio_bitrate"`
	AudioFormat    string         `db:"audio_format"`
	VideoQuality   string         `db:"video_quality"`
	SubtitleLang   sql.NullString `db:"subtitle_lang"`
	YoutubeDubLang sql.NullString `db:"youtube_dub_lang"`
}

func (d *DB) GetUserSettings(ctx context.Context, userID int64) (user.Settings, error) {
	const query = `
SELECT
	user_id,
	audio_bitrate,
	audio_format,
	video_quality,
	subtitle_lang,
	youtube_dub_lang
FROM user_default_settings
WHERE user_id = ?;
`

	var row userSettingsRow
	err := d.sqlDB.GetContext(ctx, &row, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.Settings{}, ErrUserSettingsNotFound
		}
		return user.Settings{}, fmt.Errorf("get user settings by user_id=%d: %w", userID, err)
	}

	return row.toUserSettings(), nil
}

func (d *DB) UpsertUserSettings(ctx context.Context, settings user.Settings) error {
	const query = `
INSERT INTO user_default_settings (
	user_id,
	audio_bitrate,
	audio_format,
	video_quality,
	subtitle_lang,
	youtube_dub_lang
)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
	audio_bitrate = excluded.audio_bitrate,
	audio_format = excluded.audio_format,
	video_quality = excluded.video_quality,
	subtitle_lang = excluded.subtitle_lang,
	youtube_dub_lang = excluded.youtube_dub_lang;
`

	_, err := d.sqlDB.ExecContext(
		ctx,
		query,
		settings.UserID,
		settings.AudioBitrate,
		settings.AudioFormat,
		settings.VideoQuality,
		stringToNullString(settings.SubtitleLang),
		stringToNullString(settings.YoutubeDubLang),
	)
	if err != nil {
		return fmt.Errorf("upsert user settings by user_id=%d: %w", settings.UserID, err)
	}

	return nil
}

func (d *DB) EnsureUserSettings(ctx context.Context, userID int64) error {
	const query = `
	INSERT INTO user_default_settings (user_id)
	VALUES (?)
	ON CONFLICT(user_id) DO NOTHING;`
	_, err := d.sqlDB.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("ensure user settings by user_id=%d: %w", userID, err)
	}
	return nil
}

func (d *DB) DeleteUserSettings(ctx context.Context, userID int64) error {
	const query = `DELETE FROM user_default_settings WHERE user_id = ?;`

	result, err := d.sqlDB.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete user settings by user_id=%d: %w", userID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows for delete user settings by user_id=%d: %w", userID, err)
	}
	if rowsAffected == 0 {
		return ErrUserSettingsNotFound
	}

	return nil
}

func (r userSettingsRow) toUserSettings() user.Settings {
	return user.Settings{
		UserID:         r.UserID,
		AudioBitrate:   r.AudioBitrate,
		AudioFormat:    r.AudioFormat,
		VideoQuality:   r.VideoQuality,
		SubtitleLang:   nullStringToSubtitleLanguage(r.SubtitleLang),
		YoutubeDubLang: nullStringToSubtitleLanguage(r.YoutubeDubLang),
	}
}

func stringToNullString(v *string) sql.NullString {
	if v == nil || *v == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *v,
		Valid:  true,
	}
}

func nullStringToSubtitleLanguage(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}

	lang := v.String
	return &lang
}
