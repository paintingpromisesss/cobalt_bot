package urlvalidator

import (
	"strings"
)

type URLValidator struct {
	allowlist []string
}

func NewURLValidator(availableServices []string) *URLValidator {
	return &URLValidator{
		allowlist: buildAllowlist(availableServices),
	}
}

func (v *URLValidator) Validate(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || v == nil || len(v.allowlist) == 0 {
		return "", false
	}
	if strings.ContainsAny(value, " \t\r\n") {
		return "", false
	}

	if v.isAllowed(value) {
		return value, true
	}

	return "", false
}
