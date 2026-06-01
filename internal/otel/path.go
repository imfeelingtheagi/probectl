package otel

import "strconv"

// Path/traceroute attribute keys (S10 / finalized S22). The destination uses the
// OTel destination.address convention; path specifics use netctl.path.*.
const (
	AttrPathTarget   = "netctl.path.target"
	AttrPathMode     = "netctl.path.mode"
	AttrPathHopCount = "netctl.path.hop_count"
	AttrPathReached  = "netctl.path.destination_reached"
)

func init() {
	for _, k := range []string{AttrPathTarget, AttrPathMode, AttrPathHopCount, AttrPathReached} {
		KnownAttributes[k] = true
	}
}

// PathSummary is the path-signal input for OTel mapping — decoupled from
// internal/path so the conventions package depends only on the gen protos.
type PathSummary struct {
	TenantID           string
	Target             string
	TargetIP           string
	Mode               string
	HopCount           int
	DestinationReached bool
}

// PathAttributes maps a discovered path to its OTel/netctl attributes.
func PathAttributes(p PathSummary) map[string]string {
	attrs := map[string]string{
		AttrTenantID:     p.TenantID,
		AttrPathHopCount: strconv.Itoa(p.HopCount),
		AttrPathReached:  strconv.FormatBool(p.DestinationReached),
	}
	put := func(k, v string) {
		if v != "" {
			attrs[k] = v
		}
	}
	put(AttrPathTarget, p.Target)
	put(AttrDestinationAddress, p.TargetIP)
	put(AttrPathMode, p.Mode)
	return attrs
}
