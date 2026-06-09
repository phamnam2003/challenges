# CNI (Container Network Interface) in Kubernetes

## What is CNI?

**CNI (Container Network Interface)** is a specification and set of libraries that standardize how network interfaces are configured for Linux containers — developed under the **CNCF**. CNI does not provide networking itself; it defines the *protocol* for container runtimes (containerd, CRI-O) to invoke network plugins that attach Pods to a network, assign IPs, and set up routing. Every Kubernetes cluster must have at least one CNI plugin installed for Pods to communicate with each other.

---

## Problems It Solves

- **No networking standard across runtimes:** Before CNI, each container runtime implemented its own networking — no cross-compatibility, making it difficult to migrate workloads between platforms.
- **Tight coupling between runtime and network:** Networking logic was embedded in the runtime, so changing the network model required changing the runtime itself.
- **No extensibility mechanism:** Multiple plugins (IPAM, firewall, bandwidth) could not be chained in sequence — all logic had to be packed into a single binary.
- **No lifecycle management:** No standard protocol for creating, checking, and cleaning up network resources when Pods appear or get deleted.

CNI solves these by decoupling networking into an independent plugin layer — the runtime only needs to follow the protocol, and plugins handle everything else.

---

## How It Works

### Pod Creation Flow

```
API Server
  → Kubelet receives Pod creation request
    → Container Runtime creates a new network namespace
      → Runtime reads config from /etc/cni/net.d/
        → Runtime invokes CNI plugin (ADD) with config via stdin
          → Plugin creates veth pair, attaches to namespace, assigns IP
            → Plugin returns result (interfaces, IPs, routes) via stdout
              → Container starts in the prepared namespace
```

When a Pod is deleted, the runtime calls the plugin with **DEL** — the plugin cleans up interfaces, releases IPs, and removes routes in reverse order.

> Starting from Kubernetes 1.24, the **container runtime** (not kubelet) is responsible for loading and managing CNI plugins.

### CNI Operations

The CNI spec defines 6 operations, passed via the `CNI_COMMAND` environment variable:

| Operation | Purpose |
|-----------|---------|
| `ADD` | Create a network interface for the container — assign IP, set up routes |
| `DEL` | Remove interface, release resources — runs plugins in reverse order |
| `CHECK` | Verify current network state matches the expected configuration |
| `VERSION` | Return the list of spec versions the plugin supports |
| `GC` | Clean up stale resources no longer belonging to any valid attachment |
| `STATUS` | Check whether the plugin is ready to serve ADD requests |

### Environment Variables Passed to Plugins

The runtime passes execution context to plugins via environment variables:

| Variable | Purpose |
|----------|---------|
| `CNI_COMMAND` | Operation to execute (ADD, DEL, CHECK...) |
| `CNI_CONTAINERID` | Unique container identifier |
| `CNI_NETNS` | Path to the network namespace (`/var/run/netns/...`) |
| `CNI_IFNAME` | Interface name to create (typically `eth0`) |
| `CNI_ARGS` | Extra parameters as key=value pairs, separated by `;` |
| `CNI_PATH` | List of directories containing plugin binaries |

### Plugin Chaining

CNI allows chaining multiple plugins in sequence — each plugin receives the previous plugin's result via `prevResult`. This mechanism separates concerns: one plugin creates the interface, another assigns IPs, another limits bandwidth.

```
bridge (creates veth pair)
  → host-local (assigns IP from subnet pool)
    → portmap (maps hostPort to container)
      → bandwidth (limits throughput)
```

For `ADD` and `CHECK`, plugins run in declared order. For `DEL`, plugins run in **reverse order**.

---

## Plugin Types

CNI plugins are organized into 3 categories:

| Type | Purpose | Examples |
|------|---------|----------|
| **Main** | Create the actual network interface | `bridge`, `ptp`, `ipvlan`, `macvlan`, `vlan` |
| **IPAM** | Manage and allocate IP addresses | `host-local`, `dhcp`, `static` |
| **Meta** | Augment other plugins — do not create interfaces | `portmap`, `bandwidth`, `tuning`, `firewall` |

---

## CNI Configuration

### Default Paths

```
/etc/cni/net.d/      # Directory for config files (.conflist, .conf)
/opt/cni/bin/         # Directory for plugin binaries
```

### Config File Structure

```json
{
  "cniVersion": "1.0.0",
  "name": "k8s-pod-network",
  "plugins": [
    {
      "type": "calico",                        // binary name in /opt/cni/bin/
      "datastore_type": "kubernetes",
      "nodename": "worker-01",
      "ipam": {
        "type": "host-local",                  // delegate IPAM to host-local plugin
        "subnet": "usePodCidr"
      },
      "policy": {
        "type": "k8s"                          // use NetworkPolicy from Kubernetes API
      },
      "kubernetes": {
        "kubeconfig": "/etc/cni/net.d/calico-kubeconfig"
      }
    },
    {
      "type": "portmap",                       // meta plugin — maps hostPort
      "capabilities": { "portMappings": true }
    },
    {
      "type": "bandwidth",                     // meta plugin — limits traffic
      "capabilities": { "bandwidth": true }
    }
  ]
}
```

### Traffic Shaping via Annotations

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: limited-pod
  annotations:
    kubernetes.io/ingress-bandwidth: 10M    # limit incoming traffic
    kubernetes.io/egress-bandwidth: 5M      # limit outgoing traffic
spec:
  containers:
  - name: app
    image: nginx:1.27
