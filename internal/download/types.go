package download

import (
	"fmt"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
)

type Request struct {
	UserID      int64
	SourceURL   string
	PickerIndex *int
}

func (r Request) Validate() error {
	if r.UserID <= 0 {
		return fmt.Errorf("user id must be positive, got %d", r.UserID)
	}
	if strings.TrimSpace(r.SourceURL) == "" {
		return fmt.Errorf("source url is required")
	}
	return nil
}

type PickerOption struct {
	Index    int
	URL      string
	Label    string
	Filename string
}

type Result struct {
	Status        string
	File          *cobalt.CobaltDownloadedFile
	PickerOptions []PickerOption
}
