// SPDX-License-Identifier: LicenseRef-probectl-TBD

package flowstore

import (
	"math"
	"sort"
)

// detectAnomalies is the shared baseline detector: for each (exporter, iface)
// series, the last bucket is compared against the mean + k*stddev of the
// preceding buckets. Pure function — Memory and ClickHouse both feed it their
// capacity series so the two backends flag identically (the memory store is
// the reference implementation for the SQL).
func detectAnomalies(points []CapacityPoint, q AnomalyQuery) []Anomaly {
	type key struct {
		exporter string
		iface    uint32
	}
	series := make(map[key][]CapacityPoint)
	for _, p := range points {
		k := key{p.Exporter, p.Iface}
		series[k] = append(series[k], p)
	}

	var out []Anomaly
	for k, pts := range series {
		sort.Slice(pts, func(i, j int) bool { return pts[i].TS.Before(pts[j].TS) })
		// Need a baseline of at least 3 buckets plus the bucket under test.
		if len(pts) < 4 {
			continue
		}
		base := pts[:len(pts)-1]
		cur := pts[len(pts)-1]

		var sum float64
		for _, p := range base {
			sum += p.Bps
		}
		mean := sum / float64(len(base))
		var sq float64
		for _, p := range base {
			d := p.Bps - mean
			sq += d * d
		}
		std := math.Sqrt(sq / float64(len(base)))

		if cur.Bps < q.MinBps {
			continue
		}
		threshold := mean + q.Sensitivity*std
		if std == 0 {
			// A flat baseline: anything materially above it is anomalous.
			if cur.Bps <= mean*1.5 {
				continue
			}
		} else if cur.Bps <= threshold {
			continue
		}
		sigma := 0.0
		if std > 0 {
			sigma = (cur.Bps - mean) / std
		}
		out = append(out, Anomaly{
			Exporter:    k.exporter,
			Iface:       k.iface,
			TS:          cur.TS,
			CurrentBps:  cur.Bps,
			BaselineBps: mean,
			StdDevBps:   std,
			Sigma:       sigma,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Sigma > out[j].Sigma })
	return out
}
