// SPDX-License-Identifier: LicenseRef-probectl-TBD

//go:build !linux

package canary

import "syscall"

// dialControl is a no-op off Linux (DSCP marking is best-effort and the agent
// targets Linux).
func dialControl(_ int) func(network, address string, c syscall.RawConn) error {
	return nil
}
