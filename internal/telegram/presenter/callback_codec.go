package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
)

func EncodeCobaltPickerCallbackData(action picker.CobaltAction, sessionID string, optionIdx int) string {
	if optionIdx >= 0 {
		return string(action) + ":" + sessionID + ":" + strconv.Itoa(optionIdx)
	}
	return string(action) + ":" + sessionID
}

func ParseCobaltPickerCallbackData(data string) (action picker.CobaltAction, sessionID string, optionIdx int, err error) {
	parts := strings.Split(strings.TrimSpace(data), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return "", "", 0, fmt.Errorf("invalid callback data format")
	}

	action, sessionID, optionIdx = picker.CobaltAction(parts[0]), parts[1], -1
	if len(parts) == 3 {
		idx, convErr := strconv.Atoi(parts[2])
		if convErr != nil {
			return "", "", 0, fmt.Errorf("invalid option index: %v", convErr)
		}
		optionIdx = idx
	}
	return action, sessionID, optionIdx, nil
}

func EncodeYtDLPPickerCallbackData(action picker.YtDLPAction, sessionID string, tab picker.YtDLPTab, optionIdx int) string {
	if optionIdx >= 0 {
		return string(action) + ":" + sessionID + ":" + string(tab) + ":" + strconv.Itoa(optionIdx)
	}
	if tab != picker.YtDLPTabNone && tab != "" {
		return string(action) + ":" + sessionID + ":" + string(tab)
	}
	return string(action) + ":" + sessionID
}

func ParseYtDLPPickerCallbackData(data string) (action picker.YtDLPAction, sessionID string, tab picker.YtDLPTab, optionIdx int, err error) {
	parts := strings.Split(strings.TrimSpace(data), ":")
	if len(parts) < 2 || len(parts) > 4 {
		return "", "", picker.YtDLPTabNone, -1, fmt.Errorf("invalid callback data format")
	}

	action, sessionID, tab, optionIdx = picker.YtDLPAction(parts[0]), parts[1], picker.YtDLPTabNone, -1
	if len(parts) >= 3 {
		tab = picker.YtDLPTab(parts[2])
	}
	if len(parts) == 4 {
		idx, convErr := strconv.Atoi(parts[3])
		if convErr != nil {
			return "", "", picker.YtDLPTabNone, -1, fmt.Errorf("invalid option index: %v", convErr)
		}
		optionIdx = idx
	}
	return action, sessionID, tab, optionIdx, nil
}
