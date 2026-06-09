# Calico CNI in Kubernetes

## What is Calico CNI?

**Calico** is a CNI plugin that provides high-performance, scalable pod networking using a pure Layer 3 approach. Unlike overlay-only CNI plugins, Calico can route pod traffic directly via **BGP** without encapsulation — making pod IPs natively routable across the physical network. It also supports **VXLAN** and **IP-in-IP** overlay modes for environments where BGP peering is not available. This document focuses exclusively on Calico's CNI networking functionality — IPAM, routing, encapsulation modes, and dataplane options.

---

## Problems It Solves

- **Overlay overhead at scale:** CNI plugins that rely solely on VXLAN/overlay add encapsulation cost to every packet — header overhead, reduced MTU, extra CPU for encap/decap. Calico's BGP mode routes traffic at L3 with zero encapsulation.
- **Pod IPs not routable outside the cluster:** With overlay-only CNIs, pod IPs are invisible to the physical network — debugging, monitoring, and firewall rules must work through node IPs. Calico BGP mode advertises pod subnets to the network fabric, making pod IPs directly reachable from outside the cluster.
- **Single dataplane, no flexibility:** Most CNIs lock you into one dataplane implementation. Calico offers a pluggable dataplane — iptables (stable default), eBPF (high performance, kube-proxy replacement), or nftables — so you can choose based on your scale and performance requirements.
- **IPAM rigidity:** Basic CNIs assign a fixed subnet per node with no flexibility. Calico IPAM allocates IP blocks dynamically from IP pools, supports per-namespace/per-pod pool selection via annotations, and can reclaim unused blocks.

---

## Architecture

### Core Components

```
┌─────────────────────────────────────────────────────┐
│  Kubernetes API Server / etcd                       │
│      ↓ watches                                      │
│  Typha (optional, recommended for 200+ nodes)       │
│      ↓ fans out updates                             │
│  ┌──────────── Per Node ──────────────────────┐     │
│  │  Felix          — programs routes + ACLs   │     │
│  │  BIRD           — BGP speaker (if BGP mode)│     │
│  │  calico (CNI)   — creates veth, calls IPAM │     │
│  │  calico-ipam    — assigns IP from pool     │     │
│  └────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────┘
```

| Component | Role |
|-----------|------|
| **calico** (CNI binary) | Invoked by the container runtime on pod create/delete — creates veth pair, calls IPAM, configures routing |
| **calico-ipam** | Allocates IPs from Calico IP pools, assigns /26 blocks to nodes, reclaims unused blocks |
| **Felix** | Daemon on each node — programs routes into the kernel routing table, writes iptables/eBPF rules for forwarding |
| **BIRD** | BGP daemon — advertises pod subnets to peers (other nodes, ToR routers). Only active in BGP mode |
| **Typha** | Caching proxy between API server and Felix — prevents N×watch-event storms on large clusters |
| **confd** | Watches datastore for BGP config changes (AS number, peerings, IP pools) and regenerates BIRD config |

### Pod Creation Flow

```
kubelet → container runtime → calico CNI binary
  1. calico reads /etc/cni/net.d/10-calico.conflist
  2. calico-ipam assigns IP from an IP pool block
  3. calico creates veth pair (caliXXXX ↔ eth0 in pod)
  4. calico programs host-side route: pod-IP → caliXXXX
  5. Felix picks up the new endpoint, programs ACLs
  6. BIRD (BGP mode) advertises the pod subnet to peers
```

---

## Networking Modes

Calico supports 4 networking modes. The choice depends on your network infrastructure and performance requirements.

| Mode | Encapsulation | Pod IPs Routable Externally | Performance | When to Use |
|------|---------------|----------------------------|-------------|-------------|
| **BGP (unencapsulated)** | None | Yes | Best — no overhead | On-prem with BGP-capable ToR routers |
| **IP-in-IP** | IP-in-IP header | No | Slight CPU + MTU reduction | On-prem without BGP, not on Azure |
| **VXLAN** | UDP/VXLAN header | No | Slight CPU + MTU reduction | Cloud, multi-cloud, Azure (blocks IP-in-IP) |
| **CrossSubnet** | Only cross-subnet | Within subnet: yes | Best within subnet, overlay only across | Multi-subnet on-prem — best of both worlds |

### BGP (Unencapsulated)

The default and most performant mode. Each node runs a **BIRD** BGP speaker that advertises its pod CIDR block to peers. Traffic between pods on different nodes uses standard IP routing — no tunnel, no encapsulation.

```
Pod A (10.244.0.5) on Node 1
  → kernel route: 10.244.1.0/24 via Node 2
    → physical network routes packet to Node 2
      → Node 2 kernel route: 10.244.1.8 → caliXXXX
        → Pod B (10.244.1.8)
```

