//go:build !linux && !darwin && !windows

package endpoint

import "context"

// unsupportedWiFi is the fallback on platforms with no Wi-Fi reader. It reports
// "no Wi-Fi present" so the rest of the sample (gateway/last-mile/sessions) still
// forms and attribution simply skips the Wi-Fi layer (graceful degradation).
type unsupportedWiFi struct{}

func (unsupportedWiFi) Collect(context.Context) (WiFi, error) { return WiFi{}, nil }

func newPlatformWiFiCollector() WiFiCollector { return unsupportedWiFi{} }
