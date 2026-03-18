package pickersession

import (
	"errors"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
)

var (
	ErrSessionNotFound   = errors.New("picker session not found")
	ErrSessionForbidden  = errors.New("picker session access forbidden")
	ErrSessionExpired    = errors.New("picker session expired")
	ErrInvalidOptionIdx  = errors.New("invalid option index")
	ErrNoOptionsSelected = errors.New("no options selected")
	ErrWrongSessionType  = errors.New("wrong session type")

	ErrInvalidYtDLPTab = errors.New("invalid yt-dlp picker tab")
)

type PickerSessionType string

const (
	PickerSessionTypeCobalt PickerSessionType = "cobalt"
	PickerSessionTypeYtDLP  PickerSessionType = "yt-dlp"
)

type pickerSession struct {
	sessionType PickerSessionType
	userID      int64
	cobalt      *picker.CobaltState
	ytdlp       *picker.YtDLPState
	expiresAt   time.Time
}
