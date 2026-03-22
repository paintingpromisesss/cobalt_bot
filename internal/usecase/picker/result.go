package picker

import domainpicker "github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"

type CobaltInput struct {
	Action    domainpicker.CobaltAction
	SessionID string
	UserID    int64
	OptionIdx int
}

type YtDLPInput struct {
	Action    domainpicker.YtDLPAction
	SessionID string
	UserID    int64
	Tab       domainpicker.YtDLPTab
	OptionIdx int
}

type InitCobaltInput struct {
	UserID int64
	Data   domainpicker.CobaltInitData
}

type InitYtDLPInput struct {
	UserID int64
	Data   domainpicker.YtDLPInitData
}

type ResultKind string

const (
	ResultKindView         ResultKind = "view"
	ResultKindDownload     ResultKind = "download"
	ResultKindConfirmation ResultKind = "confirmation"
	ResultKindCanceled     ResultKind = "canceled"
)

type CobaltResult struct {
	Kind      ResultKind
	SessionID string
	View      *domainpicker.CobaltView
	Options   []domainpicker.CobaltOption
}

type YtDLPResult struct {
	Kind      ResultKind
	SessionID string
	View      *domainpicker.YtDLPView
	Option    *domainpicker.YtDLPOption
}
