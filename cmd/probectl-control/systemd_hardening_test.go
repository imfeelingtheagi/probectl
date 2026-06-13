// SPDX-License-Identifier: LicenseRef-probectl-TBD

package main

import (
	"strings"
	"testing"
)

// EBPF-002: the eBPF host agent is privileged, so its systemd confinement must
// not differ by distribution channel. The template unit (deploy/agent, shipped
// by Helm/install.sh) and the packaged unit (deploy/packaging/systemd, shipped
// by deb/rpm) must carry the SAME hardening directives — otherwise deb/rpm
// operators silently get a weaker, OOM-backstop-less profile. This guard reads
// both literal units and fails the build if the packaged one drifts below the
// hardened baseline.
func TestEBPFSystemdUnitsHardeningParity(t *testing.T) {
	template := readArtifact(t, "deploy/agent/probectl-ebpf-agent.service")
	packaged := readArtifact(t, "deploy/packaging/systemd/probectl-ebpf-agent.service")

	// Directives that MUST be present (with these values) on BOTH units. These
	// are the kernel/namespace confinement + the resource ceiling whose absence
	// was the finding.
	required := []string{
		"MemoryMax=512M",
		"MemoryHigh=384M",
		"TasksMax=128",
		"CPUQuota=100%",
		"MemoryDenyWriteExecute=yes",
		"ProtectKernelModules=yes",
		"ProtectKernelLogs=yes",
		"ProtectControlGroups=yes",
		"ProtectClock=yes",
		"RestrictNamespaces=yes",
		"RestrictRealtime=yes",
		"RestrictSUIDSGID=yes",
		"RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK",
		"ProtectSystem=strict",
		"SystemCallArchitectures=native",
		"CapabilityBoundingSet=CAP_BPF CAP_PERFMON",
	}
	for _, d := range required {
		if !strings.Contains(template, d) {
			t.Errorf("template unit missing hardening directive %q", d)
		}
		if !strings.Contains(packaged, d) {
			t.Errorf("packaged (deb/rpm) unit missing hardening directive %q — confinement drift below the hardened baseline (EBPF-002)", d)
		}
	}

	// The syscall denylist (the ~-prefixed exclusions) must appear on both so
	// neither channel allows mount/module/ptrace etc.
	denylist := "SystemCallFilter=~@mount @module @reboot @swap @obsolete @cpu-emulation @keyring ptrace"
	if !strings.Contains(packaged, denylist) {
		t.Errorf("packaged unit missing the syscall denylist %q (EBPF-002)", denylist)
	}
}