**BGP topology options:**

| Topology | Description | Scale |
|----------|-------------|-------|
| **Full mesh** (default) | Every node peers with every other node | < 100 nodes |
| **Route reflectors** | Dedicated nodes aggregate and redistribute routes | 100–1000+ nodes |
| **External peering** | Nodes peer with ToR/spine routers | Integrates with physical network fabric |

### VXLAN

Encapsulates L2 frames in UDP packets. Does **not** use BIRD/BGP — Calico programs routes directly using Felix.

```yaml
# Installation CRD — VXLAN mode
apiVersion: operator.tigera.io/v1
kind: Installation
metadata:
  name: default
spec:
  calicoNetwork:
    bgp: Disabled                      # no BIRD needed in VXLAN-only mode
    ipPools:
    - cidr: 10.244.0.0/16
      encapsulation: VXLAN             # or VXLANCrossSubnet
      natOutgoing: Enabled
```

> Use VXLAN when your network blocks IP-in-IP (Azure) or when you don't have BGP-capable infrastructure.

### IP-in-IP

Wraps the original IP packet inside another IP header. Uses BIRD for route distribution, unlike VXLAN.

```yaml
ipPools:
- cidr: 10.244.0.0/16
  encapsulation: IPIP                  # or IPIPCrossSubnet
  natOutgoing: Enabled
```

### CrossSubnet

Available for both VXLAN and IP-in-IP. Packets within the same L2 subnet are sent **unencapsulated** (direct routing), and only packets crossing subnet boundaries are encapsulated. This gives you near-BGP performance for intra-subnet traffic while still working across subnets.

```yaml
encapsulation: VXLANCrossSubnet        # or IPIPCrossSubnet
```

---

## Dataplane Options

| Dataplane | Mechanism | kube-proxy Replacement | When to Use |
|-----------|-----------|------------------------|-------------|
| **iptables** (default) | iptables rules per pod | No | Stable, well-understood, < 500 services |
| **eBPF** | eBPF programs attached to kernel hooks | Yes | High scale, high throughput, 500+ services |
| **nftables** | nftables rules (modern iptables successor) | No | Modern Linux kernels, transitioning from iptables |

### eBPF Dataplane

Calico's eBPF dataplane bypasses iptables and kube-proxy entirely — handling service load-balancing, policy enforcement, and forwarding in eBPF programs attached directly to network interfaces.

Benefits over iptables:
- **O(1) service lookup** instead of O(n) iptables chain traversal
- **Source IP preservation** without `externalTrafficPolicy: Local`
- **Lower latency** at high service/pod counts
- Replaces kube-proxy — one less component to manage

---

## Installation

### Operator-based (recommended)

```bash
# 1. Install CRDs
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/v1_crd_projectcalico_org.yaml

# 2. Install Tigera operator
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/tigera-operator.yaml

# 3. Create Installation CR (customize encapsulation, CIDR, MTU here)
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/custom-resources.yaml

# 4. Wait for all components
watch kubectl get tigerastatus
```

> The operator manages Calico's lifecycle — upgrades, scaling, config changes. Prefer operator over raw manifests for production.

### Verify Installation

```bash
kubectl get pods -n calico-system          # all pods Running
kubectl get nodes                           # all nodes Ready
kubectl get ippools                         # verify IP pool CIDR and encapsulation
```

For detailed installation steps including kubeadm setup, manifest customization, and eBPF dataplane — see [INSTALL.md](INSTALL.md).

---

## CNI Configuration

### Config File Location

```
/etc/cni/net.d/10-calico.conflist         # CNI config (managed by operator)
/opt/cni/bin/calico                        # CNI binary
/opt/cni/bin/calico-ipam                   # IPAM binary
```

### Config File Structure

```json
{
  "name": "k8s-pod-network",
  "cniVersion": "0.3.1",
  "plugins": [
    {
      "type": "calico",
      "datastore_type": "kubernetes",          // kubernetes or etcdv3
      "log_level": "INFO",                     // ERROR, WARNING, INFO, DEBUG
      "log_file_path": "/var/log/calico/cni/cni.log",
      "ipam": {
        "type": "calico-ipam",                 // Calico's own IPAM — dynamic block allocation
        "assign_ipv4": "true",
        "assign_ipv6": "false",
        "ipv4_pools": ["10.244.0.0/16"]        // restrict to specific IP pool
      },
      "container_settings": {
        "allow_ip_forwarding": false           // true only if pod needs to act as router
      },
      "policy": {
        "type": "k8s"                          // use Kubernetes NetworkPolicy API
      },
      "kubernetes": {
        "kubeconfig": "/etc/cni/net.d/calico-kubeconfig"
      }
    },
    {
      "type": "portmap",                       // enables hostPort support
      "snat": true,
      "capabilities": { "portMappings": true }
    },
    {
      "type": "bandwidth",                     // optional — traffic shaping
      "capabilities": { "bandwidth": true }
    }
  ]
}
```

