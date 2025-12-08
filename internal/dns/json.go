package dns

import (
	"sort"

	"github.com/pranaovs/headnscale/internal/types"
)

// Ref: https://github.com/juanfont/headscale/blob/main/docs/ref/dns.md
func CreateJSON(subdomains []string, baseDomain string, node types.Node) []map[string]any {
	records := make([]map[string]any, 0)

	for _, subdomain := range subdomains {
		// Create A record for IPv4
		if node.IP.IPv4 != nil {
			record := map[string]any{
				"name":  subdomain + "." + baseDomain,
				"type":  "A",
				"value": node.IP.IPv4.String(),
			}
			records = append(records, record)
		}

		// Create AAAA record for IPv6 if available
		if node.IP.IPv6 != nil {
			record := map[string]any{
				"name":  subdomain + "." + baseDomain,
				"type":  "AAAA",
				"value": node.IP.IPv6.String(),
			}
			records = append(records, record)
		}
	}

	return records
}

func SortJSON(records []map[string]any) []map[string]any {
	// Sort the keys
	// "Be sure to "sort keys" and produce a stable output in case you generate the JSON file with a script.
	// Headscale uses a checksum to detect changes to the file and a stable output avoids unnecessary processing."
	sort.Slice(records, func(i, j int) bool {
		nameI := records[i]["name"].(string)
		nameJ := records[j]["name"].(string)

		if nameI != nameJ {
			return nameI < nameJ
		}

		typeI := records[i]["type"].(string)
		typeJ := records[j]["type"].(string)
		return typeI < typeJ
	})

	return records
}
