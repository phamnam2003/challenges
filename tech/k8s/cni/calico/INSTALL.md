# Installing Calico CNI on Kubernetes

## Prerequisites

- Kubernetes cluster initialized (kubeadm, kubespray, or equivalent)
- `kubectl` configured to point to the cluster
- `kubeadm init` ran with `--pod-network-cidr` — this CIDR must match the Calico IP pool
- No other CNI plugin installed on nodes (`/etc/cni/net.d/` is empty)

---

## Method 1 — Operator (recommended for production)

The Tigera Operator manages Calico's full lifecycle — installation, upgrades, scaling, config changes. This is the official installation method.

### Step 1 — Initialize cluster with kubeadm

```bash
sudo kubeadm init \
  --pod-network-cidr=10.244.0.0/16 \       # Pod CIDR — must match ipPools below
  --control-plane-endpoint=<IP_CONTROL_PLANE>
```

> If using Calico's default CIDR `192.168.0.0/16`, no changes are needed in step 3.

### Step 2 — Install Tigera Operator and CRDs

```bash
# CRDs — define Calico custom resources
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/v1_crd_projectcalico_org.yaml

# Operator — controller that manages Calico components
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/tigera-operator.yaml
```

### Step 3 — Download and customize Installation CR

```bash
# Download default manifest
curl -O https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/custom-resources.yaml
```

Edit `custom-resources.yaml` to match your environment:

```yaml
apiVersion: operator.tigera.io/v1
kind: Installation
metadata:
  name: default
spec:
  calicoNetwork:
    # Dataplane option — defaults to Iptables
    # linuxDataplane: BPF                   # uncomment to use eBPF dataplane

    # MTU — adjust based on encapsulation mode
    mtu: 1450                               # 1500 for BGP, 1480 for IPIP, 1450 for VXLAN

    ipPools:
    - name: default-ipv4-ippool
      cidr: 10.244.0.0/16                  # MUST match --pod-network-cidr from kubeadm
      blockSize: 26                         # /26 = 64 IPs per block — default, rarely needs changing
      natOutgoing: Enabled                  # SNAT for traffic leaving the cluster
      nodeSelector: all()

      # Choose encapsulation mode matching your infrastructure
      encapsulation: VXLANCrossSubnet       # see table below

    # bgp: Disabled                         # uncomment for VXLAN-only (disables BIRD)
---
apiVersion: operator.tigera.io/v1
kind: APIServer
metadata:
  name: default
spec: {}
```

#### Encapsulation selection table

| Value | Behavior | When to Use |
|-------|----------|-------------|
| `None` | No encapsulation — BGP pure L3 | On-prem with BGP-capable ToR routers |
| `IPIP` | IP-in-IP for all traffic | On-prem without BGP, not on Azure |
| `IPIPCrossSubnet` | IP-in-IP only across subnets | On-prem multi-subnet |
| `VXLAN` | VXLAN for all traffic | Cloud, Azure (blocks IP-in-IP) |
| `VXLANCrossSubnet` | VXLAN only across subnets **(default)** | Multi-subnet — best of both worlds |

If your CIDR differs from the default `192.168.0.0/16`, quick fix with sed:

```bash
sed -i 's+192.168.0.0/16+10.244.0.0/16+' custom-resources.yaml
```

### Step 4 — Apply Installation CR

```bash
kubectl create -f custom-resources.yaml
```

### Step 5 — Wait and verify

```bash
# Monitor deployment status (3–5 minutes)
watch kubectl get tigerastatus
```

Expected output — all components show `AVAILABLE: True`:

```
NAME                  AVAILABLE   PROGRESSING   DEGRADED   SINCE
apiserver             True        False         False      2m
calico                True        False         False      30s
```

Additional checks:

```bash
kubectl get pods -n calico-system              # all pods Running
kubectl get nodes                               # all nodes Ready
kubectl get ippools -o wide                     # confirm CIDR and encapsulation
```

---

## Method 2 — Manifest (for special cases)

> Use only when you need deep customization of Kubernetes resources that the operator doesn't support. The operator won't manage the lifecycle — you handle upgrades yourself.