```

> Requires the `bandwidth` plugin in the chain and its binary present in `/opt/cni/bin/`.

---

## Plugin Result Structure

On a successful `ADD`, the plugin returns JSON via stdout:

```json
{
  "cniVersion": "1.0.0",
  "interfaces": [
    { "name": "eth0", "mac": "0a:58:0a:f4:00:05", "sandbox": "/var/run/netns/cni-xxxxx" }
  ],
  "ips": [
    { "address": "10.244.0.5/24", "gateway": "10.244.0.1", "interface": 0 }
  ],
  "routes": [
    { "dst": "0.0.0.0/0", "gw": "10.244.0.1" }
  ],
  "dns": {
    "nameservers": ["10.96.0.10"]
  }
}
```

On failure, the plugin returns an error code (0–99 reserved, 100+ plugin-defined) along with `msg` and `details`.

---

## Popular CNI Plugins Compared

| | Flannel | Calico | Cilium | Weave Net |
|---|---|---|---|---|
| **Network Architecture** | VXLAN overlay | BGP / L3 (± overlay) | eBPF | Mesh overlay |
| **Network Policy** | Not supported | Full support | Identity-based (L3–L7) | Basic |
| **Traffic Encryption** | None | Optional (WireGuard) | Optional (WireGuard) | Enabled by default |
| **Complexity** | Low | Medium | High | Low–Medium |
| **Performance** | Moderate | High (L3 mode) | Highest | Moderate |
| **Observability** | None | Flow logs | Hubble (deep, real-time) | Basic |
| **Best For** | Dev/test, small clusters | Enterprise, multi-tenant | Security-first, high scale | Multi-cloud, encryption needed |

### Flannel

**Flannel** is the simplest CNI — the `flanneld` daemon on each node manages subnets, and kernel VXLAN handles forwarding. No NetworkPolicy support, no encryption. Suitable for dev/test clusters or environments that don't require network policies.

### Calico

**Calico** uses BGP to route traffic directly at L3 — no overlay needed, significantly reducing overhead. Full NetworkPolicy support, with an optional eBPF dataplane that achieves performance close to Cilium. A popular choice for production enterprise clusters.

### Cilium

**Cilium** leverages eBPF in the Linux kernel for kernel-level packet processing — completely bypassing iptables. Supports identity-based policies (L3–L7), deep packet inspection, and observability via Hubble. Best suited for large clusters requiring high security and performance.

### Weave Net

**Weave Net** creates a mesh overlay between all nodes with encryption enabled by default. Simple setup, suitable for multi-cloud or hybrid deployments that require security but accept the performance trade-off.

---

## Reference Benchmarks (1000 pods)

| CNI | Throughput | Latency |
|-----|-----------|---------|
| Cilium (eBPF) | ~9.2 Gbps | 0.20 ms |
| Calico (BGP) | ~8.5 Gbps | 0.25 ms |
| Flannel (VXLAN) | ~6.5 Gbps | 0.40 ms |
| Weave Net | ~6.0 Gbps | 0.45 ms |

> Data from community benchmarks — actual results depend on hardware, MTU, and workload patterns.

---

## Common Pitfalls

**Changing CNI on a live cluster is extremely risky:** CNI is the foundation of all Pod networking. Swapping CNI on a running cluster requires draining nodes, removing old configs, installing new plugins, and recreating all Pods. A single misstep can cause a cluster-wide network partition. Choose the right CNI during initial cluster design.

**Pods stuck in ContainerCreating — missing CNI binary or config:** If the config file in `/etc/cni/net.d/` references a plugin binary that doesn't exist in `/opt/cni/bin/`, all new Pods will be stuck in `ContainerCreating`. Check with `kubectl describe pod` — the event will show `NetworkPlugin cni failed`.

**IP exhaustion — subnet too small for Pod count:** Each Pod requires its own IP. If the Pod CIDR is too small (e.g., `/26` = 62 IPs/node), the cluster will run out of IPs when scaling up. Calculate Pod CIDR based on `maxPods` per node × expected node count.

**NetworkPolicy silently ignored — using a CNI that doesn't support it:** Flannel does not support NetworkPolicy. Creating a `NetworkPolicy` resource still succeeds (the Kubernetes API accepts it) but no plugin enforces it — traffic flows freely, creating a false sense of security.

**Storage and Pod traffic competing for bandwidth:** Most CNI plugins don't differentiate storage traffic from Pod traffic — both share the same interface. When running heavy database workloads (e.g., PostgreSQL backups), storage I/O can saturate bandwidth and cause application latency. Consider separating interfaces for the storage network on bare-metal.

**Not enabling default-deny policy from the start:** By default, Kubernetes allows all Pods to communicate with each other. Without a default-deny NetworkPolicy from day one, every Pod is reachable — a compromised Pod can perform lateral movement across the entire cluster.

---

## References

- [Network Plugins - Kubernetes Docs](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/)
- [CNI Specification](https://www.cni.dev/docs/spec/)
- [CNI Plugins Repository](https://github.com/containernetworking/plugins)
- [Cluster Networking - Kubernetes Docs](https://kubernetes.io/docs/concepts/cluster-administration/networking/)
- [Calico Documentation](https://docs.tigera.io/calico/latest/about/)
- [Cilium Documentation](https://docs.cilium.io/en/stable/)
- [Flannel Repository](https://github.com/flannel-io/flannel)
- [Gateway API](../networking/gateway-api/README.md)
