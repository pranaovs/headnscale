package tailscale

import "strings"

func hostNameFromDNSName(dnsName string) string {
	// Remove trailing dot, if present
	dnsName = strings.TrimSuffix(dnsName, ".")

	before, _, _ := strings.Cut(dnsName, ".")
	return before
}
