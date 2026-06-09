# Cilium CNI in Kubernetes

## What is Cilium CNI?

**Cilium** is a CNI plugin that uses **eBPF** (extended Berkeley Packet Filter) to process packets directly in the Linux kernel — instead of relying on iptables or IPVS like traditional CNIs. As a CNI, Cilium handles creating network namespaces, assigning IPs, setting up routing, and connecting Pod-to-Pod across the entire cluster. All major cloud providers (GKE, EKS, AKS) have adopted Cilium as their default or integrated CNI.

> Cilium also provides many other features (L3–L7 security policy, observability via Hubble, service mesh). This document focuses solely on Cilium's **CNI — networking** functionality.

---

## Problems Cilium CNI Solves

- **iptables doesn't scale:** Traditional CNIs (Calico iptables mode, Flannel) translate each Service/NetworkPolicy into iptables rules. In large clusters (thousands of Services), long chains cause linearly increasing latency and slow rule updates.
- **Modern kernels underutilized:** iptables/IPVS are legacy mechanisms that cannot leverage the programmable datapath capabilities of kernel >= 4.19 — wasting resources on unnecessary context switching and stack traversal.
- **kube-proxy is a separate bottleneck:** kube-proxy runs independently, maintaining a parallel rule set alongside the CNI — adding complexity, making debugging harder, and synchronizing slowly when Services change.
- **Overlay networking overhead:** VXLAN/Geneve encapsulation adds headers, reducing effective MTU and increasing CPU usage — especially impactful for high-throughput workloads.

Cilium solves these by moving the entire datapath (routing, load balancing, policy enforcement) into **eBPF programs running directly in the kernel** — constant-time lookup instead of linear chain traversal.

---

## Architecture

### Core Components

| Component | Role |
|-----------|------|
| **Cilium Agent** | DaemonSet running on every node — watches Kubernetes API, compiles eBPF programs, manages endpoints |
| **Cilium Operator** | Centralized Deployment — manages IPAM, garbage collects stale identities, handles cluster-wide tasks |
| **Cilium CNI Plugin** | Binary at `/opt/cni/bin/cilium-cni` — invoked by kubelet when a Pod is scheduled, communicates with the Agent via Unix socket |
| **eBPF Datapath** | Programs attached to TC (Traffic Control) hooks on network interfaces — processes packets at kernel level |

### Pod Creation Flow

```
Kubelet receives Pod assignment
  → Container Runtime creates network namespace
    → Runtime invokes cilium-cni (ADD)
      → cilium-cni sends request to Cilium Agent via Unix socket
        → Agent creates veth pair (pod eth0 ↔ host lxc*)
          → Agent assigns IP from IPAM pool
            → Agent loads eBPF program into TC hook of veth
              → Pod is ready to communicate
```

When a Pod is deleted, the runtime calls `cilium-cni` with `DEL` — the Agent unloads the eBPF program, releases the IP, and removes the veth pair.

---

## How It Works — eBPF Datapath

### Packet Flow (Pod-to-Pod on the same node)

```
Pod A (eth0) → veth peer (lxc*) on host
  → eBPF program on TC ingress hook receives packet
    → Lookup destination in eBPF map
      → Forward directly to Pod B's veth peer
        → Pod B (eth0) receives packet
```

The core difference: **eBPF host-routing** completely bypasses iptables and the upper network stack in the host namespace. Packets are picked up directly from the network device and delivered straight into the Pod namespace — significantly reducing context switching.

### Packet Flow (Pod-to-Pod on different nodes)

Depends on the networking mode:

- **Tunnel mode (VXLAN/Geneve):** Packets are encapsulated at the source node, sent through the tunnel, and decapsulated at the destination node. Works on any network topology but has encapsulation overhead.
- **Native routing:** Packets are routed directly through the network infrastructure (BGP, cloud VPC routing). No encapsulation overhead but requires the underlying network to know how to route the Pod CIDR.

---

## Networking Modes

| Mode | Encapsulation | Network Requirements | Performance | Use Case |
|------|---------------|---------------------|-------------|----------|
| **VXLAN** (default) | UDP port 8472 | Only IP connectivity between nodes | Good — ~50 bytes/packet overhead | Any environment, especially when you don't control the underlying network |
| **Geneve** | UDP port 6081 | Only IP connectivity between nodes | Similar to VXLAN — supports extensible metadata | When passing metadata between tunnel endpoints |
| **Native Routing** | None | Network must be able to route Pod CIDR | Highest — no overhead | Cloud VPC, bare-metal with BGP, nodes on the same L2 segment |

