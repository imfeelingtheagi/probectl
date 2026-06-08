// SPDX-License-Identifier: LicenseRef-probectl-TBD

//go:build isolation

package tenancy_test

import (
	"context"
	"strings"
	"testing"

	"github.com/imfeelingtheagi/probectl/internal/tenancy"
)

// TENANT-104: the boot self-check passes against a correctly-migrated database
// (probectl_app is non-super, non-bypassrls; every tenant table FORCEs RLS).
func TestAssertIsolationPosturePasses(t *testing.T) {
	ctx := context.Background()
	pool := setup(ctx, t)
	defer pool.Close()

	if err := tenancy.AssertIsolationPosture(ctx, pool); err != nil {
		t.Fatalf("posture check must pass on a migrated DB: %v", err)
	}
}

// The app role is provably non-bypassrls + non-super (the migration's promise,
// verified through the same role the runtime assumes).
func TestAppRoleCannotBypassRLS(t *testing.T) {
	ctx := context.Background()
	pool := setup(ctx, t)
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE "+tenancy.AppRole); err != nil {
		t.Fatalf("assume app role: %v", err)
	}
	var super, bypass bool
	if err := tx.QueryRow(ctx,
		`SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user`).Scan(&super, &bypass); err != nil {
		t.Fatalf("read role: %v", err)
	}
	if super || bypass {
		t.Fatalf("app role must be non-super non-bypassrls, got super=%t bypass=%t", super, bypass)
	}
}

// TENANT-104: the check is not vacuous — if a tenant table loses FORCE RLS,
// the posture assertion FAILS (we toggle it inside a rolled-back tx so the
// database is never left weakened).
func TestAssertIsolationPostureCatchesUnforcedRLS(t *testing.T) {
	ctx := context.Background()
	pool := setup(ctx, t)
	defer pool.Close()

	// Disable FORCE on one tenant table within a transaction, run the check on
	// that same connection, then roll back. Using a single dedicated conn keeps
	// the DDL visible to the check and the rollback total.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "ALTER TABLE tests NO FORCE ROW LEVEL SECURITY"); err != nil {
		t.Fatalf("toggle force off: %v", err)
	}
	// Assume the app role within this SAME tx (reversible at rollback) and run
	// the real check core against this connection's view — it must reject the
	// unforced table. This proves the check is not vacuous.
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE "+tenancy.AppRole); err != nil {
		t.Fatalf("assume app role: %v", err)
	}
	if err := tenancy.AssertPostureTx(ctx, tx); err == nil || !strings.Contains(err.Error(), "FORCE ROW LEVEL SECURITY") {
		t.Fatalf("posture check must reject an unforced tenant table, got %v", err)
	}
}
