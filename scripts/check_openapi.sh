#!/usr/bin/env bash
#
# OpenAPI completeness gate (S19, F23 / CLAUDE.md §6, §8): no undocumented routes
# may ship at a GA milestone. This gate:
#   1. validates internal/control/openapi.json is well-formed JSON,
#   2. checks it is an OpenAPI 3.1 document with a non-empty paths object,
#   3. runs the Go check that the registered /v1 route table EXACTLY matches the
#      documented /v1 operations (neither undocumented handlers nor documented
#      phantom routes).
set -euo pipefail

cd "$(dirname "$0")/.."

SPEC="internal/control/openapi.json"

echo ">> openapi: validating ${SPEC}"
python3 - "$SPEC" <<'PY'
import json, sys
spec = sys.argv[1]
with open(spec) as f:
    doc = json.load(f)  # raises on malformed JSON
version = str(doc.get("openapi", ""))
if not version.startswith("3.1"):
    sys.exit(f"openapi gate: expected OpenAPI 3.1, got {version!r}")
paths = doc.get("paths") or {}
if not paths:
    sys.exit("openapi gate: no paths documented")
v1 = [p for p in paths if p.startswith("/v1/")]
if not v1:
    sys.exit("openapi gate: no /v1 operations documented")
print(f"   openapi {version}, {len(paths)} paths ({len(v1)} under /v1)")
PY

echo ">> openapi: routes <-> spec completeness"
${GO:-go} test -count=1 -run '^TestOpenAPIMatchesRoutes$' ./internal/control/

echo "openapi gate: OK (no undocumented routes)"
