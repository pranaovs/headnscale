package tailscale

import "strings"

func hostNameFromDNSName(dnsName string) string {
	// Remove trailing dot, if present
	dnsName = strings.TrimSuffix(dnsName, ".")

	// Split on the first dot and return the host
	if before, _, ok := strings.Cut(dnsName, "."); ok {
		return before
	}

	// Fallback: entire string is the host
	return dnsName
}
