package cobalt

import (
	"encoding/json"
	"fmt"
)

type AudioBitrate string
type AudioFormat string
type DownloadMode string
type FilenameStyle string
type VideoQuality string
type LocalProcessing string
type SubtitleLanguage string
type YoutubeVideoCodec string
type YoutubeVideoContainer string

type Status string
type LocalProcessingType string
type LocalProcessingService string
type PickerType string

const (
	Bitrate320 AudioBitrate = "320"
	Bitrate256 AudioBitrate = "256"
	Bitrate128 AudioBitrate = "128"
	Bitrate96  AudioBitrate = "96"
	Bitrate64  AudioBitrate = "64"
	Bitrate8   AudioBitrate = "8"

	FormatBest AudioFormat = "best"
	FormatMP3  AudioFormat = "mp3"
	FormatOGG  AudioFormat = "ogg"
	FormatWAV  AudioFormat = "wav"
	FormatOPUS AudioFormat = "opus"

	ModeAuto  DownloadMode = "auto"
	ModeAudio DownloadMode = "audio"
	ModeMute  DownloadMode = "mute"

	StyleClassic FilenameStyle = "classic"
	StylePretty  FilenameStyle = "pretty"
	StyleBasic   FilenameStyle = "basic"
	StyleNerdy   FilenameStyle = "nerdy"

	QualityMax  VideoQuality = "max"
	Quality4320 VideoQuality = "4320"
	Quality2160 VideoQuality = "2160"
	Quality1440 VideoQuality = "1440"
	Quality1080 VideoQuality = "1080"
	Quality720  VideoQuality = "720"
	Quality480  VideoQuality = "480"
	Quality360  VideoQuality = "360"
	Quality240  VideoQuality = "240"
	Quality144  VideoQuality = "144"

	ProcessingDisabled  LocalProcessing = "disabled"
	ProcessingPreferred LocalProcessing = "preferred"
	ProcessingForced    LocalProcessing = "forced"

	YoutubeCodecH264 YoutubeVideoCodec = "h264"
	YoutubeCodecAV1  YoutubeVideoCodec = "av1"
	YoutubeCodecVP9  YoutubeVideoCodec = "vp9"

	YoutubeContainerAuto YoutubeVideoContainer = "auto"
	YoutubeContainerMP4  YoutubeVideoContainer = "mp4"
	YoutubeContainerWEBM YoutubeVideoContainer = "webm"
	YoutubeContainerMKV  YoutubeVideoContainer = "mkv"

	StatusTunnel          Status = "tunnel"
	StatusLocalProcessing Status = "local-processing"
	StatusRedirect        Status = "redirect"
	StatusPicker          Status = "picker"
	StatusError           Status = "error"

	LocalProcessingMerge LocalProcessingType = "merge"
	LocalProcessingMute  LocalProcessingType = "mute"
	LocalProcessingAudio LocalProcessingType = "audio"
	LocalProcessingGif   LocalProcessingType = "gif"
	LocalProcessingRemux LocalProcessingType = "remux"

	LocalProcessingServiceBilibili    LocalProcessingService = "bilibili"
	LocalProcessingServiceBluesky     LocalProcessingService = "bluesky"
	LocalProcessingServiceDailymotion LocalProcessingService = "dailymotion"
	LocalProcessingServiceInstagram   LocalProcessingService = "instagram"
	LocalProcessingServiceFacebook    LocalProcessingService = "facebook"
	LocalProcessingServiceLoom        LocalProcessingService = "loom"
	LocalProcessingServiceNewgrounds  LocalProcessingService = "newgrounds"
	LocalProcessingServiceOkRu        LocalProcessingService = "ok"
	LocalProcessingServicePinterest   LocalProcessingService = "pinterest"
	LocalProcessingServiceReddit      LocalProcessingService = "reddit"
	LocalProcessingServiceRutube      LocalProcessingService = "rutube"
	LocalProcessingServiceSnapchat    LocalProcessingService = "snapchat"
	LocalProcessingServiceSoundcloud  LocalProcessingService = "soundcloud"
	LocalProcessingServiceStreamable  LocalProcessingService = "streamable"
	LocalProcessingServiceTiktok      LocalProcessingService = "tiktok"
	LocalProcessingServiceTumblr      LocalProcessingService = "tumblr"
	LocalProcessingServiceTwitch      LocalProcessingService = "twitch"
	LocalProcessingServiceTwitter     LocalProcessingService = "twitter"
	LocalProcessingServiceVimeo       LocalProcessingService = "vimeo"
	LocalProcessingServiceVk          LocalProcessingService = "vk"
	LocalProcessingServiceXiaohongshu LocalProcessingService = "xiaohongshu"
	LocalProcessingServiceYoutube     LocalProcessingService = "youtube"

	PickerTypePhoto PickerType = "photo"
	PickerTypeVideo PickerType = "video"
	PickerTypeGif   PickerType = "gif"
)

