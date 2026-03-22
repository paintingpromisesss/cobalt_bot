package cobalt

import "github.com/paintingpromisesss/cobalt_bot/internal/domain/user"

type MainRequest struct {
	Url                   string                `json:"url"`
	AudioBitrate          AudioBitrate          `json:"audioBitrate,omitempty"`
	AudioFormat           AudioFormat           `json:"audioFormat,omitempty"`
	DownloadMode          DownloadMode          `json:"downloadMode,omitempty"`
	FilenameStyle         FilenameStyle         `json:"filenameStyle,omitempty"`
	VideoQuality          VideoQuality          `json:"videoQuality,omitempty"`
	DisableMetadata       *bool                 `json:"disableMetadata,omitempty"`
	AlwaysProxy           *bool                 `json:"alwaysProxy,omitempty"`
	LocalProcessing       LocalProcessing       `json:"localProcessing,omitempty"`
	SubtitleLang          SubtitleLanguage      `json:"subtitleLang,omitempty"`
	YoutubeVideoCodec     YoutubeVideoCodec     `json:"youtubeVideoCodec,omitempty"`
	YoutubeVideoContainer YoutubeVideoContainer `json:"youtubeVideoContainer,omitempty"`
	YoutubeDubLang        SubtitleLanguage      `json:"youtubeDubLang,omitempty"`
	ConvertGif            *bool                 `json:"convertGif,omitempty"`
	AllowH265             *bool                 `json:"allowH265,omitempty"`
	TiktokFullAudio       *bool                 `json:"tiktokFullAudio,omitempty"`
	YoutubeBetterAudio    *bool                 `json:"youtubeBetterAudio,omitempty"`
	YoutubeHLS            *bool                 `json:"youtubeHLS,omitempty"`
}

func NewRequest(url string, settings user.Settings) MainRequest {
	return MainRequest{
		Url:          url,
		AudioBitrate: AudioBitrate(settings.AudioBitrate),
		AudioFormat:  AudioFormat(settings.AudioFormat),
		VideoQuality: VideoQuality(settings.VideoQuality),
	}
}
