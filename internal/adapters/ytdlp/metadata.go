package ytdlp

type Metadata struct {
	ID                 string                `json:"id"`
	Title              string                `json:"title"`
	Thumbnail          string                `json:"thumbnail"`
	IsLive             bool                  `json:"is_live"`
	MediaType          string                `json:"media_type"`
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
