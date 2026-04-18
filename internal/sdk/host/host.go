package host

import "strings"

// ParseHosts splits a comma-separated hosts string into a slice of URLs,
// prefixing with http:// if no scheme is present.
func ParseHosts(hosts string) []string {
	parts := strings.Split(hosts, ",")
	addresses := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if !strings.HasPrefix(p, "http") {
			p = "http://" + p
		}

		addresses = append(addresses, p)
	}

	return addresses
}
