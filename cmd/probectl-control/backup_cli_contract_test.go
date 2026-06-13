// SPDX-License-Identifier: LicenseRef-probectl-TBD

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// OPS-001 / RESIL-001 / OPS-003: the shipped restore Job and the documented PITR
// recipes invoke `probectl-control backup-seal` / `backup-open` as STDIN→STDOUT
// filters. backup.go defines ONLY -key-file/-key-id (flag.ContinueOnError), so a
// recipe that passes --in/--out is a hard break: the binary errors with "flag
// provided but not defined". These guards read the literal artifacts an operator
// runs and fail the build if any recipe drifts back to a flag the CLI lacks —
// the original defect that hid because no test exercised the encrypted path.

// repoRoot walks up from the test's CWD to the module root (the dir with go.mod).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("could not locate repo root (go.mod)")
	return ""
}

func readArtifact(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(b)
}

// backupBadFlags are the flags backup.go does NOT define. Their appearance next
// to a backup-seal/backup-open invocation is the OPS-001/003 defect.
var backupBadFlags = []string{"--in ", "--in=", "--out ", "--out=", "--in %p", "--out %p", "--in -", "--out -"}

// stripComments drops shell/YAML/markdown comment lines (a leading '#', possibly
// indented) so that an explanatory note like "it has NO --in/--out flags" does
// not trip the guard — only EXECUTED command lines are checked.
func stripComments(body string) string {
	var b strings.Builder
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func assertNoBadBackupFlags(t *testing.T, name, body string) {
	t.Helper()
	if !strings.Contains(body, "backup-seal") && !strings.Contains(body, "backup-open") {
		t.Fatalf("%s: expected a backup-seal/backup-open invocation to guard, found none (did the file move?)", name)
	}
	executable := stripComments(body)
	for _, bad := range backupBadFlags {
		if strings.Contains(executable, bad) {
			t.Errorf("%s: contains %q — backup.go defines NO such flag; it reads stdin/stdout. "+
				"Use shell redirection (< in, > out) and the KEK env/-key-file instead.", name, bad)
		}
	}
}

func TestRestoreJobInvokesBackupOpenAsRealCLI(t *testing.T) {
	body := readArtifact(t, "deploy/helm/probectl/templates/restore-job.yaml")
	assertNoBadBackupFlags(t, "restore-job.yaml", body)

	// The control-plane image installs the binary at /usr/local/bin/app
	// (deploy/docker/Dockerfile), NOT /usr/bin/probectl-control.
	if strings.Contains(body, "/usr/bin/probectl-control") {
		t.Error("restore-job.yaml stages /usr/bin/probectl-control, but the image only ships /usr/local/bin/app")
	}
	if !strings.Contains(body, "/usr/local/bin/app") {
		t.Error("restore-job.yaml must copy the binary from /usr/local/bin/app (the real Dockerfile path)")
	}
	// backup-open must read the sealed file on stdin (shell redirection).
	if !strings.Contains(body, "backup-open") || !strings.Contains(body, "< \"/backups/") {
		t.Error("restore-job.yaml must pipe the sealed backup into backup-open via stdin redirection")
	}
}

func TestPITRRecipesUseRealCLI(t *testing.T) {
	body := readArtifact(t, "docs/ops/backup-restore.md")
	assertNoBadBackupFlags(t, "docs/ops/backup-restore.md", body)
}

// TestBackupFlagSetIsStdinStdoutOnly pins the CLI contract the recipes depend on:
// backup-seal/open accept only -key-file/-key-id and otherwise read stdin/write
// stdout. If someone ADDS --in/--out to backup.go in future, this stays green —
// but the guards above would then need updating, which is the intended coupling.
func TestBackupFlagSetIsStdinStdoutOnly(t *testing.T) {
	src := readArtifact(t, "cmd/probectl-control/backup.go")
	for _, want := range []string{`fs.String("key-file"`, `fs.String("key-id"`} {
		if !strings.Contains(src, want) {
			t.Errorf("backup.go: expected flag definition %q (CLI contract changed?)", want)
		}
	}
	for _, forbidden := range []string{`fs.String("in"`, `fs.String("out"`} {
		if strings.Contains(src, forbidden) {
			t.Errorf("backup.go: defines %q — update the restore Job / PITR recipe guards to match", forbidden)
		}
	}
}
