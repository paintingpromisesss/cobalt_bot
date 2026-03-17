package pickersession

import (
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
)

type YtDLPPickerTab string

const (
	YtDLPPickerTabAudioOnly  YtDLPPickerTab = "audio_only"
	YtDLPPickerTabVideoOnly  YtDLPPickerTab = "video_only"
	YtDLPPickerTabAudioVideo YtDLPPickerTab = "audio_video"
	YtDLPPickerTabNone       YtDLPPickerTab = "none"
	YtDLPPickerTabSubtitles  YtDLPPickerTab = "subtitles"
)

type YtDLPPickerOption struct {
	DisplayName  string
	FormatID     string
	ThumbnailURL string
	ContentURL   string
	FileSize     int64
	Duration     time.Duration
	Format       ytdlp.Format
}

type YtDLPPickerState struct {
	ContentName string

	ActiveTab    YtDLPPickerTab
	OptionsByTab map[YtDLPPickerTab][]YtDLPPickerOption

	ChosenTab   YtDLPPickerTab
	ChosenIndex int
	HasChosen   bool
}

type YtDLPPickerView struct {
	ContentName string
	ActiveTab   YtDLPPickerTab
	Tabs        []YtDLPPickerTab
	Options     []YtDLPPickerOption
}

type YtDLPURLType string

const (
	YtDLPURLTypeAudio YtDLPURLType = "audio"
	YtDLPURLTypeVideo YtDLPURLType = "video"
	YtDLPURLTypeMuxed YtDLPURLType = "muxed"
)

func (m *PickerSessionManager) CreateYtDLPSession(userID int64, metadata *ytdlp.Metadata) (string, error) {
	optsByTab := ParseYtDLPMetadata(metadata)

	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := m.newUniqueSessionIDLocked()
	if err != nil {
		return "", err
	}

	m.sessions[id] = &pickerSession{
		sessionType: PickerSessionTypeYtDLP,
		userID:      userID,
		ytdlp: &YtDLPPickerState{
			ContentName:  metadata.Title,
			ActiveTab:    YtDLPPickerTabNone,
			OptionsByTab: optsByTab,
		},
		expiresAt: time.Now().Add(m.ttl),
	}

	return id, nil
}

func (m *PickerSessionManager) GetYtDLPPickerView(sessionID string, userID int64) (YtDLPPickerView, error) {
	return m.withYtDLPSessionView(sessionID, userID, func(s *pickerSession) error {
		return nil
	})
}

func (m *PickerSessionManager) SetYtDLPActiveTab(sessionID string, userID int64, tab YtDLPPickerTab) (YtDLPPickerView, error) {
	return m.withYtDLPSessionView(sessionID, userID, func(s *pickerSession) error {
		if tab == YtDLPPickerTabNone {
			s.ytdlp.ActiveTab = tab
			return nil
		}
		options, ok := s.ytdlp.OptionsByTab[tab]
		if !ok || len(options) == 0 {
			return ErrInvalidYtDLPTab
		}
		s.ytdlp.ActiveTab = tab

		if s.ytdlp.HasChosen && s.ytdlp.ChosenTab != tab {
			s.ytdlp.HasChosen = false
			s.ytdlp.ChosenTab = YtDLPPickerTabNone
			s.ytdlp.ChosenIndex = -1
		}
		return nil
	})
}

func (m *PickerSessionManager) ChooseYtDLPOption(sessionID string, userID int64, optionIdx int) (YtDLPPickerOption, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeYtDLP)
	if err != nil {
		return YtDLPPickerOption{}, err
	}

	options := s.ytdlp.OptionsByTab[s.ytdlp.ActiveTab]
	if optionIdx < 0 || optionIdx >= len(options) {
		return YtDLPPickerOption{}, ErrInvalidOptionIdx
	}

	s.ytdlp.ChosenTab = s.ytdlp.ActiveTab
	s.ytdlp.ChosenIndex = optionIdx
	s.ytdlp.HasChosen = true

	return options[optionIdx], nil
}

func (m *PickerSessionManager) ClearChosenYtDLPOption(sessionID string, userID int64) (YtDLPPickerView, error) {
	return m.withYtDLPSessionView(sessionID, userID, func(s *pickerSession) error {
		s.ytdlp.HasChosen = false
		s.ytdlp.ChosenTab = YtDLPPickerTabNone
		s.ytdlp.ChosenIndex = -1
		return nil
	})
}