### When to use which mode?

- **VXLAN/Geneve:** When you don't control the network infrastructure (shared datacenter, multi-cloud) or need quick setup without depending on router/switch configuration.
- **Native Routing:** When running on cloud VPCs (AWS, GCP, Azure all route Pod CIDRs natively) or bare-metal with BGP peering. Optimal performance due to no encapsulation.

---

## IPAM (IP Address Management)

Cilium supports multiple IPAM modes. **You cannot change IPAM mode on a running cluster** — choose correctly from the start.

| Mode | How It Works | When to Use |
|------|-------------|-------------|
| **Cluster Pool** (default) | Cilium Operator divides a large CIDR into per-node pools via `CiliumNode` CRD | Self-managed clusters, no dependency on Kubernetes CIDR allocation |
| **Kubernetes** | Delegates to the Kubernetes node CIDR allocator (`--allocate-node-cidrs`) | Clusters where `controller-manager` is already configured to assign PodCIDRs |
| **AWS ENI** | Each Pod gets a real VPC ENI IP — fully routable | EKS or self-managed on AWS where Pod IPs must be routable within the VPC |
| **Azure IPAM** | Similar to ENI, uses Azure network interfaces | AKS or self-managed on Azure |

### Cluster Pool — Configuration

```yaml
# Helm values for Cluster Pool IPAM
ipam:
  mode: cluster-pool
  operator:
    clusterPoolIPv4PodCIDRList:
      - "10.244.0.0/16"        # total CIDR — must not overlap with node network
    clusterPoolIPv4MaskSize: 24 # each node gets a /24 = 254 Pod IPs
```

> `clusterPoolIPv4PodCIDRList` defaults to `10.0.0.0/8`. If your node network also uses the `10.x` range — **you will lose connectivity between nodes**. Always set an explicit CIDR.

---

## Installation

### Helm (recommended)

```bash
# Add Helm repo
helm repo add cilium https://helm.cilium.io/
helm repo update

# Basic installation — tunnel mode, cluster-pool IPAM
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  --set ipam.operator.clusterPoolIPv4PodCIDRList="{10.244.0.0/16}" \
  --set ipam.operator.clusterPoolIPv4MaskSize=24
```

### Replacing kube-proxy

```bash
# Install Cilium with kube-proxy replacement
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  --set kubeProxyReplacement=true \
  --set k8sServiceHost=<API_SERVER_IP> \      # required when kube-proxy is absent
  --set k8sServicePort=6443 \
  --set ipam.operator.clusterPoolIPv4PodCIDRList="{10.244.0.0/16}"
```

> When enabling `kubeProxyReplacement`, you must set `k8sServiceHost` and `k8sServicePort` because there is no kube-proxy to resolve the Kubernetes Service IP.

### Native Routing on cloud

```bash
# AWS VPC — native routing + ENI IPAM
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  --set eni.enabled=true \
  --set ipam.mode=eni \
  --set routingMode=native \
  --set kubeProxyReplacement=true \
  --set k8sServiceHost=<API_SERVER_IP> \
  --set k8sServicePort=443
```

### Verify

```bash
# Check Cilium agent is running on all nodes
kubectl get pods -n kube-system -l k8s-app=cilium

# Overall health check
cilium status

# End-to-end connectivity test (creates test pods, sends traffic)
cilium connectivity test
```

---

## CNI Configuration File

The Cilium Agent automatically writes and maintains the config file at `/etc/cni/net.d/05-cilium.conflist`:

```json
{
  "cniVersion": "0.3.1",
  "name": "cilium",
  "plugins": [
    {
      "type": "cilium-cni",
      "enable-debug": false,
      "log-file": "/var/run/cilium/cilium-cni.log"
    }
  ]
}
```

Related Helm values:

| Helm Value | Default | Description |
|------------|---------|-------------|
| `cni.install` | `true` | Automatically install CNI binary and config |
| `cni.exclusive` | `true` | Remove other CNI plugin configs — ensures Cilium is the only CNI |
| `cni.customConf` | `false` | `true` = don't overwrite config — use when managing config manually |

