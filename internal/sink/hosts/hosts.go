package hosts

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/pranaovs/headnscale/internal/types"
)

func create(nodes []types.Node, baseDomain string) []string {
	hosts := make([]string, 0)

	if baseDomain == "" {
		baseDomain = ""
	} else {
		baseDomain = "." + baseDomain
	}

	for _, node := range nodes {
		subdomain := node.Hostname

		if node.IP.IPv4 != nil {
			hosts = append(hosts, fmt.Sprintf("%s %s%s\n", node.IP.IPv4.String(), subdomain, baseDomain))
		}
		if node.IP.IPv6 != nil {
			hosts = append(hosts, fmt.Sprintf("%s %s%s\n", node.IP.IPv6.String(), subdomain, baseDomain))
		}
	}

	return hosts
}

func sort(hosts []string) []string {
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

func write(path string, hosts []string) error {
	data := strings.Join(hosts, "")
	return os.WriteFile(path, []byte(data), 0o644)
}