func (m *PickerSessionManager) ConsumeChosenYtDLPOption(sessionID string, userID int64) (YtDLPPickerOption, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeYtDLP)
	if err != nil {
		return YtDLPPickerOption{}, err
	}

	if !s.ytdlp.HasChosen {
		return YtDLPPickerOption{}, ErrNoOptionsSelected
	}

	options := s.ytdlp.OptionsByTab[s.ytdlp.ChosenTab]
	if s.ytdlp.ChosenIndex < 0 || s.ytdlp.ChosenIndex >= len(options) {
		return YtDLPPickerOption{}, ErrInvalidOptionIdx
	}

	chosenOption := options[s.ytdlp.ChosenIndex]

	delete(m.sessions, sessionID)

	return chosenOption, nil
}

func (m *PickerSessionManager) withYtDLPSessionView(sessionID string, userID int64, fn func(*pickerSession) error) (YtDLPPickerView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeYtDLP)
	if err != nil {
		return YtDLPPickerView{}, err
	}

	if err := fn(s); err != nil {
		return YtDLPPickerView{}, err
	}
	return buildYtDLPPickerView(s), nil
}

func buildYtDLPPickerView(s *pickerSession) YtDLPPickerView {
	sourceOptions := s.ytdlp.OptionsByTab[s.ytdlp.ActiveTab]
	options := make([]YtDLPPickerOption, len(sourceOptions))
	copy(options, sourceOptions)

	tabs := make([]YtDLPPickerTab, 0, len(s.ytdlp.OptionsByTab))
	for _, tab := range orderedYtDLPTabs() {
		if len(s.ytdlp.OptionsByTab[tab]) > 0 {
			tabs = append(tabs, tab)
		}
	}

	return YtDLPPickerView{
		ContentName: s.ytdlp.ContentName,
		ActiveTab:   s.ytdlp.ActiveTab,
		Tabs:        tabs,
		Options:     options,
	}
}

func ParseYtDLPMetadata(metadata *ytdlp.Metadata) map[YtDLPPickerTab][]YtDLPPickerOption {
	optsByTab := make(map[YtDLPPickerTab][]YtDLPPickerOption)
	thumbnailURL := metadata.Thumbnail
	originalURL := metadata.OriginalURL

	for _, format := range metadata.Formats {
		tab := detectTabForFormat(format)
		if tab == "" {
			continue
		}

		option := YtDLPPickerOption{
			DisplayName:  format.GetDisplayName(nil, nil),
			ThumbnailURL: thumbnailURL,
			ContentURL:   originalURL,
			FormatID:     format.FormatID,
			FileSize:     format.FileSize,
			Duration:     time.Duration(metadata.Duration) * time.Second,
			Format:       format,
		}

		optsByTab[tab] = append(optsByTab[tab], option)
	}

	if len(metadata.RequestedDownloads) == 0 {
		return optsByTab
	}

	bestAudioFormat := metadata.RequestedDownloads[0].GetBestAudioFormat()
	if bestAudioFormat != nil {
		for _, format := range optsByTab[YtDLPPickerTabVideoOnly] {
			option := YtDLPPickerOption{
				DisplayName:  format.Format.GetDisplayName(bestAudioFormat, &format.Format),
				ContentURL:   originalURL,
				FormatID:     format.FormatID + "+" + bestAudioFormat.FormatID,
				ThumbnailURL: thumbnailURL,
				FileSize:     bestAudioFormat.FileSize + format.FileSize,
				Duration:     time.Duration(metadata.Duration) * time.Second,
			}
			optsByTab[YtDLPPickerTabAudioVideo] = append(optsByTab[YtDLPPickerTabAudioVideo], option)
		}
	}

	return optsByTab
}

func detectTabForFormat(format ytdlp.Format) YtDLPPickerTab {
	hasVideo := format.IsVideo()
	hasAudio := format.IsAudio()

	switch {
	case hasVideo && hasAudio:
		return YtDLPPickerTabAudioVideo
	case hasVideo && !hasAudio:
		return YtDLPPickerTabVideoOnly
	case !hasVideo && hasAudio:
		return YtDLPPickerTabAudioOnly
	default:
		return ""
	}
}

func orderedYtDLPTabs() []YtDLPPickerTab {
	return []YtDLPPickerTab{
		YtDLPPickerTabAudioOnly,
		YtDLPPickerTabVideoOnly,
		YtDLPPickerTabAudioVideo,
		YtDLPPickerTabSubtitles,
	}
}
