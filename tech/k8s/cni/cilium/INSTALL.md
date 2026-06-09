# Installing Cilium CNI as kube-proxy Replacement

This document covers 2 scenarios:

1. **New cluster** — init the cluster without kube-proxy from the start, using Cilium as a complete replacement.
2. **Existing cluster** — migrate from kube-proxy to Cilium kube-proxy replacement on a production cluster.

---

## Prerequisites

- Kubernetes >= 1.16 (`--skip-phases` support in kubeadm)
- Linux kernel >= 4.19 (recommend >= 5.10 for full eBPF feature support)
- Helm >= 3.x
- Cilium CLI (optional, used for verification)
- `kubectl` configured with cluster admin access

Check kernel version on all nodes:

```bash
uname -r
# Expect: 5.10+ for production
```

---

## Scenario 1 — New Cluster Without kube-proxy

### Overview

When initializing a cluster with kubeadm, kube-proxy DaemonSet is deployed by default. To use Cilium as a complete replacement, skip the `addon/kube-proxy` phase during init — kube-proxy will never be deployed.

### Step 1 — Init control plane, skip kube-proxy

**Option 1: Command line flag**

```bash
sudo kubeadm init \
  --skip-phases=addon/kube-proxy \
  --pod-network-cidr=10.244.0.0/16   # Pod network CIDR — Cilium IPAM will use this range
```

**Option 2: kubeadm config file** (recommended for production — can be version controlled)

```yaml
# kubeadm-config.yaml
apiVersion: kubeadm.k8s.io/v1beta4
kind: InitConfiguration
skipPhases:
  - addon/kube-proxy                  # do not deploy kube-proxy
nodeRegistration:
  kubeletExtraArgs:
    - name: node-ip
      value: "192.168.1.10"           # primary node IP — important if node has multiple interfaces
---
apiVersion: kubeadm.k8s.io/v1beta4
kind: ClusterConfiguration
kubernetesVersion: "v1.32.0"
networking:
  podSubnet: "10.244.0.0/16"         # must match clusterPoolIPv4PodCIDRList below
  serviceSubnet: "10.96.0.0/12"
controlPlaneEndpoint: "192.168.1.10:6443"
```

```bash
sudo kubeadm init --config kubeadm-config.yaml
```

> **Important:** Note the `API_SERVER_IP` and `API_SERVER_PORT` from the `kubeadm init` output — Cilium needs this information since there is no kube-proxy to resolve the `kubernetes.default` Service.

### Step 2 — Install Cilium with kube-proxy replacement

```bash
# Add Helm repo
helm repo add cilium https://helm.cilium.io/
helm repo update

# Install Cilium — fully replacing kube-proxy
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  --set kubeProxyReplacement=true \
  --set k8sServiceHost="192.168.1.10" \
  --set k8sServicePort=6443 \
  --set ipam.operator.clusterPoolIPv4PodCIDRList="{10.244.0.0/16}" \
  --set ipam.operator.clusterPoolIPv4MaskSize=24
```

| Helm Value | Description |
|------------|------------|
| `kubeProxyReplacement=true` | Cilium handles all Service load balancing via eBPF — replaces kube-proxy's iptables/IPVS |
| `k8sServiceHost` | API Server IP — **required** since there is no kube-proxy to route the `kubernetes.default` Service |
| `k8sServicePort` | API Server port (usually `6443`) |
| `clusterPoolIPv4PodCIDRList` | Pod CIDR — must match `--pod-network-cidr` from init |
| `clusterPoolIPv4MaskSize` | Subnet size per node — `/24` = 254 Pods/node |

### Step 3 — Join worker nodes

```bash
# On each worker node — use token from kubeadm init output
sudo kubeadm join 192.168.1.10:6443 \
  --token <token> \
  --discovery-token-ca-cert-hash sha256:<hash>
```

Worker nodes join normally — no additional configuration needed. Cilium Agent (DaemonSet) automatically deploys to new workers.

### Step 4 — Verify

