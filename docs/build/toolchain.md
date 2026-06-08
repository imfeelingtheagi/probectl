# Go toolchain provenance (SUPPLY-005)

The build toolchain is the **official upstream Go release** named by the
`go` directive in `go.mod` (currently `go1.26.4`; when the directive names the exact patch, a separate `toolchain` line is redundant and `go mod tidy` folds it in):

- **Acquisition:** the Go toolchain downloads it from the canonical module
  mirror exactly like any module, and verifies it against the **public Go
  checksum database (sum.golang.org)** before first use — a tampered or
  substituted toolchain fails verification and refuses to run.
- **Pinning:** the `go 1.26.4` directive in go.mod pins the exact patch release for
  every developer and CI runner; CI's `setup-go` uses the same version, so
  there is one toolchain, everywhere, by construction.
- **Why this patch level:** pinned forward to pick up upstream **stdlib CVE
  fixes** flagged by `govulncheck` (see the pinned commit history); bumps
  land via the normal PR + green-CI path.
- **No vendored or forked toolchain exists in this repository** — provenance
  is upstream-official + sumdb-verified, full stop.
