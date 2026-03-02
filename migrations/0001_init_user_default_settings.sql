CREATE TABLE IF NOT EXISTS user_default_settings (
	user_id INTEGER PRIMARY KEY,
	audio_bitrate TEXT NOT NULL DEFAULT '128' CHECK (audio_bitrate IN ('320', '256', '128', '96', '64', '8')),
	audio_format TEXT NOT NULL DEFAULT 'mp3' CHECK (audio_format IN ('best', 'mp3', 'ogg', 'wav', 'opus')),
	video_quality TEXT NOT NULL DEFAULT '1080' CHECK (video_quality IN ('max', '4320', '2160', '1440', '1080', '720', '480', '360', '240', '144')),
	subtitle_lang TEXT,
	youtube_dub_lang TEXT
);