type CobaltObject struct {
	Version          string   `json:"version"`
	Url              string   `json:"url"`
	StartTime        string   `json:"startTime"`
	TurnstileSitekey string   `json:"turnstileSitekey,omitempty"`
	Services         []string `json:"services"`
}

type MetadataObject struct {
	Album       string `json:"album,omitempty"`
	Composer    string `json:"composer,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Copyright   string `json:"copyright,omitempty"`
	Title       string `json:"title,omitempty"`
	Artist      string `json:"artist,omitempty"`
	AlbumArtist string `json:"album_artist,omitempty"`
	Track       string `json:"track,omitempty"`
	Date        string `json:"date,omitempty"`
	Sublanguage string `json:"sublanguage,omitempty"`
}

type OutputObject struct {
	Type      string          `json:"type"`
	Filename  string          `json:"filename"`
	Metadata  *MetadataObject `json:"metadata,omitempty"`
	Subtitles bool            `json:"subtitles"`
}

type AudioLocalProcessingObject struct {
	Copy      bool   `json:"copy"`
	Format    string `json:"format"`
	Bitrate   string `json:"bitrate"`
	Cover     bool   `json:"cover,omitempty"`
	CropCover bool   `json:"cropCover,omitempty"`
}

type PickerObject struct {
	Type  PickerType `json:"type"`
	Url   string     `json:"url"`
	Thumb string     `json:"thumb,omitempty"`
}

type ErrorObject struct {
	Code    string              `json:"code"`
	Context *ErrorContextObject `json:"context,omitempty"`
}

type ErrorContextObject struct {
	Service string  `json:"service,omitempty"`
	Limit   float64 `json:"limit,omitempty"`
}

type GitObject struct {
	Commit string `json:"commit"`
	Branch string `json:"branch"`
	Remote string `json:"remote"`
}

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

type MainResponse interface {
	GetStatus() Status
}

type MainResponseEnvelope struct {
	Status Status `json:"status"`
}

type TunnelOrRedirectResponse struct {
	Status   Status `json:"status"`
	Url      string `json:"url"`
	Filename string `json:"filename"`
}

func (r TunnelOrRedirectResponse) GetStatus() Status { return r.Status }

type LocalProcessingResponse struct {
	Status  Status                      `json:"status"`
	Type    LocalProcessingType         `json:"type"`
	Service LocalProcessingService      `json:"service"`
	Tunnel  []string                    `json:"tunnel"`
	Output  OutputObject                `json:"output"`
	Audio   *AudioLocalProcessingObject `json:"audio,omitempty"`
	IsHLS   *bool                       `json:"isHLS,omitempty"`
}

func (r LocalProcessingResponse) GetStatus() Status { return r.Status }

type PickerResponse struct {
	Status        Status         `json:"status"`
	Audio         *string        `json:"audio,omitempty"`
	AudioFilename *string        `json:"audioFilename,omitempty"`
	Picker        []PickerObject `json:"picker"`
}

func (r PickerResponse) GetStatus() Status { return r.Status }

type ErrorResponse struct {
	Status Status      `json:"status"`
	Error  ErrorObject `json:"error"`
}

func (r ErrorResponse) GetStatus() Status { return r.Status }

func ParseMainResponse(data []byte) (MainResponse, error) {
	var envelope MainResponseEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode response envelope: %w", err)
	}

	switch envelope.Status {
	case StatusTunnel, StatusRedirect:
		var r TunnelOrRedirectResponse
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("decode %q response: %w", envelope.Status, err)
		}
		return r, nil
	case StatusLocalProcessing:
		var r LocalProcessingResponse
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("decode %q response: %w", envelope.Status, err)
		}
		return r, nil
	case StatusPicker:
		var r PickerResponse
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("decode %q response: %w", envelope.Status, err)
		}
		return r, nil
	case StatusError:
		var r ErrorResponse
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("decode %q response: %w", envelope.Status, err)
		}
		return r, nil
	default:
		return nil, fmt.Errorf("unsupported response status %q", envelope.Status)
	}
}

type InstanceResponse struct {
	Cobalt CobaltObject `json:"cobalt"`
	Git    GitObject    `json:"git"`
}
