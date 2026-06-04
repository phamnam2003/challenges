# Kubernetes Gateway API

## What is Gateway API?

**Gateway API** is the next generation of **Ingress**, **Load Balancing**, and **Service Mesh APIs** in Kubernetes — developed officially by the SIG Network group. Rather than a single `Ingress` resource driven by vendor-specific annotations, Gateway API separates concerns across independent resources following a role-oriented model: infrastructure is managed by cluster operators, routing logic is defined by developers. Core differences from Ingress: native support for header-based matching, traffic weighting, cross-namespace routing, and gRPC/TCP routing without any custom annotations. Gateway API is a portable standard — the same manifest runs on Cilium, Istio, NGINX Gateway Fabric, or Envoy Gateway without modification.

---

## Problems It Solves

- **Vendor annotation lock-in** — Ingress requires controller-specific annotations (`nginx.ingress.kubernetes.io/...`, `haproxy.org/...`). Switching controllers means rewriting all annotations. Gateway API uses standard fields that are portable across every implementation.
- **No routing permission delegation** — Ingress is cluster-scoped, so cluster operators must manage every routing rule for all teams. Gateway API lets developers create `HTTPRoute` in their own namespace and attach it to a Gateway without needing cluster-admin rights.
- **Limited routing capabilities** — Ingress only supports path-based and hostname routing. Header matching, traffic splitting (canary 10%/90%), redirects, and rewrites all require non-standard annotations. Gateway API supports all of these natively in the spec.
- **No service mesh routing** — Ingress only handles north-south traffic (external to cluster). Gateway API has a `Mesh` profile that handles east-west traffic between services inside the cluster without separate sidecar injection.
- **Unsafe cross-namespace routing** — Ingress has no trust model for cross-namespace references. Gateway API uses `ReferenceGrant` so the backend namespace must explicitly allow routes from other namespaces to attach.

---

## Architecture

```
  GatewayClass  ──────────────────────────────────────────────────────────────
  (cluster-wide)   controllerName: gateway.envoyproxy.io/gatewayclass-controller
  ──────────────────────────────────────────────────────────────────────────────
         │
         ▼
  Gateway (Cluster Operator)
    listeners:
      - name: http,  port: 80,  protocol: HTTP
      - name: https, port: 443, protocol: HTTPS, tls: Terminate
    allowedRoutes: All namespaces
         │
         │  parentRefs
         ▼
  HTTPRoute (Application Developer NS)
    hostnames: ["app.example.com"]
    rules:
      - match: header x-env=canary  → svc-canary:8080
      - match: path prefix /api     → svc-api:8080
      - (default)                   → svc-main:80
         │
         ▼
  Service  →  Pod(s)
```

### Core Resources

| Resource | Scope | Managed by | Role |
|----------|-------|------------|------|
| `GatewayClass` | Cluster | Infrastructure Provider | Registers the controller implementation |
| `Gateway` | Namespace | Cluster Operator | Defines the traffic entry point (port, protocol, TLS) |
| `HTTPRoute` | Namespace | Application Developer | Routing rules for HTTP/HTTPS |
| `GRPCRoute` | Namespace | Application Developer | Routing rules for gRPC |
| `TLSRoute` | Namespace | Application Developer | TLS passthrough routing |
| `TCPRoute` | Namespace | Application Developer | Raw TCP routing (alpha) |
| `ReferenceGrant` | Namespace | Backend Owner | Permits cross-namespace references |

---

## Manifest Structure

```yaml
# 1. GatewayClass — cluster-scoped, created by the infrastructure provider
# Usually provisioned automatically when the controller is installed
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  # must match the controllerName the Envoy Gateway deployment advertises
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
# 2. Gateway — created by the cluster operator, defines the entry point
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: main-gateway
  namespace: gateway-infra
spec:
  gatewayClassName: eg   # references the GatewayClass above
  listeners:
  - name: http
    port: 80
    protocol: HTTP
    allowedRoutes:
      namespaces:
        # "All" lets HTTPRoutes from any namespace attach to this listener
        # Use "Same" to restrict to the Gateway's own namespace
        # Use "Selector" with labelSelector for controlled multi-tenant setups
        from: All
  - name: https
    port: 443
    protocol: HTTPS
    tls:
      # terminate TLS at the Gateway, forward plain HTTP to backends
      mode: Terminate
      certificateRefs:
      - name: example-tls-secret  # Secret of type kubernetes.io/tls in the same namespace
        kind: Secret
    allowedRoutes:
      namespaces:
        from: All
---
# 3. HTTPRoute — created by the application developer in their own namespace
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: app-route
  namespace: app-ns
spec:
  parentRefs:
  - name: main-gateway
    namespace: gateway-infra  # required when the Gateway is in a different namespace
    sectionName: https         # attach only to the "https" listener, not "http"
  hostnames:
  - "app.example.com"
  rules:
  # Rule 1: canary header routing — specific matches must come before generic ones
  - matches:
    - headers:
      - type: Exact
        name: x-env
        value: canary
    backendRefs:
    - name: app-svc-canary
      port: 8080
      weight: 1   # weight is only meaningful when multiple backendRefs share the same rule
  # Rule 2: /api path prefix → dedicated API service
  - matches:
    - path:
        type: PathPrefix
        value: /api
    backendRefs:
    - name: app-svc-api
      port: 8080
  # Rule 3: default catch-all — no matches field means it accepts all remaining traffic
  - backendRefs:
    - name: app-svc
      port: 80
---
# 4. ReferenceGrant — required when an HTTPRoute references a Service in another namespace
# The namespace that owns the Service must create this resource to allow the cross-namespace ref
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata:
  name: allow-gateway-infra
  namespace: app-ns          # the namespace WHERE the backend Service lives
spec:
  from:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    namespace: route-ns      # the namespace where the HTTPRoute lives
  to:
  - group: ""
    kind: Service
```

