// SPDX-License-Identifier: LicenseRef-probectl-TBD

package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ARCH-004: the embedded UI handler serves index.html at the prefix root and
// falls back to it for unknown (client-routed) paths, so a deep link loads the
// app instead of 404ing.
func TestHandlerServesEmbeddedUI(t *testing.T) {
	h := Handler("/ui/")

	for _, path := range []string{"/ui/", "/ui/incidents/42"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "<html") {
			t.Fatalf("%s: did not serve the SPA index", path)
		}
	}

	// The committed placeholder is not a real bundle.
	if Built() {
		t.Skip("a real UI bundle is embedded (release build) — placeholder assertion N/A")
	}
}
