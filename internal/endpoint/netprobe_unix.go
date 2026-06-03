//go:build !windows

package endpoint

import (
	"context"
	"strconv"
)

// newPlatformLastMileCollector traces the path with the system `traceroute -n`
// (numeric, so no DNS-leak of the hops; -q probes per hop; -w wait; -m max hops).
// Unprivileged traceroute uses UDP/ICMP as the OS allows; a missing binary or a
// non-zero exit degrades to "unavailable" without panicking.
func newPlatformLastMileCollector(probes, maxHops int) LastMileCollector {
	return cmdLastMileCollector{run: func(ctx context.Context, target string) (string, error) {
		return execText(ctx, "traceroute", "-n",
			"-q", strconv.Itoa(probes), "-w", "2", "-m", strconv.Itoa(maxHops), target)
	}}
}
