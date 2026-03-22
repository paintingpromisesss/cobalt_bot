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

func (s *YtDLPState) SelectTab(tab YtDLPTab) error {
	if tab != YtDLPTabNone {
		options, ok := s.OptionsByTab[tab]
		if !ok || len(options) == 0 {
			return ErrInvalidYtDLPTab
		}
	}

	s.ActiveTab = tab
	if s.HasChosen && s.ChosenTab != tab {
		s.ResetChoice()
	}

	return nil
}

func (s *YtDLPState) ChooseActiveOption(idx int) (YtDLPOption, error) {
	options := s.OptionsByTab[s.ActiveTab]
	if idx < 0 || idx >= len(options) {
		return YtDLPOption{}, ErrInvalidOptionIdx
	}

	s.ChosenTab = s.ActiveTab
	s.ChosenIndex = idx
	s.HasChosen = true

	return options[idx], nil
}

func (s *YtDLPState) ResetChoice() {
	s.HasChosen = false
	s.ChosenTab = YtDLPTabNone
	s.ChosenIndex = -1
}

func (s YtDLPState) SelectedOption() (YtDLPOption, error) {
	if !s.HasChosen {
		return YtDLPOption{}, ErrNoOptionsSelected
	}

	options := s.OptionsByTab[s.ChosenTab]
	if s.ChosenIndex < 0 || s.ChosenIndex >= len(options) {
		return YtDLPOption{}, ErrInvalidOptionIdx
	}

	return options[s.ChosenIndex], nil
}
