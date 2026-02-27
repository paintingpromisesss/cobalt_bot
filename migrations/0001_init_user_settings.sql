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
);
