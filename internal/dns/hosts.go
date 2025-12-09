package dns

import (
	"fmt"
	"slices"

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
	slices.Sort(hosts)
	return hosts
}
