package picker

import (
	"fmt"
	domainpicker "github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) InitCobalt(input InitCobaltInput) (CobaltResult, error) {
	sessionID, err := s.store.CreateCobaltSession(input.UserID, buildCobaltState(input.Data))
	if err != nil {
		return CobaltResult{}, err
	}

	state, err := s.store.GetCobaltState(sessionID, input.UserID)
	if err != nil {
		return CobaltResult{}, err
	}

	return CobaltResult{
		Kind:      ResultKindView,
		SessionID: sessionID,
		View:      buildCobaltView(state),
	}, nil
}

func (s *Service) HandleCobalt(input CobaltInput) (CobaltResult, error) {
	state, err := s.store.GetCobaltState(input.SessionID, input.UserID)
	if err != nil {
		return CobaltResult{}, err
	}

	switch input.Action {
	case domainpicker.CobaltActionToggle:
		if err := state.ToggleOption(input.OptionIdx); err != nil {
			return CobaltResult{}, err
		}
		if err := s.store.SaveCobaltState(input.SessionID, input.UserID, state); err != nil {
			return CobaltResult{}, err
		}
		return CobaltResult{Kind: ResultKindView, SessionID: input.SessionID, View: buildCobaltView(state)}, nil
	case domainpicker.CobaltActionSelectAll:
		state.SelectAll()
		if err := s.store.SaveCobaltState(input.SessionID, input.UserID, state); err != nil {
			return CobaltResult{}, err
		}
		return CobaltResult{Kind: ResultKindView, SessionID: input.SessionID, View: buildCobaltView(state)}, nil
	case domainpicker.CobaltActionClearAll:
		state.ClearAll()
		if err := s.store.SaveCobaltState(input.SessionID, input.UserID, state); err != nil {
			return CobaltResult{}, err
		}
		return CobaltResult{Kind: ResultKindView, SessionID: input.SessionID, View: buildCobaltView(state)}, nil
	case domainpicker.CobaltActionDownload:
		options, err := state.SelectedOptions()
		if err != nil {
			return CobaltResult{}, err
		}
		if err := s.store.DeleteCobaltSession(input.SessionID, input.UserID); err != nil {
			return CobaltResult{}, err
		}
		return CobaltResult{Kind: ResultKindDownload, Options: options}, nil
	case domainpicker.CobaltActionCancel:
		if err := s.store.DeleteCobaltSession(input.SessionID, input.UserID); err != nil {
			return CobaltResult{}, err
		}
		return CobaltResult{Kind: ResultKindCanceled}, nil
	default:
		return CobaltResult{}, fmt.Errorf("unknown cobalt picker action: %q", input.Action)
	}
}

func (s *Service) InitYtDLP(input InitYtDLPInput) (YtDLPResult, error) {
	sessionID, err := s.store.CreateYtDLPSession(input.UserID, buildYtDLPState(input.Data))
	if err != nil {
		return YtDLPResult{}, err
	}

	state, err := s.store.GetYtDLPState(sessionID, input.UserID)
	if err != nil {
		return YtDLPResult{}, err
	}

	return YtDLPResult{
		Kind:      ResultKindView,
		SessionID: sessionID,
		View:      buildYtDLPView(state),
	}, nil
}

