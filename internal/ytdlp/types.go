package ytdlp

import "fmt"

type MediaType string
type FormatNote string
type VCodec string
type ACodec string
type Ext string
type Container string
type YtDLPClient string

const (
	Video      MediaType = "video"
	Livestream MediaType = "livestream"
	Audio      MediaType = "audio"
	Short      MediaType = "short"

	Storyboard  FormatNote = "storyboard"
	AudioLow    FormatNote = "low"
	AudioMedium FormatNote = "medium"
	AudioHigh   FormatNote = "high"

	ClientDefault      YtDLPClient = "default"
	ClientWeb          YtDLPClient = "web"
	ClientWebEmbedded  YtDLPClient = "web_embedded"
	ClientWebSafari    YtDLPClient = "web_safari"
	ClientMWeb         YtDLPClient = "mweb"
	ClientWebMusic     YtDLPClient = "web_music"
	ClientWebCreator   YtDLPClient = "web_creator"
	ClientIOS          YtDLPClient = "ios"
	ClientAndroid      YtDLPClient = "android"
	ClientAndroidVR    YtDLPClient = "android_vr"
	ClientAndroidMusic YtDLPClient = "android_music"
	ClientTV           YtDLPClient = "tv"
	ClientTVDowngraded YtDLPClient = "tv_downgraded"
	ClientTVSimply     YtDLPClient = "tv_simply"
	ClientAll          YtDLPClient = "all"

	// Not sure that i will user this format notes, but i want to leave it here for maybe future use
	Video144p  FormatNote = "144p"
	Video240p  FormatNote = "240p"
	Video360p  FormatNote = "360p"
	Video480p  FormatNote = "480p"
	Video720p  FormatNote = "720p"
	Video1080p FormatNote = "1080p"
	Video1440p FormatNote = "1440p"
	Video2160p FormatNote = "2160p"

	Video144p60fps  FormatNote = "144p60"
	Video240p60fps  FormatNote = "240p60"
	Video360p60fps  FormatNote = "360p60"
	Video480p60fps  FormatNote = "480p60"
	Video720p60fps  FormatNote = "720p60"
	Video1080p60fps FormatNote = "1080p60"
	Video1440p60fps FormatNote = "1440p60"
	Video2160p60fps FormatNote = "2160p60"

	Video144pHDR  FormatNote = "144p HDR"
	Video240pHDR  FormatNote = "240p HDR"
	Video360pHDR  FormatNote = "360p HDR"
	Video480pHDR  FormatNote = "480p HDR"
	Video720pHDR  FormatNote = "720p HDR"
	Video1080pHDR FormatNote = "1080p HDR"
	Video1440pHDR FormatNote = "1440p HDR"
	Video2160pHDR FormatNote = "2160p HDR"

	Video144p60fpsHDR  FormatNote = "144p60 HDR"
	Video240p60fpsHDR  FormatNote = "240p60 HDR"
	Video360p60fpsHDR  FormatNote = "360p60 HDR"
	Video480p60fpsHDR  FormatNote = "480p60 HDR"
	Video720p60fpsHDR  FormatNote = "720p60 HDR"
	Video1080p60fpsHDR FormatNote = "1080p60 HDR"
	Video1440p60fpsHDR FormatNote = "1440p60 HDR"
	Video2160p60fpsHDR FormatNote = "2160p60 HDR"
)

type Metadata struct {
	ID                 string                `json:"id"`
	Title              string                `json:"title"`
	Thumbnail          string                `json:"thumbnail"`
	IsLive             bool                  `json:"is_live"`
	MediaType          MediaType             `json:"media_type"`
	OriginalURL        string                `json:"original_url"`
	Duration           int                   `json:"duration"`
	Formats            []Format              `json:"formats"`
	Subtitles          map[string][]Subtitle `json:"subtitles"`
	AutomaticCaptions  map[string][]Subtitle `json:"automatic_captions"`
	RequestedDownloads []RequestedDownloads  `json:"requested_downloads"`
}

type Subtitle struct {
	Ext         string      `json:"ext"`
	URL         string      `json:"url"`
	Name        string      `json:"name"`
	Impersonate bool        `json:"impersonate"`
	YtDLPClient YtDLPClient `json:"__yt_dlp_client"`
}

type HttpHeaders struct {
	UserAgent    string `json:"User-Agent"`
	Accept       string `json:"Accept"`
	AcceptLang   string `json:"Accept-Language"`
	SecFetchMode string `json:"Sec-Fetch-Mode"`
}

type RequestedDownloads struct {
	RequestedFormats []Format `json:"requested_formats"`
}

type Format struct {
	FormatID       string      `json:"format_id"`
	FormatNote     string      `json:"format_note"`
	FileSize       int64       `json:"filesize"`
	FileSizeApprox int64       `json:"filesize_approx"`
	Language       string      `json:"language"`
	LanguagePref   int         `json:"language_preference"`
	ACodec         string      `json:"acodec"`
	VCodec         string      `json:"vcodec"`
	Ext            string      `json:"ext"`
	Container      string      `json:"container"`
	Width          int         `json:"width"`
	Height         int         `json:"height"`
	FPS            float64     `json:"fps"`
	URL            string      `json:"url"`
	ABR            float64     `json:"abr"`
	VBR            float64     `json:"vbr"`
	Resolution     string      `json:"resolution"`
	HttpHeaders    HttpHeaders `json:"http_headers"`
}

func (f Format) IsVideo() bool {
	return f.VCodec != "none"
}

func (f Format) IsAudio() bool {
	return f.ACodec != "none"
}

func (f RequestedDownloads) GetBestAudioFormat() *Format {
	var bestAudio *Format
	for _, format := range f.RequestedFormats {
		if format.IsAudio() {
			bestAudio = &format
			break
		}
	}
	return bestAudio
}

func (f RequestedDownloads) GetBestVideoFormat() *Format {
	var bestVideo *Format
	for _, format := range f.RequestedFormats {
		if format.IsVideo() {
			bestVideo = &format
			break
		}
	}
	return bestVideo
}

func (f Format) GetRoundedABR() int {
	if f.ABR == 0 {
		return 0
	}
	return int(f.ABR + 0.5)
}

func (f Format) GetRoundedVBR() int {
	if f.VBR == 0 {
		return 0
	}
	return int(f.VBR + 0.5)
}

func (f Format) GetDisplayName(audioFormat, videoFormat *Format) string {
	if f.FormatNote != "" {
		return f.FormatNote
	}

	if f.IsAudio() && !f.IsVideo() {
		return fmt.Sprintf("Audio %dkbps", f.GetRoundedABR())
	}
	if f.IsVideo() && !f.IsAudio() {
		return fmt.Sprintf("Video %dx%d", f.Width, f.Height)
	}

	if f.IsAudio() && f.IsVideo() {
		return fmt.Sprintf("Video %dx%d + Audio %dkbps (muxed)", f.Width, f.Height, f.GetRoundedABR())
	}

	if audioFormat != nil && videoFormat != nil {
		return fmt.Sprintf("Video %dx%d + Audio %dkbps (merged)", videoFormat.Width, videoFormat.Height, audioFormat.GetRoundedABR())
	}
	return f.FormatID
}
