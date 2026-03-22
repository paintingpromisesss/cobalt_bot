package urlpolicy

import (
	"slices"
	"strings"
)

func buildAllowlist(availableServices []string) []string {
	seen := make(map[string]struct{})
	allowlist := make([]string, 0)

	for _, service := range availableServices {
		serviceKey := strings.TrimSpace(strings.ToLower(service))
		domains, ok := serviceDomainMap[serviceKey]
		if !ok {
			continue
		}

		for _, domain := range domains {
			host := normalizeHost(domain)
			if host == "" {
				continue
			}
			if strings.Contains(host, ":") || !strings.Contains(host, ".") {
				continue
			}
			if _, ok := seen[host]; ok {
				continue
			}
			seen[host] = struct{}{}
			allowlist = append(allowlist, host)
		}
	}

	slices.Sort(allowlist)
	return allowlist
}
