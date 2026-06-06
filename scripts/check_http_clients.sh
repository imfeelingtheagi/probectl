#!/usr/bin/env bash
# check_http_clients.sh — egress TLS-policy ratchet (U-036).
#
# Every NEW outbound HTTP client must come from the approved constructor
# (crypto.HardenedHTTPClient — TLS 1.2+, AEAD-only, verification always on) or
# build its transport on crypto.HardenedClientTLSConfig. A bare &http.Client{}
# silently bypasses the hardened policy (how the remote-write client shipped
# plain — O:WIRE-002).
#
# This is a RATCHET: the allowlist below freezes the sites that predate the
# rule (each is tracked for migration in the diligence register; several pull
# their transports from injected/hardened configs already). Adding a NEW bare
# client fails this gate — migrate an allowlisted file and DELETE its line.
set -euo pipefail
cd "$(dirname "$0")/.."

# file:reason — frozen pre-existing sites. Do not add to this list.
allow='
internal/crypto/tls.go            # the approved constructor itself
internal/canary/dns.go            # DoH probe client (SSRF-guard scope, U-002 follow-up)
internal/canary/http.go           # http canary builds its transport per-test (guarded)
internal/alert/channel.go         # webhook channel (injected in tests)
internal/opendata/peeringdb.go    # opendata fetcher (cached, TLS-verified URL)
internal/opendata/atlas.go        # opendata fetcher (cached, TLS-verified URL)
internal/otel/otlp/exporter.go    # builds on hardened TLS transport explicitly
internal/cli/cli.go               # CLI talking to the operator-chosen endpoint
internal/browser/httpdriver.go    # browser-synthetic driver (per-run jar/transport)
internal/store/pathstore/clickhouse.go  # in-cluster store client (TLS via URL)
internal/store/flowstore/clickhouse.go  # in-cluster store client (TLS via URL)
'

fail=0
while IFS=: read -r file line _; do
  base=${file#./}
  if ! grep -qF "$base" <<<"$allow"; then
    echo "UNAPPROVED bare http.Client (use crypto.HardenedHTTPClient — U-036): $base:$line" >&2
    fail=1
  fi
done < <(grep -rn --include='*.go' 'http\.Client{' cmd internal ee 2>/dev/null | grep -v '_test\.go')

if [[ $fail -ne 0 ]]; then
  exit 1
fi
echo "check_http_clients: no unapproved bare http.Client egress."
