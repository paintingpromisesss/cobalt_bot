package picker

import (
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
)

type YtDLPTab string

const (
	YtDLPTabAudioOnly  YtDLPTab = "audio_only"
	YtDLPTabVideoOnly  YtDLPTab = "video_only"
	YtDLPTabAudioVideo YtDLPTab = "audio_video"
	YtDLPTabNone       YtDLPTab = "none"
	YtDLPTabSubtitles  YtDLPTab = "subtitles"
)

type YtDLPOption struct {
	DisplayName  string
	FormatID     string
	ThumbnailURL string
	ContentURL   string
	FileSize     int64
	Duration     time.Duration
	Format       media.DownloadFormat
}

type YtDLPState struct {
	ContentName string

	ActiveTab    YtDLPTab
	OptionsByTab map[YtDLPTab][]YtDLPOption

	ChosenTab   YtDLPTab
	ChosenIndex int
	HasChosen   bool
}

type YtDLPView struct {
	ContentName string
	ActiveTab   YtDLPTab
	Tabs        []YtDLPTab
	Options     []YtDLPOption
}
