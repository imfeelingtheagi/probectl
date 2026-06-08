#!/usr/bin/env bash
#
# apply_branch_protection.sh (TEST-002, SUPPLY-007): import/enforce the
# committed ruleset (.github/rulesets/main.json) on the repository so that the
# verify-branch-protection CI gate goes green.
#
# WHY THIS IS A SEPARATE, HUMAN-RUN STEP: the check (scripts/check_branch_protection.sh)
# only needs `contents: read` and runs in CI on every push. CREATING/UPDATING a
# ruleset needs an ADMIN-scoped token (the default GITHUB_TOKEN cannot write
# rulesets), so it is deliberately NOT wired into CI — a repo admin runs this
# once (or re-runs it after editing the committed ruleset). It is the only
# action that flips verify-branch-protection from red to green; nothing in the
# codebase can self-apply it.
#
# Usage:
#   GITHUB_REPOSITORY=imfeelingtheagi/probectl \
#   GITHUB_TOKEN=<admin PAT with `administration: write`> \
#     bash scripts/apply_branch_protection.sh
#
# Equivalent UI path (no token needed): Settings -> Rules -> Rulesets ->
#   New ruleset -> Import -> upload .github/rulesets/main.json -> Enforcement
#   = Active -> Create.
#
# Dependency-light: curl + python3 (matches check_branch_protection.sh; no jq).
set -euo pipefail

REPO="${GITHUB_REPOSITORY:?set GITHUB_REPOSITORY=owner/repo}"
TOKEN="${GITHUB_TOKEN:?set GITHUB_TOKEN to an admin PAT (administration: write)}"
RULESET_FILE="${RULESET_FILE:-.github/rulesets/main.json}"
API="https://api.github.com"

test -f "$RULESET_FILE" || { echo "apply: $RULESET_FILE not found (run from the repo root)" >&2; exit 2; }

gh_api() { # METHOD URL [BODY_FILE]
  local method="$1" url="$2" body="${3:-}"
  local args=(-sS --fail-with-body -X "$method"
    -H "Authorization: Bearer ${TOKEN}"
    -H "Accept: application/vnd.github+json"
    -H "X-GitHub-Api-Version: 2022-11-28")
  [ -n "$body" ] && args+=(-H "Content-Type: application/json" --data-binary "@${body}")
  curl "${args[@]}" "$url"
}

# Find an existing repo ruleset named the same as the committed one (idempotent
# re-apply updates it in place rather than creating duplicates).
want_name="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["name"])' "$RULESET_FILE")"
echo "apply: ensuring ruleset $want_name on $REPO ..."

existing_id="$(gh_api GET "${API}/repos/${REPO}/rulesets?includes_parents=false" \
  | python3 -c "import json,sys; rs=json.load(sys.stdin); print(next((str(r['id']) for r in rs if r.get('name')==sys.argv[1]),''))" "$want_name")"

if [ -n "$existing_id" ]; then
  echo "apply: updating existing ruleset id=${existing_id}"
  gh_api PUT "${API}/repos/${REPO}/rulesets/${existing_id}" "$RULESET_FILE" >/dev/null
else
  echo "apply: creating new ruleset"
  gh_api POST "${API}/repos/${REPO}/rulesets" "$RULESET_FILE" >/dev/null
fi

echo "apply: done. Verifying it is now live (same check CI runs) ..."
GITHUB_REPOSITORY="$REPO" GITHUB_TOKEN="$TOKEN" bash scripts/check_branch_protection.sh
