package pickersession

import (
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
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
		ytdlp: &picker.YtDLPState{
			ContentName:  metadata.Title,
			ActiveTab:    picker.YtDLPTabNone,
			OptionsByTab: optsByTab,
		},
		expiresAt: time.Now().Add(m.ttl),
	}

	return id, nil
}

func (m *PickerSessionManager) GetYtDLPPickerView(sessionID string, userID int64) (picker.YtDLPView, error) {
	return m.withYtDLPSessionView(sessionID, userID, func(s *pickerSession) error {
		return nil
	})
}

func (m *PickerSessionManager) SetYtDLPActiveTab(sessionID string, userID int64, tab picker.YtDLPTab) (picker.YtDLPView, error) {
	return m.withYtDLPSessionView(sessionID, userID, func(s *pickerSession) error {
		if tab == picker.YtDLPTabNone {
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
			s.ytdlp.ChosenTab = picker.YtDLPTabNone
			s.ytdlp.ChosenIndex = -1
		}
		return nil
	})
}

func (m *PickerSessionManager) ChooseYtDLPOption(sessionID string, userID int64, optionIdx int) (picker.YtDLPOption, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeYtDLP)
	if err != nil {
		return picker.YtDLPOption{}, err
	}

	options := s.ytdlp.OptionsByTab[s.ytdlp.ActiveTab]
	if optionIdx < 0 || optionIdx >= len(options) {
		return picker.YtDLPOption{}, ErrInvalidOptionIdx
	}

	s.ytdlp.ChosenTab = s.ytdlp.ActiveTab
	s.ytdlp.ChosenIndex = optionIdx
	s.ytdlp.HasChosen = true

	return options[optionIdx], nil
}

func (m *PickerSessionManager) ClearChosenYtDLPOption(sessionID string, userID int64) (picker.YtDLPView, error) {
	return m.withYtDLPSessionView(sessionID, userID, func(s *pickerSession) error {
		s.ytdlp.HasChosen = false
		s.ytdlp.ChosenTab = picker.YtDLPTabNone
		s.ytdlp.ChosenIndex = -1
		return nil
	})
}

func (m *PickerSessionManager) ConsumeChosenYtDLPOption(sessionID string, userID int64) (picker.YtDLPOption, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeYtDLP)
	if err != nil {
		return picker.YtDLPOption{}, err
	}

	if !s.ytdlp.HasChosen {
		return picker.YtDLPOption{}, ErrNoOptionsSelected
	}

	options := s.ytdlp.OptionsByTab[s.ytdlp.ChosenTab]
	if s.ytdlp.ChosenIndex < 0 || s.ytdlp.ChosenIndex >= len(options) {
		return picker.YtDLPOption{}, ErrInvalidOptionIdx
	}

	chosenOption := options[s.ytdlp.ChosenIndex]

	delete(m.sessions, sessionID)

	return chosenOption, nil
}

func (m *PickerSessionManager) withYtDLPSessionView(sessionID string, userID int64, fn func(*pickerSession) error) (picker.YtDLPView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeYtDLP)
	if err != nil {
		return picker.YtDLPView{}, err
	}

	if err := fn(s); err != nil {
		return picker.YtDLPView{}, err
	}
	return buildYtDLPPickerView(s), nil
}

func buildYtDLPPickerView(s *pickerSession) picker.YtDLPView {
	sourceOptions := s.ytdlp.OptionsByTab[s.ytdlp.ActiveTab]
	options := make([]picker.YtDLPOption, len(sourceOptions))
	copy(options, sourceOptions)

	tabs := make([]picker.YtDLPTab, 0, len(s.ytdlp.OptionsByTab))
	for _, tab := range orderedYtDLPTabs() {
		if len(s.ytdlp.OptionsByTab[tab]) > 0 {
			tabs = append(tabs, tab)
		}
	}

	return picker.YtDLPView{
		ContentName: s.ytdlp.ContentName,
		ActiveTab:   s.ytdlp.ActiveTab,
		Tabs:        tabs,
		Options:     options,
	}
}

func ParseYtDLPMetadata(metadata *ytdlp.Metadata) map[picker.YtDLPTab][]picker.YtDLPOption {
	optsByTab := make(map[picker.YtDLPTab][]picker.YtDLPOption)
	thumbnailURL := metadata.Thumbnail
	originalURL := metadata.OriginalURL
	duration := time.Duration(metadata.Duration) * time.Second

	for _, format := range metadata.Formats {
		tab := detectTabForFormat(format)
		if tab == "" {
			continue
		}

		option := picker.YtDLPOption{
			DisplayName:  format.GetDisplayName(nil, nil),
			ThumbnailURL: thumbnailURL,
			ContentURL:   originalURL,
			FormatID:     format.FormatID,
			FileSize:     format.FileSize,
			Duration:     duration,
			Format: media.DownloadFormat{
				HasAudio: format.IsAudio(),
				HasVideo: format.IsVideo(),
			},
		}

		optsByTab[tab] = append(optsByTab[tab], option)
	}

	if len(metadata.RequestedDownloads) == 0 {
		return optsByTab
	}

	bestAudioFormat := metadata.RequestedDownloads[0].GetBestAudioFormat()
	if bestAudioFormat != nil {
		for _, format := range metadata.Formats {
			if detectTabForFormat(format) != picker.YtDLPTabVideoOnly {
				continue
			}

			option := picker.YtDLPOption{
				DisplayName:  format.GetDisplayName(bestAudioFormat, &format),
				ContentURL:   originalURL,
				FormatID:     format.FormatID + "+" + bestAudioFormat.FormatID,
				ThumbnailURL: thumbnailURL,
				FileSize:     bestAudioFormat.FileSize + format.FileSize,
				Duration:     duration,
				Format: media.DownloadFormat{
					HasAudio: true,
					HasVideo: true,
				},
			}
			optsByTab[picker.YtDLPTabAudioVideo] = append(optsByTab[picker.YtDLPTabAudioVideo], option)
		}
	}

	return optsByTab
}

func detectTabForFormat(format ytdlp.Format) picker.YtDLPTab {
	hasVideo := format.IsVideo()
	hasAudio := format.IsAudio()

	switch {
	case hasVideo && hasAudio:
		return picker.YtDLPTabAudioVideo
	case hasVideo && !hasAudio:
		return picker.YtDLPTabVideoOnly
	case !hasVideo && hasAudio:
		return picker.YtDLPTabAudioOnly
	default:
		return ""
	}
}

func orderedYtDLPTabs() []picker.YtDLPTab {
	return []picker.YtDLPTab{
		picker.YtDLPTabAudioOnly,
		picker.YtDLPTabVideoOnly,
		picker.YtDLPTabAudioVideo,
		picker.YtDLPTabSubtitles,
	}
}
