#!/usr/bin/env bash
# SPDX-License-Identifier: LicenseRef-probectl-TBD
#
# check-test-clocks.sh (TEST-002) — fail if a _test.go file hardcodes a FUTURE
# calendar date. Such dates are time-bombs: the test passes today and silently
# starts failing the day real time crosses the literal. Tests must derive time
# from an injected clock (WithClock / a fixed base) instead. This is the
# durable guard so the time-bomb class can't creep back after the sweep.
#
# Allowlist: a file may opt out with the marker comment
#   //clocklint:allow <reason>
# on the line with the literal (for a deliberately-far-future constant).
set -euo pipefail

# Future years to flag (this + the next several). Update the floor yearly, or
# wire it to $(date +%Y) in CI for an always-current check.
PATTERN='(time\.Date\([[:space:]]*20(2[7-9]|[3-9][0-9])|"20(2[7-9]|[3-9][0-9])-[0-1][0-9]-[0-3][0-9])'

hits="$(grep -rnE "$PATTERN" --include='*_test.go' . 2>/dev/null | grep -v 'clocklint:allow' || true)"
if [ -n "$hits" ]; then
    echo "::error::TEST-002: hardcoded FUTURE dates in tests are time-bombs — use an injected clock (WithClock / fixed base), or annotate the line //clocklint:allow <reason>:"
    echo "$hits"
    exit 1
fi
echo "no hardcoded future-date time-bombs in tests"
