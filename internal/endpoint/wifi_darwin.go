//go:build darwin

package endpoint

import "context"

// airportBin is Apple's private Wi-Fi diagnostic tool; `-I` prints the current
// association report (RSSI/noise/SSID/channel/tx-rate).
const airportBin = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"

func newPlatformWiFiCollector() WiFiCollector {
	return cmdWiFiCollector{
		run:   func(ctx context.Context) (string, error) { return execText(ctx, airportBin, "-I") },
		parse: parseAirportI,
	}
}
