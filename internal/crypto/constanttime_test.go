// SPDX-License-Identifier: LicenseRef-probectl-TBD

package crypto

import "testing"

func TestConstantTimeEqual(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"secret-token", "secret-token", true},
		{"secret-token", "secret-tokeX", false},
		{"secret-token", "secret", false}, // different lengths
		{"", "", true},
		{"x", "", false},
	}
	for _, c := range cases {
		if got := ConstantTimeEqual([]byte(c.a), []byte(c.b)); got != c.want {
			t.Errorf("ConstantTimeEqual(%q,%q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