```bash
# Cilium agent running on all nodes
kubectl get pods -n kube-system -l k8s-app=cilium

# Confirm kube-proxy replacement is active
cilium status | grep KubeProxyReplacement
# Expected output: KubeProxyReplacement:   True

# Confirm kube-proxy does not exist
kubectl get ds -n kube-system kube-proxy 2>&1
# Expected: Error from server (NotFound)

# End-to-end connectivity test
cilium connectivity test
```

---

## Scenario 2 — Migrating an Existing Cluster from kube-proxy

### Overview

The cluster is currently using kube-proxy (iptables/IPVS mode). Goal: switch to Cilium for Service load balancing, then remove kube-proxy.

> **Warning:** kube-proxy and Cilium kube-proxy replacement operate **independently** — their NAT tables are unaware of each other. During the switch, **existing connections may be disrupted**. Perform this during a maintenance window.

Two approaches are available:

| Approach | Risk | Downtime | Best For |
|----------|------|----------|----------|
| **All-at-once** | Higher | Short (rolling restart) | Dev/staging, small clusters |
| **Gradual (node-by-node)** | Lower | Near-zero | Production, large clusters |

---

### Approach A — All-at-once Migration

#### Step 1 — Ensure Cilium is running

If the cluster already uses Cilium as CNI (but kube-proxy replacement is not yet enabled):

```bash
cilium status
# Confirm: KubeProxyReplacement: False (or not shown)
```

If Cilium is not yet installed, install it first with `kubeProxyReplacement=false` (default) and wait for it to stabilize:

```bash
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  --set ipam.operator.clusterPoolIPv4PodCIDRList="{10.244.0.0/16}"

# Wait for all Cilium agents to be Ready
kubectl rollout status ds/cilium -n kube-system
```

#### Step 2 — Enable kube-proxy replacement

```bash
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --reuse-values \
  --set kubeProxyReplacement=true \
  --set k8sServiceHost="192.168.1.10" \
  --set k8sServicePort=6443
```

Cilium agents on all nodes will perform a rolling restart. After restart, each agent loads eBPF programs for Service load balancing.

```bash
# Wait for rollout to complete
kubectl rollout status ds/cilium -n kube-system

# Verify
cilium status | grep KubeProxyReplacement
# Expected: True
```

#### Step 3 — Delete kube-proxy

```bash
# Delete the DaemonSet
kubectl -n kube-system delete ds kube-proxy

# Delete the ConfigMap (optional — cleanup)
kubectl -n kube-system delete cm kube-proxy
```

#### Step 4 — Clean up kube-proxy iptables rules

kube-proxy is deleted but its iptables rules remain on each node. Clean up to avoid conflicts:

```bash
# Run on EVERY node (SSH in or use a DaemonSet)
# Remove all KUBE-* chains
iptables-save | grep -v KUBE | iptables-restore
ip6tables-save | grep -v KUBE | ip6tables-restore
```

**Or use a DaemonSet to clean up automatically across all nodes:**

```yaml
# kube-proxy-cleanup.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kube-proxy-cleanup
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: kube-proxy-cleanup
  template:
    metadata:
      labels:
        app: kube-proxy-cleanup
    spec:
      hostNetwork: true
      tolerations:
        - operator: Exists                     # run on all nodes including control plane
      containers:
        - name: cleanup
          image: alpine:3.20
          securityContext:
            privileged: true                   # requires privileges to modify iptables
          command:
            - /bin/sh
            - -c
            - |
              apk add --no-cache iptables ip6tables
              iptables-save | grep -v KUBE | iptables-restore
              ip6tables-save | grep -v KUBE | ip6tables-restore
              echo "Cleanup done on $(hostname)"
              sleep infinity                   # keep pod alive to check logs
```

```bash
kubectl apply -f kube-proxy-cleanup.yaml

# Check logs — ensure cleanup succeeded on each node
kubectl logs -n kube-system -l app=kube-proxy-cleanup --all-containers

# Delete the cleanup DaemonSet after verification
kubectl delete -f kube-proxy-cleanup.yaml
```

#### Step 5 — Full verification

