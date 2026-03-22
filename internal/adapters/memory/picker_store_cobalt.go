package memory

import (
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
)

func (m *PickerStore) CreateCobaltSession(userID int64, state picker.CobaltState) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := m.newUniqueSessionIDLocked()
	if err != nil {
		return "", err
	}

	m.sessions[id] = &pickerSession{
		sessionType: PickerSessionTypeCobalt,
		userID:      userID,
		cobalt:      cloneCobaltState(state),
		expiresAt:   time.Now().Add(m.ttl),
	}

	return id, nil
}

func (m *PickerStore) GetCobaltState(sessionID string, userID int64) (picker.CobaltState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeCobalt)
	if err != nil {
		return picker.CobaltState{}, err
	}

	return *cloneCobaltState(*s.cobalt), nil
}

func (m *PickerStore) SaveCobaltState(sessionID string, userID int64, state picker.CobaltState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeCobalt)
	if err != nil {
		return err
	}

	s.cobalt = cloneCobaltState(state)

	return nil
}

func (m *PickerStore) DeleteCobaltSession(sessionID string, userID int64) error {
	return m.deleteSession(sessionID, userID, PickerSessionTypeCobalt)
}

func cloneCobaltState(state picker.CobaltState) *picker.CobaltState {
	selected := make([]bool, len(state.Selected))
	copy(selected, state.Selected)
	options := make([]picker.CobaltOption, len(state.Options))
	copy(options, state.Options)

	return &picker.CobaltState{
		Selected: selected,
		Options:  options,
	}
}