---

## Key Fields

### `listeners[].allowedRoutes.namespaces.from`

Controls which namespaces can attach Routes to this listener. Determines the trust model between cluster operators and developers.

| Value | Behavior |
|-------|----------|
| `Same` | Only Routes in the same namespace as the Gateway can attach |
| `All` | Any namespace can attach — suitable for a shared gateway |
| `Selector` | Only namespaces matching a `labelSelector` can attach — controlled multi-tenant setup |

### `listeners[].tls.mode`

Determines who terminates TLS.

| Value | Behavior |
|-------|----------|
| `Terminate` | Gateway decrypts TLS, forwards plain HTTP/gRPC to the backend |
| `Passthrough` | Gateway forwards TLS as-is to the backend — the backend terminates it |

### `rules[].matches`

Defines the conditions that activate a rule. A request must satisfy **all** conditions within the same `matches` entry.

| Match type | Field | Example |
|------------|-------|---------|
| Path prefix | `path.type: PathPrefix` | `/api` matches `/api`, `/api/v1`, `/api/users` |
| Exact path | `path.type: Exact` | `/health` matches only `/health` |
| Header | `headers[].type: Exact` | `x-env: canary` |
| Query param | `queryParams[].type: Exact` | `version=2` |
| HTTP method | `method: GET` | Accepts GET requests only |

### `rules[].backendRefs[].weight`

Traffic splitting for A/B testing and canary rollouts. Weights do not need to sum to 100 — the controller computes relative ratios.

```yaml
backendRefs:
- name: app-svc-v2
  port: 80
  weight: 10   # 10% of traffic
- name: app-svc-v1
  port: 80
  weight: 90   # 90% of traffic
```

---

## Common Implementations

| Implementation | Dataplane | Best for |
|----------------|-----------|----------|
| **Cilium** | eBPF | Integrated cluster networking + gateway, no sidecar needed |
| **Envoy Gateway (kgateway)** | Envoy Proxy | Standalone API gateway with rich L7 features |
| **Istio** | Envoy (sidecar mesh) | Service mesh combined with ingress gateway |
| **NGINX Gateway Fabric** | NGINX | Familiar for teams already using NGINX Ingress |
| **Traefik Proxy** | Traefik | Cloud-native with automatic service discovery |

---

## Common Pitfalls

**Missing `parentRefs.namespace`:** When an HTTPRoute is in a different namespace from the Gateway, `namespace` in `parentRefs` is required. If omitted, the controller assumes the same namespace → the Route never attaches → no traffic flows.

**Missing `ReferenceGrant`:** An HTTPRoute references a Service in another namespace but no `ReferenceGrant` exists in the Service's namespace → the controller silently rejects it; the Route shows `ResolvedRefs: False` with no obvious error message.

**Rule order not considered:** Gateway API uses "most specific match wins" — but when two rules have equal specificity, the **rule that appears first in the array wins**. Always place catch-all rules (no `matches` field) last, not first.

**`sectionName` left empty with multiple listeners:** If a Gateway has both HTTP and HTTPS listeners and an HTTPRoute omits `sectionName`, the Route attaches to both listeners. If the HTTPS listener requires a TLS cert but the Route is not configured accordingly, the listener rejects the attachment.

**`allowedRoutes.from: Same` on a shared gateway:** The cluster operator sets `from: Same` but developers create Routes in their own namespaces → every Route is rejected. Use `All` or `Selector` with an appropriate label.

**GatewayClass with no controller installed:** Creating a `GatewayClass` without installing the matching controller → the Gateway stays in `Accepted: False`; nothing processes the `controllerName`. Install the controller (Envoy Gateway, Cilium, etc.) first.

**TLS Secret in a different namespace than the Gateway:** `certificateRefs` in a TLS listener only resolves Secrets in the same namespace as the Gateway by default. A Secret in another namespace requires its own `ReferenceGrant` for the TLS reference.

---

## References

- [Gateway API Official Docs](https://gateway-api.sigs.k8s.io/)
- [API Concepts Overview](https://gateway-api.sigs.k8s.io/docs/concepts/api-overview/)
- [HTTP Routing Guide](https://gateway-api.sigs.k8s.io/guides/http-routing/)
- [Implementations List](https://gateway-api.sigs.k8s.io/implementations/)
- [NGINX Ingress Controller](../ingress/nginx/README.md)
