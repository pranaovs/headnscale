package dns

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pranaovs/headnscale/internal/types"
)

func CreateHosts(subdomains []string, baseDomain string, node types.Node) []string {
	hosts := make([]string, 0)

	for _, subdomain := range subdomains {
		// Create A record for IPv4
		if node.IP.IPv4 != nil {
			hosts = append(hosts, fmt.Sprintf("%s %s.%s\n", node.IP.IPv4.String(), subdomain, baseDomain))
		}

		// Create AAAA record for IPv6 if available
		if node.IP.IPv6 != nil {
			hosts = append(hosts, fmt.Sprintf("%s %s.%s\n", node.IP.IPv6.String(), subdomain, baseDomain))
		}
	}
	return hosts
}

func SortHosts(hosts []string) []string {
	slices.SortFunc(hosts, func(a, b string) int {
		// Split into fields: ["IP", "hostname"]
		partsA := strings.Fields(a)
		partsB := strings.Fields(b)

		// Safety check: fallback to normal string compare
		if len(partsA) < 2 || len(partsB) < 2 {
			return strings.Compare(a, b)
		}

		// Compare the SECOND field
		return strings.Compare(partsA[1], partsB[1])
	})

	return hosts
}