```bash
# kube-proxy no longer exists
kubectl get ds -n kube-system kube-proxy 2>&1
# Expected: NotFound

# Cilium is replacing kube-proxy
cilium status | grep KubeProxyReplacement
# Expected: True

# Verify no KUBE iptables rules remain (run on node)
iptables-save | grep -c KUBE
# Expected: 0

# Test Service connectivity
cilium connectivity test
```

---

### Approach B — Gradual Migration (Node-by-Node)

This approach uses the **`CiliumNodeConfig`** CRD to enable kube-proxy replacement one node at a time — reducing blast radius.

#### Step 1 — Ensure Cilium has `k8sServiceHost` configured

If not yet set, upgrade first:

```bash
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --reuse-values \
  --set k8sServiceHost="192.168.1.10" \
  --set k8sServicePort=6443
```

> The Cilium Agent needs to know the API Server IP before kube-proxy is removed from that node. Without it, the Agent loses connectivity to the API Server as soon as kube-proxy stops.

#### Step 2 — Patch kube-proxy to only run on unmigrated nodes

```bash
kubectl -n kube-system patch ds kube-proxy -p '{
  "spec": {
    "template": {
      "spec": {
        "affinity": {
          "nodeAffinity": {
            "requiredDuringSchedulingIgnoredDuringExecution": {
              "nodeSelectorTerms": [{
                "matchExpressions": [{
                  "key": "io.cilium.migration/kube-proxy-replacement",
                  "operator": "NotIn",
                  "values": ["true"]
                }]
              }]
            }
          }
        }
      }
    }
  }
}'
```

Result: when a node is labeled with `io.cilium.migration/kube-proxy-replacement=true`, the kube-proxy pod will be automatically evicted from that node.

#### Step 3 — Create CiliumNodeConfig for migrated nodes

```yaml
# cilium-node-config-migration.yaml
apiVersion: cilium.io/v2
kind: CiliumNodeConfig
metadata:
  name: kube-proxy-replacement-migrated
  namespace: kube-system
spec:
  nodeSelector:
    matchLabels:
      io.cilium.migration/kube-proxy-replacement: "true"
  defaults:
    kube-proxy-replacement: "true"             # enable eBPF kube-proxy on labeled nodes
```

```bash
kubectl apply -f cilium-node-config-migration.yaml
```

#### Step 4 — Migrate nodes one by one

Repeat for each node:

```bash
NODE="worker-01"

# 1. Label node — triggers: kube-proxy eviction + Cilium agent restart with kube-proxy replacement
kubectl label node $NODE io.cilium.migration/kube-proxy-replacement=true

# 2. Wait for Cilium agent to restart on the node
kubectl -n kube-system get pods -l k8s-app=cilium --field-selector spec.nodeName=$NODE -w

# 3. Verify kube-proxy replacement is active on the node
kubectl -n kube-system exec $(kubectl -n kube-system get pods -l k8s-app=cilium \
  --field-selector spec.nodeName=$NODE -o name) -- \
  cilium-dbg status | grep KubeProxyReplacement
# Expected: True

# 4. Verify kube-proxy has been evicted from the node
kubectl -n kube-system get pods -l k8s-app=kube-proxy --field-selector spec.nodeName=$NODE
# Expected: No resources found

# 5. Test Service connectivity on this node
kubectl run test-svc --image=busybox --restart=Never --overrides='{
  "spec": {"nodeName": "'$NODE'"}}' -- \
  wget -qO- --timeout=5 kubernetes.default.svc.cluster.local/healthz
kubectl delete pod test-svc

# OK — proceed to the next node
```

> **Tip:** Migrate control plane nodes last — if issues arise, the control plane still has kube-proxy to maintain API Server accessibility.

#### Step 5 — Finalize after all nodes are migrated

```bash
# 1. Set kube-proxy replacement globally (no longer dependent on CiliumNodeConfig)
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --reuse-values \
  --set kubeProxyReplacement=true

# 2. Delete kube-proxy DaemonSet (no pods running, but the DaemonSet still exists)
kubectl -n kube-system delete ds kube-proxy
kubectl -n kube-system delete cm kube-proxy

# 3. Clean up CiliumNodeConfig (no longer needed since it's set globally)
kubectl delete -f cilium-node-config-migration.yaml

# 4. Clean up labels
kubectl label nodes --all io.cilium.migration/kube-proxy-replacement-

# 5. Clean up iptables on all nodes (see DaemonSet cleanup in Approach A — Step 4)
```

