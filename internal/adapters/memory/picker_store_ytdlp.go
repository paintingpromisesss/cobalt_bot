package memory

import (
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
)

func (m *PickerStore) CreateYtDLPSession(userID int64, state picker.YtDLPState) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := m.newUniqueSessionIDLocked()
	if err != nil {
		return "", err
	}

	m.sessions[id] = &pickerSession{
		sessionType: PickerSessionTypeYtDLP,
		userID:      userID,
		ytdlp:       func() *picker.YtDLPState { cloned := cloneYtDLPState(state); return &cloned }(),
		expiresAt:   time.Now().Add(m.ttl),
	}

	return id, nil
}

func (m *PickerStore) GetYtDLPState(sessionID string, userID int64) (picker.YtDLPState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeYtDLP)
	if err != nil {
		return picker.YtDLPState{}, err
	}

	return cloneYtDLPState(*s.ytdlp), nil
}

func (m *PickerStore) SaveYtDLPState(sessionID string, userID int64, state picker.YtDLPState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSessionLocked(sessionID, userID, PickerSessionTypeYtDLP)
	if err != nil {
		return err
	}

	cloned := cloneYtDLPState(state)
	s.ytdlp = &cloned
	return nil
}

func (m *PickerStore) DeleteYtDLPSession(sessionID string, userID int64) error {
	return m.deleteSession(sessionID, userID, PickerSessionTypeYtDLP)
}

func cloneYtDLPState(state picker.YtDLPState) picker.YtDLPState {
	optionsByTab := make(map[picker.YtDLPTab][]picker.YtDLPOption, len(state.OptionsByTab))
	for tab, options := range state.OptionsByTab {
		clonedOptions := make([]picker.YtDLPOption, len(options))
		copy(clonedOptions, options)
		optionsByTab[tab] = clonedOptions
	}

	return picker.YtDLPState{
		ContentName:  state.ContentName,
		ActiveTab:    state.ActiveTab,
		OptionsByTab: optionsByTab,
		ChosenTab:    state.ChosenTab,
		ChosenIndex:  state.ChosenIndex,
		HasChosen:    state.HasChosen,
	}
}
