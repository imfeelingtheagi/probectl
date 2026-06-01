# deploy/helm/

Helm chart for deploying netctl on Kubernetes / OpenShift.

The chart lives in [`netctl/`](netctl/). It is **HTTPS-by-default**: the API is
exposed only through a TLS-terminating ingress that emits HSTS and force-redirects
HTTP → HTTPS; the Service is `ClusterIP`, so no plaintext API is reachable from
outside the cluster (CLAUDE.md §7 guardrail 12). The migration runs as an init
container; the pod runs non-root with a read-only root filesystem.

## Install (single-tenant / sovereign)

```sh
helm install netctl deploy/helm/netctl \
  --namespace netctl --create-namespace \
  --set ingress.host=netctl.example.com \
  --set ingress.tlsSecretName=netctl-tls \
  --set database.url='postgres://netctl:...@db:5432/netctl?sslmode=require' \
  --set secrets.envelopeKey="$(openssl rand -base64 32)" \
  --set control.authMode=session \
  --set oidc.issuer=https://idp.example.com \
  --set oidc.clientId=netctl --set oidc.clientSecret=... \
  --set oidc.redirectUrl=https://netctl.example.com/auth/callback
```

Provide the TLS material via cert-manager (add the issuer annotation in
`ingress.annotations`) or a pre-created secret named by `ingress.tlsSecretName`.

## Install (multi-tenant / provider, MSP)

```sh
helm install netctl deploy/helm/netctl \
  -f deploy/helm/netctl/values-multitenant.yaml \
  --set ingress.host=netctl.msp.example.com \
  --set ingress.tlsSecretName=netctl-msp-tls \
  --set database.url=... --set secrets.envelopeKey="$(openssl rand -base64 32)" \
  --set oidc.issuer=... --set oidc.clientId=... --set oidc.clientSecret=...
```

Tenant isolation is enforced by the control plane (pooled RLS scoping) regardless
of deployment shape; the multi-tenant values only size the runtime and spread
replicas. Per-tenant white-label and the provider console arrive with the S-T
track.

## Values

See [`netctl/values.yaml`](netctl/values.yaml) (single-tenant defaults) and
[`netctl/values-multitenant.yaml`](netctl/values-multitenant.yaml). `helm lint`
and `helm template` run in CI. Full guide: [`docs/install.md`](../../docs/install.md).
