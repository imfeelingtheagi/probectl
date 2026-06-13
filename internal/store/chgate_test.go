// SPDX-License-Identifier: LicenseRef-probectl-TBD

package store_test

import (
	"strings"
	"testing"

	"github.com/imfeelingtheagi/probectl/internal/store/chmigrate"
	"github.com/imfeelingtheagi/probectl/internal/store/ebpfstore"
	"github.com/imfeelingtheagi/probectl/internal/store/flowstore"
	"github.com/imfeelingtheagi/probectl/internal/store/otelstore"
	"github.com/imfeelingtheagi/probectl/internal/store/pathstore"
)

// liveCHMigrations is the LIVE migration set of every ClickHouse telemetry
// store — the same lists the stores apply at boot. The gate runs over THESE so
// a destructive change to any of flow/eBPF/OTLP/path reddens the build
// (SCHEMA-001 — these stores were previously outside the expand/contract gate).
func liveCHMigrations() map[string][]chmigrate.Migration {
	return map[string][]chmigrate.Migration{
		"flowstore": flowstore.CHMigrations(),
		"ebpfstore": ebpfstore.CHMigrations(),
		"otelstore": otelstore.CHMigrations(),
		"pathstore": pathstore.CHMigrations(),
	}
}

// TestClickHouseMigrationGate is the SCHEMA-001 migration-gate for the
// ClickHouse stores: every store's shipped chMigrations() must pass the
// destructive-DDL check. The known data-preserving rebuilds (flow/otel) and the
// re-discoverable-cache discard (path) carry the typed Destructive+Justification
// annotation and so pass; any UNannotated DROP/RENAME on a telemetry store fails.
func TestClickHouseMigrationGate(t *testing.T) {
	v := chmigrate.CheckAll(liveCHMigrations())
	if len(v) > 0 {
		var b strings.Builder
		for _, x := range v {
			b.WriteString("\n  " + x.String())
		}
		t.Fatalf("ClickHouse migrations break the destructive-DDL gate:%s", b.String())
	}
}

// TestClickHouseMigrationGateCatchesInjectedDestructive proves the gate is not
// vacuous: injecting an UNannotated DROP TABLE into a store's live list (as a
// future bad migration would) must be flagged. This is the acceptance test —
// "add a destructive statement to a store's chMigrations() and confirm the gate
// fails" — exercised against the real list rather than a synthetic one.
func TestClickHouseMigrationGateCatchesInjectedDestructive(t *testing.T) {
	sets := liveCHMigrations()
	flows := sets["flowstore"]
	// Append a hypothetical bad v99 that DROPs a telemetry table with no annotation.
	flows = append(flows, chmigrate.Migration{
		Version: 99, Name: "drop_flows_oops",
		Statements: []string{"DROP TABLE probectl_flows"},
	})
	sets["flowstore"] = flows

	v := chmigrate.CheckAll(sets)
	if len(v) == 0 {
		t.Fatal("gate must flag an unannotated DROP TABLE injected into flowstore's migrations")
	}
	found := false
	for _, x := range v {
		if x.Component == "flowstore" && x.Version == 99 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the injected flowstore v99 DROP to be flagged, got %v", v)
	}
}
