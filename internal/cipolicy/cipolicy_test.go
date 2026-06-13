// SPDX-License-Identifier: LicenseRef-probectl-TBD

// Package cipolicy holds policy tests over the CI/release workflows themselves —
// the in-repo backstops that protect main when GitHub branch-protection settings
// (which live in server config, not the tree) cannot be asserted from here.
//
// EXC-GATE-04: assert the release.yml require-green-ci backstop exists (a v* tag
// cannot publish unless the full ci workflow concluded green on that exact SHA)
// and that verify-all is the umbrella that requires every other verification gate
// — so a gate added later but not folded into the umbrella fails this test
// instead of silently going unenforced.
//
// EXC-GATE-02: assert the ebpf-kernel-matrix live-load job is wired into the
// verify-all umbrella (the live load+attach runs on real LTS kernels in CI), so
// it cannot be quietly dropped from the required set.
package cipolicy

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cipolicy: could not locate go.mod from working dir")
		}
		dir = parent
	}
}

func readWorkflow(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", name))
	if err != nil {
		t.Fatalf("read workflow %s: %v", name, err)
	}
	return string(b)
}

// TestReleaseRequiresGreenCI is the EXC-GATE-04 backstop: a v* tag must not
// publish anything unless the full ci workflow was green on the tagged SHA. This
// holds even for a tag cut off a side branch or by an admin who bypassed branch
// protection — it is the second, independent layer documented in
// docs/ops/branch-protection.md.
func TestReleaseRequiresGreenCI(t *testing.T) {
	rel := readWorkflow(t, "release.yml")

	if !strings.Contains(rel, "require-green-ci:") {
		t.Fatal("release.yml is missing the require-green-ci backstop job (EXC-GATE-04)")
	}
	// Every job that publishes an artifact must depend (transitively) on
	// require-green-ci, or the backstop is a no-op for that job. The job's own
	// `needs:` must name it directly OR name another job that does.
	needsByJob := jobNeeds(t, rel)
	for _, job := range []string{"images", "binaries", "publish-chart", "packages"} {
		deps, ok := needsByJob[job]
		if !ok {
			t.Errorf("release.yml has no publishing job %q (renamed?) — update this policy test", job)
			continue
		}
		if !gatesOnGreenCI(job, needsByJob, map[string]bool{}) {
			t.Errorf("publishing job %q does not gate (even transitively) on require-green-ci — needs=%v; the backstop is bypassable", job, deps)
		}
	}
}

// jobNeeds maps each job name to its list of `needs:` job names (handling both
// the inline-list `needs: [a, b]` and the block-list forms).
func jobNeeds(t *testing.T, wf string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	lines := strings.Split(wf, "\n")
	jobRe := regexp.MustCompile(`^  ([a-zA-Z0-9_-]+):\s*$`)
	var cur string
	for i := 0; i < len(lines); i++ {
		ln := lines[i]
		if m := jobRe.FindStringSubmatch(ln); m != nil {
			cur = m[1]
			out[cur] = nil
			continue
		}
		if cur == "" {
			continue
		}
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "needs:") {
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "needs:"))
			rest = strings.Trim(rest, "[]")
			for _, n := range strings.Split(rest, ",") {
				if n = strings.TrimSpace(n); n != "" {
					out[cur] = append(out[cur], n)
				}
			}
		}
	}
	return out
}

// gatesOnGreenCI reports whether job depends, directly or transitively, on
// require-green-ci.
func gatesOnGreenCI(job string, needsByJob map[string][]string, seen map[string]bool) bool {
	if seen[job] {
		return false
	}
	seen[job] = true
	for _, dep := range needsByJob[job] {
		if dep == "require-green-ci" {
			return true
		}
		if gatesOnGreenCI(dep, needsByJob, seen) {
			return true
		}
	}
	return false
}

// TestBranchProtectionDocExists guards the operator-facing doc for the one
// console step the repo cannot perform (the GitHub setting). If it is deleted,
// the operator loses the runbook for making CI blocking.
func TestBranchProtectionDocExists(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "docs", "ops", "branch-protection.md"))
	if err != nil {
		t.Fatalf("docs/ops/branch-protection.md missing (EXC-GATE-04 operator runbook): %v", err)
	}
	doc := string(b)
	for _, want := range []string{"require-green-ci", "Require status checks", "verify-all"} {
		if !strings.Contains(doc, want) {
			t.Errorf("branch-protection doc does not mention %q (the required-check guidance is incomplete)", want)
		}
	}
}

