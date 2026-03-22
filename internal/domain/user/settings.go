package user

type Settings struct {
	UserID         int64
	AudioBitrate   string
	AudioFormat    string
	VideoQuality   string
	SubtitleLang   *string
	YoutubeDubLang *string
}

func DefaultSettings() Settings {
	return Settings{
		AudioBitrate: "128",
		AudioFormat:  "mp3",
		VideoQuality: "1080",
	}
}
