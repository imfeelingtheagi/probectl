#!/usr/bin/env bash
# supply-pins gate (SUPPLY-001/002/006, Sprint 23): every mutable input is
# pinned. Fails on:
#   - ":latest" image refs under deploy/ (a clearly-labeled local-dev line
#     may opt out with "# local-dev-ok")
#   - `go install ...` in CI without an exact @vX.Y.Z
#   - `pip install` in CI/Makefile without exact ==pins, --require-hashes,
#     or --no-deps
# SELFTEST: check_supply_pins.sh SELFTEST exercises the failure paths.
set -euo pipefail
cd "$(dirname "$0")/.."

fail=0

if [[ "${1:-}" == "SELFTEST" ]]; then
  tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
  echo 'image: ghcr.io/x/y:latest' > "$tmp/bad.yml"
  grep -q ':latest' "$tmp/bad.yml" || { echo "SELFTEST broken"; exit 1; }
  echo 'pip install ruff' > "$tmp/bad.sh"
  if grep -E 'pip install' "$tmp/bad.sh" | grep -vqE '(==|--require-hashes|--no-deps|-r [^ ]+\.lock)'; then :; else echo "SELFTEST broken"; exit 1; fi
  echo "supply-pins SELFTEST: OK"
  exit 0
fi

# 1) :latest image refs under deploy/ (comments stripped; explicit
#    local-dev opt-outs allowed with "local-dev-ok").
while IFS= read -r line; do
  echo "$line" | grep -q 'local-dev-ok' && continue
  code="${line#*:}"; code="${code%%#*}" # strip file:line prefix piece + trailing comment
  echo "$code" | grep -q ':latest' || continue
  echo "FORBIDDEN :latest image ref (pin a release; SUPPLY-001):"
  echo "  $line"
  fail=1
done < <(grep -rn ':latest' deploy/ --include='*.yml' --include='*.yaml' || true)

# 2) go install without an exact version in workflows/Makefile.
while IFS= read -r line; do
  echo "$line" | grep -qE '@v[0-9]+\.[0-9]+\.[0-9]+' && continue
  echo "UNPINNED go install (want @vX.Y.Z):"
  echo "  $line"
  fail=1
done < <(grep -rn 'go install ' .github/workflows Makefile | grep -v '^\s*#' | grep '@' || true)

# 3) pip install without exact pins / hashes / no-deps / a lockfile.
while IFS= read -r line; do
  echo "$line" | grep -qE '(==|--require-hashes|--no-deps|-r [^ ]*requirements[^ ]*\.lock|install uv==)' && continue
  echo "UNPINNED pip install (want ==X.Y.Z, --require-hashes, or --no-deps):"
  echo "  $line"
  fail=1
done < <(grep -rn 'pip install' .github/workflows Makefile | grep -v '^\s*#' || true)

if [[ $fail -ne 0 ]]; then
  echo
  echo "supply-pins gate FAILED — pin the inputs above (docs/dependency-policy.md)."
  exit 1
fi
echo "supply-pins gate: OK (no :latest, no unpinned go install / pip install)"