func (s *Service) HandleYtDLP(input YtDLPInput) (YtDLPResult, error) {
	state, err := s.store.GetYtDLPState(input.SessionID, input.UserID)
	if err != nil {
		return YtDLPResult{}, err
	}

	switch input.Action {
	case domainpicker.YtDLPActionTab:
		if err := state.SelectTab(input.Tab); err != nil {
			return YtDLPResult{}, err
		}
		if err := s.store.SaveYtDLPState(input.SessionID, input.UserID, state); err != nil {
			return YtDLPResult{}, err
		}
		return YtDLPResult{Kind: ResultKindView, SessionID: input.SessionID, View: buildYtDLPView(state)}, nil
	case domainpicker.YtDLPActionChoose:
		option, err := state.ChooseActiveOption(input.OptionIdx)
		if err != nil {
			return YtDLPResult{}, err
		}
		if err := s.store.SaveYtDLPState(input.SessionID, input.UserID, state); err != nil {
			return YtDLPResult{}, err
		}
		return YtDLPResult{Kind: ResultKindConfirmation, SessionID: input.SessionID, Option: &option}, nil
	case domainpicker.YtDLPActionDownload:
		option, err := state.SelectedOption()
		if err != nil {
			return YtDLPResult{}, err
		}
		if err := s.store.DeleteYtDLPSession(input.SessionID, input.UserID); err != nil {
			return YtDLPResult{}, err
		}
		return YtDLPResult{Kind: ResultKindDownload, Option: &option}, nil
	case domainpicker.YtDLPActionConfirmBack:
		state.ResetChoice()
		if err := s.store.SaveYtDLPState(input.SessionID, input.UserID, state); err != nil {
			return YtDLPResult{}, err
		}
		return YtDLPResult{Kind: ResultKindView, SessionID: input.SessionID, View: buildYtDLPView(state)}, nil
	case domainpicker.YtDLPActionBack:
		if err := state.SelectTab(input.Tab); err != nil {
			return YtDLPResult{}, err
		}
		if err := s.store.SaveYtDLPState(input.SessionID, input.UserID, state); err != nil {
			return YtDLPResult{}, err
		}
		return YtDLPResult{Kind: ResultKindView, SessionID: input.SessionID, View: buildYtDLPView(state)}, nil
	case domainpicker.YtDLPActionCancel:
		if err := s.store.DeleteYtDLPSession(input.SessionID, input.UserID); err != nil {
			return YtDLPResult{}, err
		}
		return YtDLPResult{Kind: ResultKindCanceled}, nil
	default:
		return YtDLPResult{}, fmt.Errorf("unknown yt-dlp picker action: %q", input.Action)
	}
}

func buildCobaltView(state domainpicker.CobaltState) *domainpicker.CobaltView {
	view := &domainpicker.CobaltView{
		Options: make([]domainpicker.CobaltOptionView, len(state.Options)),
	}
	for i := range state.Options {
		view.Options[i] = domainpicker.CobaltOptionView{
			CobaltOption: state.Options[i],
			Selected:     state.Selected[i],
		}
	}
	return view
}

func buildYtDLPView(state domainpicker.YtDLPState) *domainpicker.YtDLPView {
	sourceOptions := state.OptionsByTab[state.ActiveTab]
	options := make([]domainpicker.YtDLPOption, len(sourceOptions))
	copy(options, sourceOptions)

	tabs := make([]domainpicker.YtDLPTab, 0, len(state.OptionsByTab))
	for _, tab := range orderedYtDLPTabs() {
		if len(state.OptionsByTab[tab]) > 0 {
			tabs = append(tabs, tab)
		}
	}

	return &domainpicker.YtDLPView{
		ContentName: state.ContentName,
		ActiveTab:   state.ActiveTab,
		Tabs:        tabs,
		Options:     options,
	}
}

func orderedYtDLPTabs() []domainpicker.YtDLPTab {
	return []domainpicker.YtDLPTab{
		domainpicker.YtDLPTabAudioOnly,
		domainpicker.YtDLPTabVideoOnly,
		domainpicker.YtDLPTabAudioVideo,
		domainpicker.YtDLPTabSubtitles,
	}
}

func buildCobaltState(data domainpicker.CobaltInitData) domainpicker.CobaltState {
	options := make([]domainpicker.CobaltOption, len(data.Options))
	copy(options, data.Options)
	return domainpicker.CobaltState{
		Selected: make([]bool, len(options)),
		Options:  options,
	}
}

func buildYtDLPState(data domainpicker.YtDLPInitData) domainpicker.YtDLPState {
	optionsByTab := make(map[domainpicker.YtDLPTab][]domainpicker.YtDLPOption, len(data.OptionsByTab))
	for tab, options := range data.OptionsByTab {
		cloned := make([]domainpicker.YtDLPOption, len(options))
		copy(cloned, options)
		optionsByTab[tab] = cloned
	}

	return domainpicker.YtDLPState{
		ContentName:  data.ContentName,
		ActiveTab:    domainpicker.YtDLPTabNone,
		OptionsByTab: optionsByTab,
		ChosenTab:    domainpicker.YtDLPTabNone,
		ChosenIndex:  -1,
	}
}
