package picker

type CobaltInitData struct {
	Options []CobaltOption
}

type YtDLPInitData struct {
	ContentName  string
	OptionsByTab map[YtDLPTab][]YtDLPOption
}
