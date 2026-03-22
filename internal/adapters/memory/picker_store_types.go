package memory

import (
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
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
