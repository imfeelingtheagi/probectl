// SPDX-License-Identifier: LicenseRef-probectl-TBD

package pipeline

import (
	"testing"
	"time"
)

// CORRECT-012: a sample stamped far in the future is clamped to ingest time and
// counted; an in-window (including slightly-future and legitimately-past)
// timestamp passes through untouched.
func TestClampFutureSample(t *testing.T) {
	now := time.Now().UnixMilli()

	// Far future (1h ahead): clamped to now, counted.
	before := FutureClamped()
	if got := clampFutureSample(now+time.Hour.Milliseconds(), now); got != now {
		t.Fatalf("far-future sample not clamped: got %d, want %d", got, now)
	}
	if FutureClamped() != before+1 {
		t.Fatal("clamp was not counted")
	}

	// Small benign skew (1m, under the 5m bound): untouched.
	near := now + time.Minute.Milliseconds()
	if got := clampFutureSample(near, now); got != near {
		t.Fatalf("in-window future skew was clamped: got %d, want %d", got, near)
	}

	// Legitimately-late drain (1h in the past): never clamped — real event time.
	past := now - time.Hour.Milliseconds()
	if got := clampFutureSample(past, now); got != past {
		t.Fatalf("past sample was altered: got %d, want %d", got, past)
	}
}
