# Envoy Gateway in Kubernetes

## What is Envoy Gateway?

**Envoy Gateway** is a **Kubernetes Gateway API** implementation — developed by the Envoy project under CNCF. It acts as a control plane that automatically translates Gateway API resources (GatewayClass, Gateway, HTTPRoute...) into xDS configuration to drive **Envoy Proxy** at the data plane. Instead of writing raw Envoy xDS config (thousands of lines of YAML), Envoy Gateway lets you declare routing, TLS, rate limiting, and authentication through standard Kubernetes CRDs — portable across any Gateway API implementation.

---

## Problems It Solves

- **Envoy config is too complex:** Envoy Proxy is powerful but xDS configuration (Listener, Route, Cluster, Endpoint) is extremely verbose and hard to maintain. Envoy Gateway abstracts this entire layer — users only write Gateway API resources.
- **No standard policy management:** The Gateway API spec only covers basic routing. Production features like rate limiting, JWT auth, CORS, and circuit breaking require vendor-specific annotations or custom CRDs. Envoy Gateway provides extension CRDs (`SecurityPolicy`, `BackendTrafficPolicy`, `ClientTrafficPolicy`) following the standard **Policy Attachment** model.
- **No standalone ingress gateway:** Istio provides an Envoy-based gateway but pulls in the entire service mesh. Envoy Gateway runs standalone — ingress only, no sidecars, no mesh overhead.
- **Manual lifecycle management:** With raw Envoy, operators must manually manage Deployments, Services, and HPAs for Envoy Proxy. Envoy Gateway automatically provisions and updates data plane infrastructure when Gateway resources change.

---

## How It Works

### Control Plane / Data Plane Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Control Plane (envoy-gateway pod)                          │
│                                                             │
│  Kubernetes Provider ──► Gateway API Translator ──► xDS IR  │
│  (watch GatewayClass,    (convert resources into     │      │
│   Gateway, HTTPRoute,     Intermediate Representation)│      │
│   *Policy CRDs)                                      │      │
│                                    ┌─────────────────┘      │
│                                    ▼                        │
│                            xDS Translator                   │
│                            (IR → xDS config)                │
│                                    │                        │
│                                    ▼                        │
│  Infra Manager ◄── Infra IR    xDS Server (gRPC)           │
│  (provision Envoy               (serve config via           │
│   Deployment + Service)          Delta xDS protocol)        │
└───────────┬──────────────────────────┬──────────────────────┘
            │                          │
            ▼                          ▼
┌──────────────────────────────────────────────────────────────┐
│  Data Plane (envoy proxy pod[s])                             │
│                                                              │
│  Envoy Proxy ◄──── xDS stream ────  receives config from    │
│  (handle traffic: routing, TLS,      control plane, applies  │
│   rate limit, auth enforcement)      in realtime)            │
└──────────────────────────────────────────────────────────────┘
```

### Request Flow

1. User creates Gateway API resources + extension policies in Kubernetes
2. **Kubernetes Provider** watches resources, passes them to the **Gateway API Translator**
3. Translator produces two types of IR:
   - **Infra IR** — describes required infrastructure (Deployment, Service for Envoy Proxy)
   - **xDS IR** — describes the desired data plane configuration
4. **Infra Manager** reads Infra IR → provisions/updates Envoy Proxy pods
5. **xDS Translator** converts xDS IR → xDS resources (Listener, Route, Cluster, Endpoint)
6. **xDS Server** pushes config to Envoy Proxy via gRPC stream (Delta xDS)
7. Envoy Proxy applies config and handles traffic — zero downtime on config changes

### Deployment Modes

| Mode | Data plane location | Isolation | Use case |
|------|-------------------|-----------|----------|
| **Controller Namespace** (default) | Same namespace as control plane | Low — simple for single-tenant | Dev, staging, single-team clusters |
| **Gateway Namespace** | Each Gateway's own namespace | High — each tenant gets its own Envoy | Production multi-tenant |

---

## Installation

### Helm (recommended)

```bash
# install Envoy Gateway — CRDs are in a sub-chart to avoid exceeding the 1MB Helm secret limit
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.8.1 \
  -n envoy-gateway-system \
  --create-namespace

# wait for control plane readiness
kubectl wait --timeout=5m -n envoy-gateway-system \
  deployment/envoy-gateway --for=condition=Available
```

### Quickstart resources

```bash
# creates GatewayClass + Gateway + demo HTTPRoute + backend app
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/v1.8.1/quickstart.yaml
```

### Common Helm customizations

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.8.1 \
  -n envoy-gateway-system --create-namespace \
  --set deployment.replicas=2 \                                           # HA for control plane
  --set config.envoyGateway.extensionApis.enableBackend=true \            # enable Backend CRD for external backends
  --set config.envoyGateway.rateLimit.backend.type=Redis \                # enable global rate limiting
  --set config.envoyGateway.rateLimit.backend.redis.url="redis.redis-system.svc.cluster.local:6379"
```

