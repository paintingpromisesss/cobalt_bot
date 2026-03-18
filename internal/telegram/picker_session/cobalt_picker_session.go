package pickersession

import (
	"fmt"
	"strings"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
)

func (m *PickerSessionManager) CreateCobaltSession(userID int64, cobaltResponse cobalt.MainResponse) (string, error) {
	opts := ParsePickerObjects(cobaltResponse)
	sel := make([]bool, len(opts))

	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := m.newUniqueSessionIDLocked()
	if err != nil {
		return "", err
	}

	m.sessions[id] = &pickerSession{
		sessionType: PickerSessionTypeCobalt,
		userID:      userID,
		cobalt: &picker.CobaltState{
			Selected: sel,
			Options:  opts,
		},
		expiresAt: time.Now().Add(m.ttl),
	}

	return id, nil
}
func (m *PickerSessionManager) GetCobaltPickerView(sessionID string, userID int64) (picker.CobaltView, error) {
	return m.withCobaltSessionView(sessionID, userID, func(s *pickerSession) error {
		return nil
	})
}

func (m *PickerSessionManager) ToggleCobaltPickerOption(sessionID string, userID int64, optionIdx int) (picker.CobaltView, error) {
	return m.withCobaltSessionView(sessionID, userID, func(s *pickerSession) error {
		if optionIdx < 0 || optionIdx >= len(s.cobalt.Options) {
			return ErrInvalidOptionIdx
		}
		s.cobalt.Selected[optionIdx] = !s.cobalt.Selected[optionIdx]
		return nil
	})
}

func (m *PickerSessionManager) MarkAllCobaltPickerOptions(sessionID string, userID int64, flag bool) (picker.CobaltView, error) {
	return m.withCobaltSessionView(sessionID, userID, func(s *pickerSession) error {
		for i := range s.cobalt.Selected {
			s.cobalt.Selected[i] = flag
		}
		return nil
	})
}

func (m *PickerSessionManager) ConsumeSelectedCobaltOptions(sessionID string, userID int64) ([]picker.CobaltOption, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeCobalt)
	if err != nil {
		return nil, err
	}

	out := make([]picker.CobaltOption, 0, len(s.cobalt.Options))
	for i, opt := range s.cobalt.Options {
		if s.cobalt.Selected[i] {
			out = append(out, opt)
		}
	}

	if len(out) == 0 {
		return nil, ErrNoOptionsSelected
	}

	delete(m.sessions, sessionID)

	return out, nil
}

func (m *PickerSessionManager) withCobaltSessionView(sessionID string, userID int64, fn func(*pickerSession) error) (picker.CobaltView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeCobalt)
	if err != nil {
		return picker.CobaltView{}, err
	}

	if err := fn(s); err != nil {
		return picker.CobaltView{}, err
	}

	return buildCobaltPickerView(s), nil

}

func buildCobaltPickerView(session *pickerSession) picker.CobaltView {
	v := picker.CobaltView{
		Options: make([]picker.CobaltOptionView, len(session.cobalt.Options)),
	}
	for i := range session.cobalt.Options {
		v.Options[i] = picker.CobaltOptionView{
			CobaltOption: session.cobalt.Options[i],
			Selected:     session.cobalt.Selected[i],
		}
	}

	return v
}

func ParsePickerObjects(cobaltResponse cobalt.MainResponse) []picker.CobaltOption {
	objects := cobaltResponse.Picker
	opts := make([]picker.CobaltOption, len(objects))
	for i, obj := range objects {
		opts[i] = picker.CobaltOption{
			Label:    fmt.Sprintf("%s #%d", strings.ToUpper(string(obj.Type)), i+1),
			URL:      obj.Url,
			Filename: cobalt.PickerFilenameByType(obj.Type, i+1),
		}
	}
	if cobaltResponse.PickerAudio != nil && cobaltResponse.AudioFilename != nil {
		opts = append(opts, picker.CobaltOption{
			Label:    "Аудио",
			URL:      *cobaltResponse.PickerAudio,
			Filename: *cobaltResponse.AudioFilename,
		})
	}
	return opts
}
