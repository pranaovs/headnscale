package tailscale

import "strings"

func HostNameFromDNSName(dnsName string) string {
	// Remove trailing dot, if present
	dnsName = strings.TrimSuffix(dnsName, ".")

	before, _, _ := strings.Cut(dnsName, ".")
	return before
}