---

## Manifest Structure

### Standard Gateway API Resources

```yaml
# GatewayClass — cluster-scoped, typically auto-created by Helm when installing Envoy Gateway
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  # must match the controllerName advertised by Envoy Gateway
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
# Gateway — created by cluster operator, defines the traffic entry point
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: main-gateway
  namespace: gateway-infra
spec:
  gatewayClassName: eg
  listeners:
  - name: http
    port: 80
    protocol: HTTP
    allowedRoutes:
      namespaces:
        from: All   # allow HTTPRoutes from any namespace to attach
  - name: https
    port: 443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
      - name: wildcard-tls      # Secret of type kubernetes.io/tls, same namespace as Gateway
        kind: Secret
    allowedRoutes:
      namespaces:
        from: Selector
        selector:
          matchLabels:
            gateway-access: "true"   # only namespaces with this label can attach
---
# HTTPRoute — created by developers in their team namespace
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: app-route
  namespace: app-team
spec:
  parentRefs:
  - name: main-gateway
    namespace: gateway-infra     # required when Gateway is in a different namespace
    sectionName: https           # attach only to the "https" listener
  hostnames:
  - "app.example.com"
  rules:
  # canary: header match — specific rules before generic ones
  - matches:
    - headers:
      - type: Exact
        name: x-env
        value: canary
    backendRefs:
    - name: app-svc-canary
      port: 8080
  # weighted traffic split — 90/10 between stable and canary
  - backendRefs:
    - name: app-svc
      port: 80
      weight: 90
    - name: app-svc-canary
      port: 8080
      weight: 10
```

### Extension Policies (Envoy Gateway specific)

```yaml
# BackendTrafficPolicy — rate limiting for a specific route
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: rate-limit
  namespace: app-team
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: app-route
  rateLimit:
    type: Global           # requires Redis backend — shares counters across all Envoy instances
    global:
      rules:
      - clientSelectors:
        - headers:
          - name: x-user-id
            type: Distinct   # each x-user-id value gets its own bucket
        limit:
          requests: 100
          unit: Minute
---
# SecurityPolicy — JWT authentication + CORS
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: jwt-auth
  namespace: app-team
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: app-route
  jwt:
    providers:
    - name: auth0
      remoteJWKS:
        uri: https://your-tenant.auth0.com/.well-known/jwks.json
  cors:
    allowOrigins:
    - exact: "https://app.example.com"
    allowMethods:
    - GET
    - POST
    - OPTIONS       # required — missing OPTIONS causes CORS preflight to fail
    allowHeaders:
    - Authorization
    - Content-Type
---
# ClientTrafficPolicy — customize client-facing behavior
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: client-settings
  namespace: gateway-infra
spec:
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: main-gateway
  http3: {}                  # enable HTTP/3 (QUIC) on listeners
  timeout:
    http:
      requestReceivedTimeout: 30s   # timeout for receiving the full request from client
  tcpKeepalive:
    idleTime: 1200           # keep TCP connections alive - prevents the upstream nginx/firewall from closing idle connections
    probes: 3
---
# BackendTLSPolicy — TLS from Gateway to backend
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: BackendTLSPolicy
metadata:
  name: backend-mtls
  namespace: app-team
spec:
  targetRefs:
  - group: ""
    kind: Service
    name: secure-backend
    sectionName: https       # applies only to the "https" port of the Service
  validation:
    caCertificateRefs:
    - name: backend-ca       # ConfigMap containing the CA cert to verify the backend
      group: ""
      kind: ConfigMap
    hostname: secure-backend.app-team.svc.cluster.local
```

---

## Envoy Gateway Extension CRDs

Envoy Gateway extends the Gateway API spec with its own CRDs following the **Policy Attachment** model — attached to Gateways or Routes via `targetRefs`.

| CRD | apiGroup | Purpose |
|-----|----------|---------|
| `BackendTrafficPolicy` | `gateway.envoyproxy.io/v1alpha1` | Rate limiting, circuit breaking, retry, timeout, load balancing |
| `ClientTrafficPolicy` | `gateway.envoyproxy.io/v1alpha1` | HTTP version, connection timeout, TCP keepalive, HTTP/3 |
| `SecurityPolicy` | `gateway.envoyproxy.io/v1alpha1` | JWT, OIDC, basic auth, API key, ext-auth, CORS, RBAC |
| `EnvoyPatchPolicy` | `gateway.envoyproxy.io/v1alpha1` | Direct xDS config patching — must be explicitly enabled |
| `EnvoyExtensionPolicy` | `gateway.envoyproxy.io/v1alpha1` | Wasm, Lua scripting, external processing |
| `EnvoyProxy` | `gateway.envoyproxy.io/v1alpha1` | Customize Envoy Proxy deployment (resources, replicas, bootstrap) |
| `Backend` | `gateway.envoyproxy.io/v1alpha1` | External backend (not a K8s Service) — must be explicitly enabled |

