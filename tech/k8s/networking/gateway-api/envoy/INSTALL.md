# Installing Envoy Gateway

This document explains the `envoy.sh` install script, and why it installs CRDs separately
from the main chart instead of using a single `helm install` command.

---

## Why not a single `helm install`?

The README shows the quick one-command Helm install. That works in dev, but runs into how
Helm manages CRDs:

- **Helm cannot update CRDs placed in the chart's `/crds` folder.** Helm installs `/crds`
  CRDs exactly once and never touches them again, so the CRD schema gets stuck at the old
  version after the controller is upgraded.
- **`helm install --skip-crds` skips ALL CRDs,** including Envoy's own. Neither option in the
  main chart manages CRDs correctly.

So the script splits the process into 3 steps: install CRDs separately via server-side apply,
install the control plane via Helm with `--skip-crds`, then verify. The CRD and control plane
lifecycles stay decoupled.

---

## Process

```
Step 1  Install CRDs    gateway-crds-helm  --> kubectl apply --server-side
Step 2  Control plane   gateway-helm        --> helm install --skip-crds
Step 3  Verify          wait for deployment Ready + print status
```

### Step 1 - Install CRDs via `gateway-crds-helm`

The officially recommended approach: render the dedicated `gateway-crds-helm` chart and pipe
it through `kubectl apply --server-side`. Server-side apply is idempotent (running it
repeatedly yields the same result) and avoids the client-side apply size limit on large CRDs.

```bash
helm template eg-crds oci://docker.io/envoyproxy/gateway-crds-helm \
  --version v1.8.1 \
  --set crds.gatewayAPI.enabled=true \
  --set crds.gatewayAPI.channel=standard \
  --set crds.envoyGateway.enabled=true \
  | kubectl apply --server-side -f -
```

- `crds.gatewayAPI.channel` selects the Gateway API CRD bundle: `standard` (stable) or
  `experimental` (adds newer TCPRoute, TLSRoute, BackendTLSPolicy...).
- After apply, wait for the CRDs to reach `Established` before installing the controller.
  Otherwise the controller may start before the API server recognizes the CRDs.

```bash
kubectl wait --for=condition=Established --timeout=60s \
  crd/gateways.gateway.networking.k8s.io \
  crd/httproutes.gateway.networking.k8s.io \
  crd/gatewayclasses.gateway.networking.k8s.io \
  crd/envoyproxies.gateway.envoyproxy.io \
  crd/envoypatchpolicies.gateway.envoyproxy.io
```

### Step 2 - Install `gateway-helm` with `--skip-crds`

CRDs were already applied in Step 1, so `--skip-crds` prevents Helm from re-applying them.

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.8.1 \
  --namespace envoy-gateway-system \
  --create-namespace \
  --skip-crds
```

### Step 3 - Verify

```bash
# wait for the control plane to be ready (up to 5 minutes)
kubectl wait --timeout=5m -n envoy-gateway-system \
  deployment/envoy-gateway --for=condition=Available

# check release and pod status
helm status eg -n envoy-gateway-system
kubectl get pods -n envoy-gateway-system -l app.kubernetes.io/name=envoy-gateway
```

---

## Using the script

```bash
# default install
./envoy.sh

# customize version / namespace / CRD channel
./envoy.sh --version v1.8.1 --namespace envoy-gateway-system \
  --release eg --gateway-api-channel experimental
```

### Parameters

| Flag                    | Environment variable  | Default                 | Meaning                                        |
|-------------------------|-----------------------|-------------------------|------------------------------------------------|
| `--version`             | `EG_VERSION`          | `v1.8.1`                | Envoy Gateway chart version                    |
| `--namespace`           | `EG_NAMESPACE`        | `envoy-gateway-system`  | Install namespace                              |
| `--release`             | `EG_RELEASE`          | `eg`                    | Helm release name                              |
| `--gateway-api-channel` | `GATEWAY_API_CHANNEL` | `standard`              | Gateway API CRD channel: `standard`/`experimental` |

Parameters can be set via environment variables instead of flags:

```bash
EG_VERSION=v1.8.1 GATEWAY_API_CHANNEL=experimental ./envoy.sh
```

---

## Notes

- **The script uses `set -euo pipefail` and `trap ... ERR`:** any failing command stops the
  script immediately and prints the failing line number. No step runs on top of an error state.
- **Line endings on Windows:** if you edit the file on Windows, run
  `sed -i 's/\r$//' envoy.sh` to strip `\r` before running it (noted at the top of the script).
- **`--gateway-api-channel` must match your needs:** use `experimental` when you need the
  unstable CRDs (TCPRoute, TLSRoute...); otherwise keep `standard`.

---

## References

- [Envoy Gateway overview](./README.md)
- [Envoy Gateway - Install with Helm](https://gateway.envoyproxy.io/docs/install/install-helm/)
- [Gateway API - CRD Channels](https://gateway-api.sigs.k8s.io/concepts/versioning/)
