package picker

type CobaltAction string

const (
	CobaltActionToggle    CobaltAction = "toggle"
	CobaltActionSelectAll CobaltAction = "select_all"
	CobaltActionClearAll  CobaltAction = "clear_all"
	CobaltActionDownload  CobaltAction = "download"
	CobaltActionCancel    CobaltAction = "cancel"
)

type YtDLPAction string

const (
	YtDLPActionTab         YtDLPAction = "select_tab"
	YtDLPActionChoose      YtDLPAction = "choose"
	YtDLPActionDownload    YtDLPAction = "download"
	YtDLPActionCancel      YtDLPAction = "cancel"
	YtDLPActionConfirmBack YtDLPAction = "confirm_back"
	YtDLPActionBack        YtDLPAction = "back"
)