---

## Networking Mode Comparison

| | VXLAN (Cilium) | Native Routing (Cilium) | Calico BGP | Flannel VXLAN |
|---|---|---|---|---|
| **Encapsulation** | VXLAN | None | None | VXLAN |
| **Datapath** | eBPF | eBPF | iptables/eBPF | iptables |
| **Throughput** (1000 pods) | ~8.5 Gbps | ~9.2 Gbps | ~8.5 Gbps (BGP) | ~6.5 Gbps |
| **Latency** | ~0.22 ms | ~0.20 ms | ~0.25 ms | ~0.40 ms |
| **Network Requirements** | IP connectivity | Route Pod CIDR | BGP peering | IP connectivity |
| **Replaces kube-proxy** | Yes | Yes | Yes (eBPF mode) | No |
| **Setup Complexity** | Low | Medium | Medium | Low |

> Benchmarks from community sources — actual results depend on hardware, MTU, and workload patterns.

---

## Common Pitfalls

**Pitfall 1 — Kernel version too old:** Cilium requires kernel >= 4.19 (recommend >= 5.10 for full feature support). On older kernels, eBPF programs fail to load — Cilium Agent enters a crash loop, and all new Pods get stuck in `ContainerCreating`. Check `uname -r` on all nodes before installing.

**Pitfall 2 — Pod CIDR overlaps with node network:** The default IPAM CIDR is `10.0.0.0/8`. If the node network also uses the `10.x` range, traffic between nodes gets misrouted into the Pod network — **cluster connectivity is lost**. Always set `clusterPoolIPv4PodCIDRList` explicitly and ensure no overlap.

**Pitfall 3 — Forgetting to set `k8sServiceHost` when enabling kube-proxy replacement:** When `kubeProxyReplacement=true` without setting `k8sServiceHost`/`k8sServicePort`, the Cilium Agent cannot connect to the API Server — the DaemonSet fails, and all networking goes down. This is the most common mistake when migrating from kube-proxy to Cilium.

**Pitfall 4 — MTU mismatch in tunnel mode:** VXLAN adds 50 bytes of overhead, Geneve adds ~58 bytes. If MTU is not reduced accordingly (e.g., node MTU 1500 → Pod MTU 1450), packets get fragmented or dropped — causing intermittent timeouts that are hard to debug. Cilium auto-detects MTU, but verify with `cilium status | grep MTU`.

**Pitfall 5 — Changing IPAM mode on a running cluster:** Cilium does not support changing the IPAM mode or `clusterPoolIPv4MaskSize` after deployment. Changes cause IP allocation failures, and Pods cannot obtain new IPs. If a change is needed, you must drain nodes and reinstall Cilium.

**Pitfall 6 — Conflicts with old CNI:** If Cilium is installed on a cluster that already has another CNI (Flannel, Calico) without cleaning up old configs in `/etc/cni/net.d/`, kubelet may load the old config instead of Cilium's (due to lexicographic ordering). The Helm value `cni.exclusive=true` (default) removes other configs, but verify with `ls /etc/cni/net.d/` after installation.

**Pitfall 7 — Agent down = no networking for new Pods on that node:** If the Cilium Agent on a node crashes or gets evicted (OOM), all new Pods scheduled on that node will be stuck in `ContainerCreating`. Existing Pods continue to work (eBPF programs are already loaded in the kernel). Monitoring Agent health is critical — set appropriate resource requests/limits.

---

## References

- [Cilium Documentation](https://docs.cilium.io/en/stable/)
- [Installation using Helm — Cilium Docs](https://docs.cilium.io/en/stable/installation/k8s-install-helm/)
- [IPAM — Cilium Docs](https://docs.cilium.io/en/stable/network/concepts/ipam/)
- [Kubernetes Configuration — Cilium Docs](https://docs.cilium.io/en/stable/network/kubernetes/configuration/)
- [High Performance Cloud Native Networking (CNI) — Cilium](https://cilium.io/use-cases/cni/)
- [CNI Benchmark — Cilium Blog](https://cilium.io/blog/2021/05/11/cni-benchmark/)
- [CNI in Kubernetes](../README.md)
- [Calico CNI](../calico/README.md)
