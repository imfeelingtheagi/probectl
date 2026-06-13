// SPDX-License-Identifier: LicenseRef-probectl-TBD

// Package cipolicy holds policy tests over the CI/release workflows — the in-repo
// backstops that protect main and the release path when server-side settings
// (GitHub branch protection) cannot be asserted from the tree. It has no runtime
// code; all assertions live in the _test.go files (EXC-GATE-02 / EXC-GATE-04).
package cipolicy
