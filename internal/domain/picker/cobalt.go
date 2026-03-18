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
