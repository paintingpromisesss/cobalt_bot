package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func migrate(db *sql.DB) error {
	query, err := loadMigrationQuery()
	if err != nil {
		return err
	}

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

func loadMigrationQuery() (string, error) {
	candidates := []string{
		filepath.Join("migrations", "0001_init_user_settings.sql"),
		filepath.Join("..", "..", "migrations", "0001_init_user_settings.sql"),
		filepath.Join("..", "..", "..", "migrations", "0001_init_user_settings.sql"),
		filepath.Join("/app", "migrations", "0001_init_user_settings.sql"),
	}

	for _, p := range candidates {
		b, err := os.ReadFile(filepath.Clean(p))
		if err == nil {
			return string(b), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("read migration file %q: %w", p, err)
		}
	}

	return defaultMigrationV1, nil
}

const defaultMigrationV1 = `
CREATE TABLE IF NOT EXISTS user_settings (
  user_id INTEGER PRIMARY KEY,
  video_quality TEXT NOT NULL DEFAULT '1080',
  download_mode TEXT NOT NULL DEFAULT 'auto',
  audio_format TEXT NOT NULL DEFAULT 'mp3',
  audio_bitrate TEXT NOT NULL DEFAULT '128',
  filename_style TEXT NOT NULL DEFAULT 'basic',
  youtube_video_codec TEXT NOT NULL DEFAULT 'h264',
  youtube_video_container TEXT NOT NULL DEFAULT 'auto',
  youtube_better_audio INTEGER NOT NULL DEFAULT 0,
  subtitle_lang TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL
);`
