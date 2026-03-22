package download

import (
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
)

type Input struct {
	UserID int64
	URL    string
}

type Result interface {
	NormalizedURL() string
	isDownloadResult()
}

type InvalidURLResult struct{}

func (InvalidURLResult) NormalizedURL() string { return "" }
func (InvalidURLResult) isDownloadResult()     {}

type CobaltDirectResult struct {
	URL  string
	File media.RemoteFile
}

func (r CobaltDirectResult) NormalizedURL() string { return r.URL }
func (CobaltDirectResult) isDownloadResult()       {}

type CobaltPickerResult struct {
	URL  string
	Data picker.CobaltInitData
}

func (r CobaltPickerResult) NormalizedURL() string { return r.URL }
func (CobaltPickerResult) isDownloadResult()       {}

type YtDLPPickerResult struct {
	URL  string
	Data picker.YtDLPInitData
}

func (r YtDLPPickerResult) NormalizedURL() string { return r.URL }
func (YtDLPPickerResult) isDownloadResult()       {}

type YtDLPDirectResult struct {
	URL    string
	Option picker.YtDLPOption
}

func (r YtDLPDirectResult) NormalizedURL() string { return r.URL }
func (YtDLPDirectResult) isDownloadResult()       {}
