package proxies

import (
	"fmt"
)

func dedupKey(proxy map[string]any) string {
	server, _ := proxy["server"].(string)
	if server == "" {
		return ""
	}
	servername, _ := proxy["servername"].(string)

	password, _ := proxy["password"].(string)
	if password == "" {
		password, _ = proxy["uuid"].(string)
	}

	sni, _ := proxy["sni"].(string)
	network, _ := proxy["network"].(string)

	return fmt.Sprintf("%s:%v:%s:%s:%s:%s", server, proxy["port"], servername, password, sni, network)
}

func DeduplicateProxies(proxies []map[string]any) []map[string]any {
	seenKeys := make(map[string]bool)
	result := make([]map[string]any, 0, len(proxies))

	for _, proxy := range proxies {
		key := dedupKey(proxy)
		if key == "" {
			continue
		}
		if !seenKeys[key] {
			seenKeys[key] = true
			result = append(result, proxy)
		}
	}

	return result
}
