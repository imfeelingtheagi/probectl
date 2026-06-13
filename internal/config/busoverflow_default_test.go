// SPDX-License-Identifier: LicenseRef-probectl-TBD

package config

import "testing"

// TestBusMemoryOverflowDefaultsToDrop is the SCALE-006 config acceptance: the
// lightweight in-memory bus overflow policy now defaults to "drop"
// (bounded drop-with-loss-accounting) instead of "block", so one stuck consumer
// cannot stall ingest for every other lane by default.
func TestBusMemoryOverflowDefaultsToDrop(t *testing.T) {
	cfg, err := Load(func(string) string { return "" }) // empty env = all defaults
	if err != nil {
		t.Fatalf("Load defaults: %v", err)
	}
	if cfg.BusMemoryOverflow != "drop" {
		t.Fatalf("default PROBECTL_BUS_MEMORY_OVERFLOW = %q, want \"drop\" (SCALE-006)", cfg.BusMemoryOverflow)
	}
	// "block" must still be selectable for operators who prefer backpressure.
	cfgBlock, err := Load(func(k string) string {
		if k == "PROBECTL_BUS_MEMORY_OVERFLOW" {
			return "block"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Load block: %v", err)
	}
	if cfgBlock.BusMemoryOverflow != "block" {
		t.Fatalf("explicit block = %q, want \"block\"", cfgBlock.BusMemoryOverflow)
	}
}