### Policy Precedence

When multiple policies target the same resource, precedence order (highest to lowest):

1. **Route rule-level** — policy targets HTTPRoute + `sectionName` (specific rule)
2. **Route-level** — policy targets HTTPRoute (no `sectionName`)
3. **Listener-level** — policy targets Gateway + `sectionName` (specific listener)
4. **Gateway-level** — policy targets Gateway (no `sectionName`)

Policy merging (`mergeType: StrategicMerge`) allows platform teams to set gateway-level defaults while app teams override at the route level.

---

## Common Pitfalls

**Regex complexity causes 404 on all routes:** Envoy has a very low default `re2.max_program_size.error_level`. Complex regexes in path matching cause all routes to return 404 after Envoy restarts. Fix: increase the limit via `EnvoyProxy` resource bootstrap config. Prefer `PathPrefix` or `Exact` over regex matching.

**CORS fails due to missing OPTIONS method:** When an HTTPRoute matches specific HTTP methods (GET, POST), CORS preflight requests (OPTIONS) do not match and get rejected. The route must always include `OPTIONS` in its method match or use a separate rule for OPTIONS.

**Rate limit is per-route, not per-gateway:** Even when `BackendTrafficPolicy` targets a Gateway, each route gets its own rate limit bucket. For a true global limit across the gateway, design counter keys carefully (e.g., use `Distinct` on source IP header).

**`0s` timeout means infinite:** Starting from v1.8, setting timeout to `0s` means **no timeout** (previously it meant immediate timeout). This is a breaking change — review all timeout configs when upgrading from older versions.

**Global rate limiting requires Redis:** `rateLimit.type: Global` requires a Redis backend configured at install time. If Redis is not available, rate limit policies are silently ignored — traffic passes through without any limiting, with no clear error.

**TLS Secret must be in the same namespace as the Gateway:** `certificateRefs` in a TLS listener only resolves Secrets in the same namespace. Referencing a Secret in another namespace requires a `ReferenceGrant` in the Secret's namespace.

**EnvoyPatchPolicy is not enabled by default:** This is an escape hatch for directly patching xDS config — it must be enabled via Helm values or ConfigMap. Use with caution as it bypasses Envoy Gateway's validation layer.

---

## Comparison with Other Implementations

| Aspect | Envoy Gateway | Istio | Cilium | Traefik |
|--------|---------------|-------|--------|---------|
| **Data plane** | Envoy Proxy | Envoy Proxy (sidecar) | eBPF (L4) + Envoy (L7) | Go-native |
| **Service mesh** | No — ingress only | Yes — full mesh | Yes — via eBPF | No |
| **L7 features** | Very strong (native Envoy) | Very strong | Good | Basic |
| **Rate limiting** | Native (BackendTrafficPolicy) | Requires separate config | Basic | Limited |
| **Auth (JWT/OIDC)** | Native (SecurityPolicy) | Native | Basic | Limited |
| **Complexity** | Medium | High | Medium–High | Low |
| **Best for** | Powerful ingress gateway without mesh | Full mesh + security | Already using Cilium as CNI | Small teams, quick setup |

**When to choose Envoy Gateway:** You need production-grade Envoy features (rate limiting, auth, traffic shaping) through a Kubernetes-native API without service mesh overhead. Especially suited when your team has committed to the Gateway API standard and needs stronger L7 features than what NGINX Gateway Fabric or Traefik provide.

---

## References

- [Envoy Gateway - Official Docs](https://gateway.envoyproxy.io/docs/)
- [Envoy Gateway - System Design](https://gateway.envoyproxy.io/contributions/design/system-design/)
- [Envoy Gateway - Install with Helm](https://gateway.envoyproxy.io/docs/install/install-helm/)
- [Envoy Gateway - SecurityPolicy](https://gateway.envoyproxy.io/docs/concepts/gateway_api_extensions/security-policy/)
- [Envoy Gateway - Global Rate Limit](https://gateway.envoyproxy.io/docs/tasks/traffic/global-rate-limit/)
- [Envoy Gateway - TLS Termination](https://gateway.envoyproxy.io/docs/tasks/security/tls-termination/)
- [Envoy Gateway - JWT Authentication](https://gateway.envoyproxy.io/docs/tasks/security/jwt-authentication/)
- [Envoy Gateway GitHub](https://github.com/envoyproxy/gateway)
- [Kubernetes Gateway API](../README.md)
