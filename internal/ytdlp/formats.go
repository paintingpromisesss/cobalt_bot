package ytdlp

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
