# Verifying release artifacts (U-006)

Every released binary and the `checksums.txt` manifest are signed with
**cosign keyless** (Sigstore): the release workflow signs with its GitHub
OIDC identity, Fulcio issues a short-lived certificate (the `.pem` next to
each artifact), and the signature is logged in the public Rekor transparency
log. There is **no private key to leak** — what you verify is that the
artifact was produced by *this repository's release workflow for a release
tag*, and nothing else.

Each release ships, per artifact: the artifact, `<artifact>.sig`,
`<artifact>.pem`, plus one `checksums.txt` (itself signed).

## Verify (copy-paste)

```sh
# 0. Install cosign: https://docs.sigstore.dev/cosign/system_config/installation/
TAG=v0.1.0
BIN=probectl-agent_${TAG}_linux_amd64
BASE=https://github.com/imfeelingtheagi/probectl/releases/download/${TAG}

curl -fsSLO ${BASE}/${BIN} -O ${BASE}/${BIN}.sig -O ${BASE}/${BIN}.pem \
     -O ${BASE}/checksums.txt -O ${BASE}/checksums.txt.sig -O ${BASE}/checksums.txt.pem

# 1. The signature must chain to THIS repo's release workflow on a release tag.
cosign verify-blob \
  --certificate ${BIN}.pem \
  --signature   ${BIN}.sig \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --certificate-identity-regexp \
    "^https://github.com/imfeelingtheagi/probectl/\.github/workflows/release\.yml@refs/tags/" \
  ${BIN}

# 2. Same check for the manifest, then the checksum.
cosign verify-blob \
  --certificate checksums.txt.pem \
  --signature   checksums.txt.sig \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --certificate-identity-regexp \
    "^https://github.com/imfeelingtheagi/probectl/\.github/workflows/release\.yml@refs/tags/" \
  checksums.txt
sha256sum --ignore-missing -c checksums.txt
```

Both `cosign verify-blob` calls print `Verified OK`; `sha256sum -c` prints
`<artifact>: OK`. **Anything else: do not run the binary.**

What the identity pin means: Fulcio bound the signing certificate to the
workflow `release.yml` in `imfeelingtheagi/probectl` running for a
`refs/tags/...` ref, authenticated by GitHub's OIDC issuer. A fork, a
different workflow, or a re-signed binary fails the
`--certificate-identity-regexp` match.

Container images are pushed with **SLSA provenance + SBOM attestations**
(docker buildx `provenance: true, sbom: true`); inspect with
`docker buildx imagetools inspect ghcr.io/imfeelingtheagi/probectl-control:<tag>`.

CI self-check: the release workflow runs `cosign verify-blob` on its own
artifacts after signing — a release that cannot be verified does not publish.
