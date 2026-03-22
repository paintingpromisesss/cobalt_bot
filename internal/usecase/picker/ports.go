package picker

import domainpicker "github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"

type Store interface {
	CreateCobaltSession(userID int64, state domainpicker.CobaltState) (string, error)
	GetCobaltState(sessionID string, userID int64) (domainpicker.CobaltState, error)
	SaveCobaltState(sessionID string, userID int64, state domainpicker.CobaltState) error
	DeleteCobaltSession(sessionID string, userID int64) error

	CreateYtDLPSession(userID int64, state domainpicker.YtDLPState) (string, error)
	GetYtDLPState(sessionID string, userID int64) (domainpicker.YtDLPState, error)
	SaveYtDLPState(sessionID string, userID int64, state domainpicker.YtDLPState) error
	DeleteYtDLPSession(sessionID string, userID int64) error
}
