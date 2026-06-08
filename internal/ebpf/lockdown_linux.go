// SPDX-License-Identifier: LicenseRef-probectl-TBD

//go:build linux

package ebpf

import "os"

// lockdownMode reads the active kernel lockdown mode, or "" when securityfs is
// not mounted / the file is absent (lockdown not built in).
func lockdownMode() string {
	data, err := os.ReadFile("/sys/kernel/security/lockdown")
	if err != nil {
		return ""
	}
	return parseLockdown(string(data))
}
