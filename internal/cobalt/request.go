package cobalt

import "github.com/paintingpromisesss/cobalt_bot/internal/domain/user"

// MainRequest matches cobalt POST / request body.
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
	YoutubeVideoCodec     YoutubeVideoCodec     `json:"youtubeVideoCodec,omitempty"`     // youtube only
	YoutubeVideoContainer YoutubeVideoContainer `json:"youtubeVideoContainer,omitempty"` // youtube only
	YoutubeDubLang        SubtitleLanguage      `json:"youtubeDubLang,omitempty"`        // youtube only
	ConvertGif            *bool                 `json:"convertGif,omitempty"`            // twitter only
	AllowH265             *bool                 `json:"allowH265,omitempty"`             // tiktok/xiaohongshu only
	TiktokFullAudio       *bool                 `json:"tiktokFullAudio,omitempty"`       // tiktok only
	YoutubeBetterAudio    *bool                 `json:"youtubeBetterAudio,omitempty"`    // youtube only
	YoutubeHLS            *bool                 `json:"youtubeHLS,omitempty"`            // youtube only
}

func GetCobaltRequest(url string, settings user.Settings) MainRequest {
	return MainRequest{
		Url:          url,
		AudioBitrate: AudioBitrate(settings.AudioBitrate),
		AudioFormat:  AudioFormat(settings.AudioFormat),
		VideoQuality: VideoQuality(settings.VideoQuality),
	}
}
