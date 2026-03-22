package picker

import "errors"

var (
	ErrSessionNotFound   = errors.New("picker session not found")
	ErrSessionForbidden  = errors.New("picker session access forbidden")
	ErrSessionExpired    = errors.New("picker session expired")
	ErrInvalidOptionIdx  = errors.New("invalid option index")
	ErrNoOptionsSelected = errors.New("no options selected")
	ErrWrongSessionType  = errors.New("wrong session type")
	ErrInvalidYtDLPTab   = errors.New("invalid yt-dlp picker tab")
)
