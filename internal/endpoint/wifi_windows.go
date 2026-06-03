//go:build windows

package endpoint

import "context"

func newPlatformWiFiCollector() WiFiCollector {
	return cmdWiFiCollector{
		run:   func(ctx context.Context) (string, error) { return execText(ctx, "netsh", "wlan", "show", "interfaces") },
		parse: parseNetshWlan,
	}
}
