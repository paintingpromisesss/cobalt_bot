package handlers

import (
	"errors"
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
)

func cobaltErrorToErr(errorObj *cobalt.ErrorObject) error {
	if errorObj == nil {
		return errors.New("cobalt returned an error response without error details")
	}

	details := errorObj.Code
	if errorObj.Context != nil {
		if errorObj.Context.Service != "" {
			details += ": " + errorObj.Context.Service
		}
		if errorObj.Context.Limit > 0 {
			details += fmt.Sprintf(", limit=%.2f", errorObj.Context.Limit)
		}
	}

	return fmt.Errorf("cobalt error: %s", details)
}