---

## Rolling Back to kube-proxy

If rollback is needed (connectivity issues, eBPF bugs on a specific kernel):

```bash
# 1. Re-deploy kube-proxy
kubectl -n kube-system apply -f /etc/kubernetes/manifests/kube-proxy.yaml
# Or if the manifest is no longer available:
kubeadm init phase addon kube-proxy --kubeconfig /etc/kubernetes/admin.conf

# 2. Wait for kube-proxy to be Ready on all nodes
kubectl rollout status ds/kube-proxy -n kube-system

# 3. Disable kube-proxy replacement in Cilium
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --reuse-values \
  --set kubeProxyReplacement=false

# 4. Restart Cilium agents
kubectl -n kube-system rollout restart ds/cilium
kubectl rollout status ds/cilium -n kube-system

# 5. Verify kube-proxy is handling Services
iptables-save | grep -c KUBE
# Expected: > 0 (kube-proxy rules are recreated)
```

---

## Helm Values Reference

Summary of important Helm values for kube-proxy replacement:

```yaml
# values-kpr.yaml — production values for kube-proxy replacement
kubeProxyReplacement: true

k8sServiceHost: "192.168.1.10"        # API Server IP — required
k8sServicePort: 6443                   # API Server port

ipam:
  operator:
    clusterPoolIPv4PodCIDRList:
      - "10.244.0.0/16"               # Pod CIDR — must not overlap with node network
    clusterPoolIPv4MaskSize: 24

# Tuning for kube-proxy replacement
bpf:
  masquerade: true                     # eBPF masquerading instead of iptables SNAT
  tproxy: true                         # transparent proxy support

# Socket-based load balancing — operates at connect() level instead of packet level
socketLB:
  enabled: true                        # Pods connect directly to backends, skipping Service NAT
  hostNamespaceOnly: false             # apply to Pod namespaces as well

# NodePort
nodePort:
  enabled: true                        # allow accessing Services via NodePort
  range: "30000-32767"                 # default range, matches kube-apiserver --service-node-port-range

# ExternalIPs
externalIPs:
  enabled: true

# HostPort
hostPort:
  enabled: true

# Session affinity
sessionAffinity: true                  # support Service sessionAffinity: ClientIP
```

```bash
# Apply from file
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  -f values-kpr.yaml
```

---

## Practical Notes

**Maintenance window for migration:** When switching from kube-proxy to Cilium, existing TCP connections through Service IPs **will be reset** — the NAT tables of both systems are independent. Schedule migration during a maintenance window, especially for stateful workloads (database connections, WebSockets).

**Multi-interface nodes:** If a node has multiple network interfaces, ensure kubelet's `--node-ip` points to the correct primary interface. Cilium kube-proxy replacement relies on `node-ip` to determine endpoint addresses — wrong interface = Service traffic routed incorrectly.

**Do not skip iptables cleanup:** After removing kube-proxy, iptables KUBE-\* rules remain active until manually removed or the node is rebooted. Leftover rules can conflict with the eBPF datapath — causing duplicate DNAT or packet loops.

**Do not maintain hybrid mode for too long:** During gradual migration, some nodes run kube-proxy while others run Cilium eBPF. The two mechanisms are compatible during the transition period, but maintaining the hybrid state long-term is not recommended — debugging issues becomes significantly harder.

---

## References

- [Kubernetes Without kube-proxy — Cilium Docs](https://docs.cilium.io/en/stable/network/kubernetes/kubeproxy-free/)
- [Installation using kubeadm — Cilium Docs](https://docs.cilium.io/en/latest/installation/k8s-install-kubeadm/)
- [Per-node Configuration (Gradual Migration) — Cilium Docs](https://docs.cilium.io/en/stable/configuration/per-node-config/)
- [Installation using Helm — Cilium Docs](https://docs.cilium.io/en/stable/installation/k8s-install-helm/)
- [Cilium CNI](./README.md)
- [CNI in Kubernetes](../README.md)
