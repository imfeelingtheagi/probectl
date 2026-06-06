#!/usr/bin/env bash
# check_swallowed_errors.sh — recurrence guard for U-058.
#
# Bans constructing an error and immediately discarding it (`_ = fmt.Errorf(...)`
# and the errors.New equivalent): the pattern looks like handling but drops the
# error on the floor — on an audit path that silently loses the trail. Log it
# (logging.FromContext(ctx)) or return it instead.
#
# errcheck's check-blank was evaluated and rejected: it would flag every
# legitimate `_ = x.Close()`-style discard repo-wide. This guard targets the
# construct-then-discard bug class specifically. Runs in `make lint-go`.
set -euo pipefail
cd "$(dirname "$0")/.."

hits=$(grep -rn --include='*.go' -E '^\s*_\s*=\s*(fmt\.Errorf|errors\.New)\(' \
  cmd internal ee pkg test 2>/dev/null || true)

if [[ -n "$hits" ]]; then
  echo "swallowed constructed errors (log or return them — U-058):" >&2
  echo "$hits" >&2
  exit 1
fi
echo "check_swallowed_errors: no constructed-then-discarded errors."