### Small clusters (≤ 50 nodes)

```bash
curl https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/calico.yaml -O

# If CIDR differs from 192.168.0.0/16, update CALICO_IPV4POOL_CIDR
sed -i 's+192.168.0.0/16+10.244.0.0/16+' calico.yaml

kubectl apply -f calico.yaml
```

### Large clusters (> 50 nodes, Typha required)

```bash
curl https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/calico-typha.yaml -o calico.yaml

# Edit Typha replicas: 1 instance / 200 nodes, minimum 3 for production
# Find the calico-typha Deployment and update replicas

kubectl apply -f calico.yaml
```

### Using etcd datastore (instead of Kubernetes API)

```bash
curl https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/calico-etcd.yaml -o calico.yaml

# Edit etcd_endpoints in ConfigMap calico-config
# e.g.: "https://10.0.0.1:2379,https://10.0.0.2:2379"

kubectl apply -f calico.yaml
```

---

## Method 3 — eBPF Dataplane

To use eBPF instead of iptables, download the dedicated manifest:

```bash
# Download eBPF custom resources
curl -O https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/custom-resources-bpf.yaml

# Customize CIDR if needed
sed -i 's+192.168.0.0/16+10.244.0.0/16+' custom-resources-bpf.yaml

# Apply (after installing the operator in step 2)
kubectl create -f custom-resources-bpf.yaml
```

> The eBPF dataplane replaces kube-proxy entirely. Once Calico eBPF is stable, you can remove the kube-proxy DaemonSet.

---

## Install calicoctl (management CLI)

```bash
# Download binary
curl -L https://github.com/projectcalico/calico/releases/download/v3.32.0/calicoctl-linux-amd64 -o /usr/local/bin/calicoctl
chmod +x /usr/local/bin/calicoctl

# Verify
calicoctl version
calicoctl get nodes
calicoctl get ippools -o wide
```

---

## Post-Installation Connectivity Test

Create 2 pods on different nodes to test cross-node networking:

```bash
# Create test pods
kubectl run test-a --image=busybox --restart=Never -- sleep 3600
kubectl run test-b --image=busybox --restart=Never -- sleep 3600

# Wait for pods to be Running
kubectl get pods -o wide

# Ping from test-a to test-b's IP
kubectl exec test-a -- ping -c 3 <IP_OF_TEST_B>

# Cleanup
kubectl delete pod test-a test-b
```

If cross-node ping succeeds → Calico CNI is working correctly.

---

## Installation Notes

**CIDR must match:** `--pod-network-cidr` (kubeadm) and `ipPools[].cidr` (Calico) must have the same value. A mismatch means pods get IPs but cross-node routing breaks silently.

**Clean up old CNI before installing:** If nodes previously had Flannel/Weave, delete files in `/etc/cni/net.d/` and binaries in `/opt/cni/bin/` before installing Calico. The container runtime may load the wrong plugin.

**Do not change CIDR after installation:** Calico IPAM does not reassign IPs to running pods. Changing the IP pool CIDR after installation only affects new pods. Plan your CIDR correctly from the start.

**Azure blocks IP-in-IP:** On Azure, use `encapsulation: VXLAN` or `VXLANCrossSubnet`. Do not use IPIP — traffic will be silently dropped.

**MTU must match encapsulation:** Incorrect MTU causes packet fragmentation → severe throughput degradation. See the [MTU table in README](README.md#mtu-considerations).

---

## References

- [Calico On-Premises Installation - Calico Docs](https://docs.tigera.io/calico/latest/getting-started/kubernetes/self-managed-onprem/onpremises)
- [Calico Quickstart - Calico Docs](https://docs.tigera.io/calico/latest/getting-started/kubernetes/quickstart)
- [Determine Best Networking - Calico Docs](https://docs.tigera.io/calico/latest/networking/determine-best-networking)
- [Calico System Requirements](https://docs.tigera.io/calico/latest/getting-started/kubernetes/requirements)
- [Calico CNI Overview](README.md)
- [CNI Overview](../README.md)