---

## IPAM

### How Calico IPAM Works

Calico divides the Pod CIDR into **/26 blocks** (64 IPs each) and assigns blocks to nodes on demand. Unlike host-local IPAM which pre-assigns a fixed /24 per node, Calico IPAM is dynamic — a node that runs more pods gets more blocks, and empty blocks are reclaimed.

```
Pod CIDR: 10.244.0.0/16 (65,536 IPs)
  → Node 1 gets block 10.244.0.0/26   (64 IPs)
  → Node 1 needs more → gets 10.244.0.64/26
  → Node 2 gets block 10.244.0.128/26
  → Node 1 scales down → 10.244.0.64/26 reclaimed
```

### Per-Pod IP Pool Selection

Use annotations to select which IP pool a pod draws from:

```yaml
metadata:
  annotations:
    cni.projectcalico.org/ipv4pools: '["production-pool"]'
```

### Per-Pod Specific IP

```yaml
metadata:
  annotations:
    cni.projectcalico.org/ipAddrs: '["10.244.5.100"]'    # request exact IP from IPAM
```

### Calico IPAM vs host-local

| | Calico IPAM | host-local |
|---|---|---|
| Block allocation | Dynamic /26 blocks, on demand | Fixed /24 per node |
| IP reclamation | Unused blocks returned to pool | No reclamation |
| Per-pod pool selection | Yes (annotation) | No |
| Specific IP request | Yes (annotation) | No |
| Best for | Production, variable workloads | Simple setups, Flannel compatibility |

---

## MTU Considerations

Encapsulation reduces the effective MTU. Configure MTU to avoid fragmentation:

| Mode | Overhead | Recommended MTU (1500 physical) |
|------|----------|--------------------------------|
| BGP (none) | 0 bytes | 1500 |
| IP-in-IP | 20 bytes | 1480 |
| VXLAN | 50 bytes | 1450 |
| WireGuard | 60 bytes | 1440 |

Set MTU in the Installation CR:

```yaml
spec:
  calicoNetwork:
    mtu: 1450                              # must match your encapsulation mode
```

---

## Common Pitfalls

**IP auto-detection picks the wrong interface:** On multi-homed nodes (multiple NICs), Calico may auto-detect the wrong IP for BGP peering or VXLAN endpoints. Nodes become `NotReady` or pods get cross-node connectivity failures. Explicitly set `IP_AUTODETECTION_METHOD` to `interface=eth0` or `can-reach=10.0.0.1` in the Installation CR.

**CIDR mismatch between kubeadm and Calico:** The Calico IP pool CIDR must fall within `--pod-network-cidr` passed to `kubeadm init`. A mismatch means Calico assigns IPs that kube-controller-manager doesn't recognize — pods get IPs but cross-node routing breaks silently.

**Changing IP pool after installation has no effect on existing pods:** Calico IPAM does not reassign IPs to running pods. If you change the pool CIDR or encapsulation mode, existing pods keep their old IPs until recreated. Plan IP pool changes before going to production or during a maintenance window.

**Using IP-in-IP on Azure:** Azure blocks IP protocol 4 (IP-in-IP). Pods on different nodes cannot communicate — traffic silently drops. Use VXLAN (`encapsulation: VXLAN`) on Azure.

**Full-mesh BGP on large clusters (100+ nodes):** The default full-mesh topology creates N×(N-1)/2 BGP sessions. At 200 nodes that's ~20,000 sessions — BIRD consumes excessive CPU and memory. Switch to route reflectors for clusters above ~100 nodes.

**Conflicting CNI plugins left on the node:** If a previous CNI (Flannel, Weave) left config files in `/etc/cni/net.d/`, the container runtime may load the wrong plugin. Clean `/etc/cni/net.d/` and `/opt/cni/bin/` of old CNI files before installing Calico.

**No Typha on large clusters:** Without Typha, every Felix agent opens its own watch to the API server. At 200+ nodes, this creates significant API server load. Enable Typha — one Typha instance can serve ~200 Felix agents.

---

## References

- [Calico Documentation](https://docs.tigera.io/calico/latest/about/)
- [Determine Best Networking - Calico Docs](https://docs.tigera.io/calico/latest/networking/determine-best-networking)
- [Configure CNI Plugins - Calico Docs](https://docs.tigera.io/calico/latest/reference/configure-cni-plugins)
- [Calico Quickstart - Calico Docs](https://docs.tigera.io/calico/latest/getting-started/kubernetes/quickstart)
- [Installation Guide](INSTALL.md)
- [CNI Overview](../README.md)
- [Gateway API](../../networking/gateway-api/README.md)
