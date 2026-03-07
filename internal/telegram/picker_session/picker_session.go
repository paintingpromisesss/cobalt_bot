package pickersession

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
)

var (
	ErrSessionNotFound   = errors.New("picker session not found")
	ErrSessionForbidden  = errors.New("picker session access forbidden")
	ErrSessionExpired    = errors.New("picker session expired")
	ErrInvalidOptionIdx  = errors.New("invalid option index")
	ErrNoOptionsSelected = errors.New("no options selected")
)

type PickerOption struct {
	Label    string
	URL      string
	Filename string
}

type pickerSession struct {
	userID    int64
	selected  []bool
	options   []PickerOption
	expiresAt time.Time
}

type PickerView struct {
	Options []PickerOptionView
}

type PickerOptionView struct {
	PickerOption
	Selected bool
}

type PickerSessionManager struct {
	sessions map[string]*pickerSession
	mu       sync.Mutex
	seq      uint64
	ttl      time.Duration
}

func NewPickerSessionManager(ttl time.Duration) *PickerSessionManager {
	return &PickerSessionManager{
		sessions: make(map[string]*pickerSession),
		ttl:      ttl,
	}
}

func (m *PickerSessionManager) CreateSession(userID int64, options []cobalt.PickerObject) string {
	id := fmt.Sprintf("%d", atomic.AddUint64(&m.seq, 1))

	opts := ParsePickerObjects(options)
	sel := make([]bool, len(opts))

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[id] = &pickerSession{
		userID:    userID,
		selected:  sel,
		options:   opts,
		expiresAt: time.Now().Add(m.ttl),
	}

	return id
}

func (m *PickerSessionManager) GetPickerView(sessionID string, userID int64) (PickerView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSession(sessionID, userID)
	if err != nil {
		return PickerView{}, err
	}

	return buildPickerView(s), nil

}

func (m *PickerSessionManager) TogglePickerOption(sessionID string, userID int64, optionIdx int) (PickerView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSession(sessionID, userID)
	if err != nil {
		return PickerView{}, err
	}
	if optionIdx < 0 || optionIdx >= len(s.options) {
		return PickerView{}, ErrInvalidOptionIdx
	}

	s.selected[optionIdx] = !s.selected[optionIdx]

	return buildPickerView(s), nil
}

func (m *PickerSessionManager) MarkAllPickerOptions(sessionID string, userID int64, flag bool) (PickerView, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSession(sessionID, userID)
	if err != nil {
		return PickerView{}, err
	}
	for i := range s.selected {
		s.selected[i] = flag
	}
	return buildPickerView(s), nil
}

func (m *PickerSessionManager) ConsumeSelectedOptions(sessionID string, userID int64) ([]PickerOption, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, err := m.validateSession(sessionID, userID)
	if err != nil {
		return nil, err
	}

	out := make([]PickerOption, 0, len(s.options))
	for i, opt := range s.options {
		if s.selected[i] {
			out = append(out, opt)
		}
	}

	if len(out) == 0 {
		return nil, ErrNoOptionsSelected
	}

	delete(m.sessions, sessionID)

	return out, nil
}

func (m *PickerSessionManager) validateSession(sessionID string, userID int64) (*pickerSession, error) {
	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if session.userID != userID {
		return nil, ErrSessionForbidden
	}

	if time.Now().After(session.expiresAt) {
		delete(m.sessions, sessionID)
		return nil, ErrSessionExpired
	}

	return session, nil
}

func buildPickerView(session *pickerSession) PickerView {
	v := PickerView{
		Options: make([]PickerOptionView, len(session.options)),
	}
	for i := range session.options {
		v.Options[i] = PickerOptionView{
			PickerOption: session.options[i],
			Selected:     session.selected[i],
		}
	}

	return v
}

func ParsePickerObjects(objects []cobalt.PickerObject) []PickerOption {
	opts := make([]PickerOption, len(objects))
	for i, obj := range objects {
		opts[i] = PickerOption{
			Label:    fmt.Sprintf("%s #%d", strings.ToUpper(string(obj.Type)), i+1),
			URL:      obj.Url,
			Filename: cobalt.PickerFilenameByType(obj.Type, i+1),
		}
	}
	return opts
}