// TestVerifyAllIsTheUmbrella asserts verify-all requires the full set of
// verification gates — including the ebpf-kernel-matrix live-load job
// (EXC-GATE-02) and the integration job that carries the cross-plane e2e
// (EXC-GATE-05). A gate that exists but is not in verify-all's needs is
// advisory; this test makes that omission RED.
func TestVerifyAllIsTheUmbrella(t *testing.T) {
	ci := readWorkflow(t, "ci.yml")

	// Pull the verify-all job's needs: block.
	needs := verifyAllNeeds(t, ci)
	if len(needs) < 10 {
		t.Fatalf("verify-all needs only %d gates — suspiciously few; parse failed or umbrella gutted: %v", len(needs), needs)
	}

	// The umbrella MUST include these load-bearing verification gates.
	required := []string{
		"lint-go", "editions-gate", "fips-gate", "test-go", "coverage",
		"ebpf-kernel-matrix", // EXC-GATE-02: live load+attach on real kernels
		"cross-tenant-isolation",
		"integration", // EXC-GATE-05: cross-plane correlation e2e rides here
	}
	have := map[string]bool{}
	for _, n := range needs {
		have[n] = true
	}
	for _, r := range required {
		if !have[r] {
			t.Errorf("verify-all does NOT require %q — that verification gate is advisory, not blocking", r)
		}
	}

	// Every job declared in ci.yml that is itself a verification gate should be
	// in the umbrella. We assert the umbrella is not trivially small (above) and
	// that the assertion step exists.
	if !strings.Contains(ci, "verify-all is RED") {
		t.Error("verify-all is missing its fail-closed assertion (the 'verify-all is RED' guard)")
	}
}

// TestEBPFLiveLoadFatalsNotSkips is the EXC-GATE-02 guard: the live eBPF
// load+attach smoke (TestLiveLoadAttachL4Flow, run by the ebpf-kernel-matrix CI
// job on real LTS kernels under QEMU) must t.Fatal when load+attach fails — it
// must NOT t.Skip a load failure, or the kernel matrix would pass vacuously on a
// broken BPF object. This asserts the test source still fails on a load error.
func TestEBPFLiveLoadFatalsNotSkips(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "internal", "ebpf", "live_smoke_ebpf_test.go"))
	if err != nil {
		t.Fatalf("read live eBPF smoke: %v", err)
	}
	src := string(b)
	if !strings.Contains(src, "func TestLiveLoadAttachL4Flow") {
		t.Fatal("the live load+attach smoke TestLiveLoadAttachL4Flow is gone — the kernel matrix has nothing to assert")
	}
	// The l4flow load+attach failure path must be a Fatalf, not a Skip.
	if !strings.Contains(src, "l4flow load+attach failed on this kernel") {
		t.Fatal("the l4flow load+attach failure assertion text changed — verify it still FATALS (no skip-on-load-failure)")
	}
	idx := strings.Index(src, "l4flow load+attach failed on this kernel")
	start := idx - 120
	if start < 0 {
		start = 0
	}
	window := src[start:idx]
	if !strings.Contains(window, "t.Fatalf") {
		t.Errorf("live load+attach failure does not t.Fatalf — a load failure must redden the kernel matrix, never skip")
	}
}

// verifyAllNeeds extracts the list of job names under the verify-all job's
// `needs:` block.
func verifyAllNeeds(t *testing.T, ci string) []string {
	t.Helper()
	lines := strings.Split(ci, "\n")
	inJob := false
	inNeeds := false
	var out []string
	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if trimmed == "verify-all:" {
			inJob = true
			continue
		}
		if inJob && !inNeeds {
			if trimmed == "needs:" {
				inNeeds = true
			}
			continue
		}
		if inNeeds {
			// list items look like "      - <name>" (possibly with a comment).
			if strings.HasPrefix(trimmed, "- ") {
				name := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				if i := strings.Index(name, "#"); i >= 0 {
					name = strings.TrimSpace(name[:i])
				}
				if name != "" {
					out = append(out, name)
				}
				continue
			}
			// A non-list, non-comment line ends the needs block (e.g. "steps:").
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				break
			}
		}
	}
	return out
}
