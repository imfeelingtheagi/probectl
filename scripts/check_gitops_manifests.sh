#!/usr/bin/env bash
# GitOps manifest gate (S35): the ArgoCD/Flux manifests must be well-formed YAML
# with an apiVersion + kind (so a `git push` reconcile can't be tripped by a typo).
# A full Argo/Flux stand-up needs a cluster (out of CI scope); this validates the
# manifests structurally.
set -euo pipefail

DIR="${1:-deploy/gitops}"

python3 - "$DIR" <<'PY'
import sys, glob, os
try:
    import yaml
except ImportError:
    sys.exit("PyYAML not available")

root = sys.argv[1]
files = sorted(glob.glob(os.path.join(root, "**", "*.yaml"), recursive=True))
if not files:
    sys.exit(f"no manifests under {root}")

errors = []
checked = 0
for f in files:
    with open(f) as fh:
        try:
            docs = list(yaml.safe_load_all(fh))
        except yaml.YAMLError as e:
            errors.append(f"{f}: invalid YAML: {e}")
            continue
    for d in docs:
        if d is None:
            continue
        checked += 1
        if "apiVersion" not in d or "kind" not in d:
            errors.append(f"{f}: a document is missing apiVersion/kind")

if errors:
    print("gitops gate: FAIL")
    for e in errors:
        print("  " + e)
    sys.exit(1)
print(f"gitops gate: OK ({checked} document(s) across {len(files)} file(s))")
PY
