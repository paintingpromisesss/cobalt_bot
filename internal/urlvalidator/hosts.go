package urlvalidator

import (
	"net"
	"net/url"
	"strings"
)

func (v *URLValidator) isAllowed(rawURL string) bool {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Host == "" {
		return false
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}

	if parsed.User != nil {
		return false
	}

	host := normalizeHost(parsed.Hostname())
	if host == "" {
		return false
	}
	if net.ParseIP(host) != nil {
		return false
	}
	if !strings.Contains(host, ".") {
		return false
	}

	for _, allowed := range v.allowlist {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}

	return false
}

func normalizeHost(value string) string {
	host := strings.TrimSpace(strings.ToLower(value))
	host = strings.TrimSuffix(host, ".")

	if strings.Contains(host, "://") {
		if parsed, err := url.Parse(host); err == nil {
			host = parsed.Hostname()
		}
	}

	if strings.Contains(host, ":") {
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			host = parsedHost
		}
	}

	return strings.TrimSpace(host)
}
