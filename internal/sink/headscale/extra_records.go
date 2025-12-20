package headscale

import (
	"encoding/json"
	"os"
	libsort "sort"

	"github.com/pranaovs/headnscale/internal/types"
)

// Ref: https://github.com/juanfont/headscale/blob/main/docs/ref/dns.md
func create(nodes []types.Node, baseDomain string) []map[string]any {
	if baseDomain == "" {
		baseDomain = ""
	} else {
		baseDomain = "." + baseDomain
	}

	records := make([]map[string]any, 0)

	for _, node := range nodes {
		subdomain := node.Hostname

		// Create A record for IPv4
		if node.IP.IPv4 != nil {
			record := map[string]any{
				"name":  subdomain + baseDomain,
				"type":  "A",
				"value": node.IP.IPv4.String(),
			}
			records = append(records, record)
		}

		// Create AAAA record for IPv6 if available
		if node.IP.IPv6 != nil {
			record := map[string]any{
				"name":  subdomain + baseDomain,
				"type":  "AAAA",
				"value": node.IP.IPv6.String(),
			}
			records = append(records, record)
		}
	}

	return records
}

func sort(records []map[string]any) []map[string]any {
	// Sort the keys
	// "Be sure to "sort keys" and produce a stable output in case you generate the JSON file with a script.
	// Headscale uses a checksum to detect changes to the file and a stable output avoids unnecessary processing."
	libsort.Slice(records, func(i, j int) bool {
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

func write(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
