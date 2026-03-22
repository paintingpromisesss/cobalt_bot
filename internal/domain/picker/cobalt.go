package picker

type CobaltOption struct {
	Label    string
	URL      string
	Filename string
}

type CobaltState struct {
	Selected []bool
	Options  []CobaltOption
}

type CobaltView struct {
	Options []CobaltOptionView
}

type CobaltOptionView struct {
	CobaltOption
	Selected bool
}

func (s *CobaltState) ToggleOption(idx int) error {
	if idx < 0 || idx >= len(s.Options) {
		return ErrInvalidOptionIdx
	}

	s.Selected[idx] = !s.Selected[idx]
	return nil
}

func (s *CobaltState) SelectAll() {
	for i := range s.Selected {
		s.Selected[i] = true
	}
}

func (s *CobaltState) ClearAll() {
	for i := range s.Selected {
		s.Selected[i] = false
	}
}

func (s CobaltState) SelectedOptions() ([]CobaltOption, error) {
	options := make([]CobaltOption, 0, len(s.Options))
	for i, opt := range s.Options {
		if s.Selected[i] {
			options = append(options, opt)
		}
	}

	if len(options) == 0 {
		return nil, ErrNoOptionsSelected
	}

	return options, nil
}
