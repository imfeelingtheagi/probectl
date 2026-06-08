// SPDX-License-Identifier: LicenseRef-probectl-TBD

package config

import "testing"

// TENANT-106: PROBECTL_REQUIRE_AT_REST_ENCRYPTION is parsed and defaults off.
// (The fatal-on-keyless enforcement lives at the main.go boot seam, exercised
// by the boot path; here we lock the config contract.)
func TestRequireAtRestEncryptionConfig(t *testing.T) {
	env := map[string]string{
		"PROBECTL_AUTH_MODE":                  "session",
		"PROBECTL_REQUIRE_AT_REST_ENCRYPTION": "true",
	}
	cfg, err := Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !cfg.RequireAtRestEncryption {
		t.Fatal("PROBECTL_REQUIRE_AT_REST_ENCRYPTION=true must set the flag")
	}

	def, err := Load(func(string) string { return "" })
	if err != nil {
		t.Fatalf("load default: %v", err)
	}
	if def.RequireAtRestEncryption {
		t.Fatal("at-rest encryption must NOT be required by default (keyless dev)")
	}
}

// TENANT-102: the ClickHouse tenant-scoping knobs parse and default off.
func TestFlowCHScopingConfig(t *testing.T) {
	env := map[string]string{
		"PROBECTL_FLOWSTORE_TENANT_SCOPING": "true",
		"PROBECTL_FLOWSTORE_READER_USER":    "probectl_reader",
	}
	cfg, err := Load(func(k string) string { return env[k] })
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !cfg.FlowCHTenantScoping || cfg.FlowCHReaderUser != "probectl_reader" {
		t.Fatalf("scoping knobs not parsed: %+v / %q", cfg.FlowCHTenantScoping, cfg.FlowCHReaderUser)
	}
	def, _ := Load(func(string) string { return "" })
	if def.FlowCHTenantScoping {
		t.Fatal("CH tenant scoping must default off")
	}
}
