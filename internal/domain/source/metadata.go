package source

type CobaltStatus string

const (
	CobaltStatusTunnel   CobaltStatus = "tunnel"
	CobaltStatusRedirect CobaltStatus = "redirect"
	CobaltStatusPicker   CobaltStatus = "picker"
	CobaltStatusError    CobaltStatus = "error"
)

type CobaltError struct {
	Code    string
	Service string
	Limit   float64
}

type CobaltContent struct {
	Status   CobaltStatus
	FileURL  string
	FileName string
	Options  []CobaltOption
	Error    *CobaltError
}

type CobaltOption struct {
	Label    string
	URL      string
	Filename string
}

type YtDLPMetadata struct {
	Title              string
	ThumbnailURL       string
	OriginalURL        string
	DurationSeconds    int
	Formats            []YtDLPFormat
	RequestedDownloads []YtDLPRequestedDownload
}

type YtDLPRequestedDownload struct {
	Formats []YtDLPFormat
}

type YtDLPFormat struct {
	FormatID    string
	DisplayName string
	FileSize    int64
	HasAudio    bool
	HasVideo    bool
}

func (f YtDLPRequestedDownload) GetBestAudioFormat() *YtDLPFormat {
	for i := range f.Formats {
		if f.Formats[i].HasAudio {
			return &f.Formats[i]
		}
	}
	return nil
}

func (f YtDLPRequestedDownload) GetBestVideoFormat() *YtDLPFormat {
	for i := range f.Formats {
		if f.Formats[i].HasVideo {
			return &f.Formats[i]
		}
	}
	return nil
}
