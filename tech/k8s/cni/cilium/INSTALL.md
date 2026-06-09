# Cài đặt Cilium CNI thay thế kube-proxy

Tài liệu này hướng dẫn 2 scenario:

1. **Cluster mới** — init cluster không có kube-proxy từ đầu, dùng Cilium thay thế hoàn toàn.
2. **Cluster đang chạy** — migrate từ kube-proxy sang Cilium kube-proxy replacement trên cluster production.

---

## Prerequisites

- Kubernetes >= 1.16 (cần `--skip-phases` support trong kubeadm)
- Linux kernel >= 4.19 (recommend >= 5.10 cho đầy đủ eBPF feature)
- Helm >= 3.x
- Cilium CLI (optional, dùng để verify)
- `kubectl` configured với cluster admin access

Kiểm tra kernel trên tất cả node:

```bash
uname -r
# Expect: 5.10+ cho production
```

---

## Scenario 1 — Cluster mới không có kube-proxy

### Tổng quan

Khi init cluster bằng kubeadm, mặc định kubeadm sẽ deploy kube-proxy DaemonSet. Để dùng Cilium thay thế hoàn toàn, cần skip phase `addon/kube-proxy` lúc init — kube-proxy sẽ không bao giờ được deploy.

### Bước 1 — Init control plane, skip kube-proxy

**Cách 1: Command line flag**

```bash
sudo kubeadm init \
  --skip-phases=addon/kube-proxy \
  --pod-network-cidr=10.244.0.0/16   # CIDR cho Pod network — Cilium IPAM sẽ dùng dải này
```

**Cách 2: kubeadm config file** (recommend cho production — version control được)

```yaml
# kubeadm-config.yaml
apiVersion: kubeadm.k8s.io/v1beta4
kind: InitConfiguration
skipPhases:
  - addon/kube-proxy                  # không deploy kube-proxy
nodeRegistration:
  kubeletExtraArgs:
    - name: node-ip
      value: "192.168.1.10"           # IP chính của node — quan trọng nếu node có nhiều interface
---
apiVersion: kubeadm.k8s.io/v1beta4
kind: ClusterConfiguration
kubernetesVersion: "v1.32.0"
networking:
  podSubnet: "10.244.0.0/16"         # phải khớp với clusterPoolIPv4PodCIDRList bên dưới
  serviceSubnet: "10.96.0.0/12"
controlPlaneEndpoint: "192.168.1.10:6443"
```

```bash
sudo kubeadm init --config kubeadm-config.yaml
```

> **Quan trọng:** Ghi lại `API_SERVER_IP` và `API_SERVER_PORT` từ output của `kubeadm init` — Cilium cần thông tin này vì không có kube-proxy để resolve Service `kubernetes.default`.

### Bước 2 — Cài Cilium với kube-proxy replacement

```bash
# Thêm Helm repo
helm repo add cilium https://helm.cilium.io/
helm repo update

# Cài Cilium — thay thế hoàn toàn kube-proxy
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  --set kubeProxyReplacement=true \
  --set k8sServiceHost="192.168.1.10" \
  --set k8sServicePort=6443 \
  --set ipam.operator.clusterPoolIPv4PodCIDRList="{10.244.0.0/16}" \
  --set ipam.operator.clusterPoolIPv4MaskSize=24
```

| Helm Value | Giải thích |
|------------|-----------|
| `kubeProxyReplacement=true` | Cilium xử lý toàn bộ Service load balancing bằng eBPF — thay thế iptables/IPVS của kube-proxy |
| `k8sServiceHost` | IP của API Server — **bắt buộc** vì không có kube-proxy để route `kubernetes.default` Service |
| `k8sServicePort` | Port của API Server (thường `6443`) |
| `clusterPoolIPv4PodCIDRList` | Pod CIDR — phải khớp với `--pod-network-cidr` lúc init |
| `clusterPoolIPv4MaskSize` | Subnet size per node — `/24` = 254 Pods/node |

### Bước 3 — Join worker nodes

```bash
# Trên mỗi worker node — dùng token từ kubeadm init output
sudo kubeadm join 192.168.1.10:6443 \
  --token <token> \
  --discovery-token-ca-cert-hash sha256:<hash>
```

Worker node join bình thường — không cần config gì thêm. Cilium Agent (DaemonSet) tự động deploy lên worker mới.

### Bước 4 — Verify

```bash
# Cilium agent chạy trên tất cả node
kubectl get pods -n kube-system -l k8s-app=cilium

# Xác nhận kube-proxy replacement active
cilium status | grep KubeProxyReplacement
# Expected output: KubeProxyReplacement:   True

# Xác nhận không có kube-proxy
kubectl get ds -n kube-system kube-proxy 2>&1
# Expected: Error from server (NotFound)

# Test connectivity end-to-end
cilium connectivity test
```

---

## Scenario 2 — Migrate cluster đang chạy kube-proxy

### Tổng quan

Cluster đang dùng kube-proxy (iptables/IPVS mode). Mục tiêu: chuyển sang Cilium xử lý Service load balancing, sau đó gỡ kube-proxy.

> **Cảnh báo:** kube-proxy và Cilium kube-proxy replacement hoạt động **độc lập** — NAT table của hai bên không biết nhau. Khi chuyển đổi, **existing connections có thể bị đứt**. Nên thực hiện trong maintenance window.

Có 2 cách tiếp cận:

| Approach | Risk | Downtime | Phù hợp |
|----------|------|----------|---------|
| **All-at-once** | Cao hơn | Ngắn (rolling restart) | Dev/staging, cluster nhỏ |
| **Gradual (node-by-node)** | Thấp | Gần như zero | Production, cluster lớn |

---

### Approach A — All-at-once Migration

#### Bước 1 — Đảm bảo Cilium đã chạy

Nếu cluster đã dùng Cilium làm CNI (nhưng chưa bật kube-proxy replacement):

```bash
cilium status
# Xác nhận: KubeProxyReplacement: False (hoặc không hiện)
```

Nếu chưa cài Cilium, cài trước với `kubeProxyReplacement=false` (default) và đợi stable:

```bash
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  --set ipam.operator.clusterPoolIPv4PodCIDRList="{10.244.0.0/16}"

# Đợi tất cả Cilium agent Ready
kubectl rollout status ds/cilium -n kube-system
```

#### Bước 2 — Enable kube-proxy replacement

```bash
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --reuse-values \
  --set kubeProxyReplacement=true \
  --set k8sServiceHost="192.168.1.10" \
  --set k8sServicePort=6443
```

Cilium agent trên tất cả node sẽ rolling restart. Sau khi restart, mỗi agent load eBPF program cho Service load balancing.

```bash
# Đợi rollout hoàn tất
kubectl rollout status ds/cilium -n kube-system

# Verify
cilium status | grep KubeProxyReplacement
# Expected: True
```

#### Bước 3 — Xoá kube-proxy

```bash
# Xoá DaemonSet
kubectl -n kube-system delete ds kube-proxy

# Xoá ConfigMap (optional — cleanup)
kubectl -n kube-system delete cm kube-proxy
```

#### Bước 4 — Cleanup iptables rules của kube-proxy

kube-proxy đã xoá nhưng iptables rules vẫn còn trên mỗi node. Cần cleanup để tránh conflict:

```bash
# Chạy trên MỖI node (ssh vào hoặc dùng DaemonSet)
# Xoá tất cả KUBE-* chains
iptables-save | grep -v KUBE | iptables-restore
ip6tables-save | grep -v KUBE | ip6tables-restore
```

**Hoặc dùng DaemonSet để cleanup tự động trên tất cả node:**

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
        - operator: Exists                     # chạy trên tất cả node kể cả control plane
      containers:
        - name: cleanup
          image: alpine:3.20
          securityContext:
            privileged: true                   # cần quyền modify iptables
          command:
            - /bin/sh
            - -c
            - |
              apk add --no-cache iptables ip6tables
              iptables-save | grep -v KUBE | iptables-restore
              ip6tables-save | grep -v KUBE | ip6tables-restore
              echo "Cleanup done on $(hostname)"
              sleep infinity                   # giữ pod alive để check logs
```

```bash
kubectl apply -f kube-proxy-cleanup.yaml

# Kiểm tra logs — đảm bảo cleanup thành công trên mỗi node
kubectl logs -n kube-system -l app=kube-proxy-cleanup --all-containers

# Xoá DaemonSet cleanup sau khi verify
kubectl delete -f kube-proxy-cleanup.yaml
```

#### Bước 5 — Verify toàn bộ

```bash
# Không còn kube-proxy
kubectl get ds -n kube-system kube-proxy 2>&1
# Expected: NotFound

# Cilium đang thay thế kube-proxy
cilium status | grep KubeProxyReplacement
# Expected: True

# Kiểm tra iptables không còn KUBE rules (chạy trên node)
iptables-save | grep -c KUBE
# Expected: 0

# Test Service connectivity
cilium connectivity test
```

---

### Approach B — Gradual Migration (Node-by-Node)

Cách này dùng **`CiliumNodeConfig`** CRD để enable kube-proxy replacement từng node một — giảm blast radius.

#### Bước 1 — Đảm bảo Cilium có `k8sServiceHost`

Nếu chưa set, upgrade trước:

```bash
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --reuse-values \
  --set k8sServiceHost="192.168.1.10" \
  --set k8sServicePort=6443
```

> Cilium Agent cần biết API Server IP trước khi kube-proxy bị gỡ trên node đó. Nếu thiếu, Agent mất kết nối tới API Server ngay khi kube-proxy dừng.

#### Bước 2 — Patch kube-proxy: chỉ chạy trên node chưa migrate

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

Kết quả: khi label node với `io.cilium.migration/kube-proxy-replacement=true`, kube-proxy pod sẽ tự evict khỏi node đó.

#### Bước 3 — Tạo CiliumNodeConfig cho node đã migrate

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
    kube-proxy-replacement: "true"             # enable eBPF kube-proxy trên node có label
```

```bash
kubectl apply -f cilium-node-config-migration.yaml
```

#### Bước 4 — Migrate từng node

Lặp lại cho mỗi node:

```bash
NODE="worker-01"

# 1. Label node — triggers: kube-proxy evict + Cilium agent restart với kube-proxy replacement
kubectl label node $NODE io.cilium.migration/kube-proxy-replacement=true

# 2. Đợi Cilium agent restart trên node
kubectl -n kube-system get pods -l k8s-app=cilium --field-selector spec.nodeName=$NODE -w

# 3. Verify kube-proxy replacement active trên node
kubectl -n kube-system exec $(kubectl -n kube-system get pods -l k8s-app=cilium \
  --field-selector spec.nodeName=$NODE -o name) -- \
  cilium-dbg status | grep KubeProxyReplacement
# Expected: True

# 4. Verify kube-proxy đã evict khỏi node
kubectl -n kube-system get pods -l k8s-app=kube-proxy --field-selector spec.nodeName=$NODE
# Expected: No resources found

# 5. Test Service trên node này
kubectl run test-svc --image=busybox --restart=Never --overrides='{
  "spec": {"nodeName": "'$NODE'"}}' -- \
  wget -qO- --timeout=5 kubernetes.default.svc.cluster.local/healthz
kubectl delete pod test-svc

# OK — chuyển sang node tiếp theo
```

> **Tip:** Migrate control plane node cuối cùng — nếu có vấn đề, control plane vẫn có kube-proxy để maintain API Server accessibility.

#### Bước 5 — Finalize sau khi migrate hết node

```bash
# 1. Set kube-proxy replacement globally (không phụ thuộc CiliumNodeConfig nữa)
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --reuse-values \
  --set kubeProxyReplacement=true

# 2. Xoá kube-proxy DaemonSet (không còn pod nào chạy, nhưng DaemonSet vẫn tồn tại)
kubectl -n kube-system delete ds kube-proxy
kubectl -n kube-system delete cm kube-proxy

# 3. Cleanup CiliumNodeConfig (không cần nữa vì đã set global)
kubectl delete -f cilium-node-config-migration.yaml

# 4. Cleanup labels
kubectl label nodes --all io.cilium.migration/kube-proxy-replacement-

# 5. Cleanup iptables trên tất cả node (xem DaemonSet cleanup ở Approach A — Bước 4)
```

---

## Rollback về kube-proxy

Nếu cần rollback (connectivity issues, eBPF bugs trên kernel cụ thể):

```bash
# 1. Re-deploy kube-proxy
kubectl -n kube-system apply -f /etc/kubernetes/manifests/kube-proxy.yaml
# Hoặc nếu không còn manifest:
kubeadm init phase addon kube-proxy --kubeconfig /etc/kubernetes/admin.conf

# 2. Đợi kube-proxy Ready trên tất cả node
kubectl rollout status ds/kube-proxy -n kube-system

# 3. Disable kube-proxy replacement trong Cilium
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --reuse-values \
  --set kubeProxyReplacement=false

# 4. Restart Cilium agents
kubectl -n kube-system rollout restart ds/cilium
kubectl rollout status ds/cilium -n kube-system

# 5. Verify kube-proxy đang xử lý Services
iptables-save | grep -c KUBE
# Expected: > 0 (kube-proxy rules được tạo lại)
```

---

## Helm Values Reference

Tổng hợp các Helm values quan trọng cho kube-proxy replacement:

```yaml
# values-kpr.yaml — production values cho kube-proxy replacement
kubeProxyReplacement: true

k8sServiceHost: "192.168.1.10"        # API Server IP — bắt buộc
k8sServicePort: 6443                   # API Server port

ipam:
  operator:
    clusterPoolIPv4PodCIDRList:
      - "10.244.0.0/16"               # Pod CIDR — không trùng node network
    clusterPoolIPv4MaskSize: 24

# Tuning cho kube-proxy replacement
bpf:
  masquerade: true                     # eBPF masquerading thay iptables SNAT
  tproxy: true                         # transparent proxy support

# Socket-based load balancing — connect() level thay vì packet level
socketLB:
  enabled: true                        # Pod connect trực tiếp tới backend, skip Service NAT
  hostNamespaceOnly: false             # apply cho cả Pod namespace

# NodePort
nodePort:
  enabled: true                        # cho phép truy cập Service qua NodePort
  range: "30000-32767"                 # range mặc định, khớp với kube-apiserver --service-node-port-range

# ExternalIPs
externalIPs:
  enabled: true

# HostPort
hostPort:
  enabled: true

# Session affinity
sessionAffinity: true                  # hỗ trợ Service sessionAffinity: ClientIP
```

```bash
# Apply từ file
helm install cilium cilium/cilium \
  --version 1.19.4 \
  --namespace kube-system \
  -f values-kpr.yaml
```

---

## Lưu ý thực tế

**Maintenance window cho migration:** Khi switch từ kube-proxy sang Cilium, existing TCP connections qua Service IP **sẽ bị reset** — NAT table của hai bên độc lập. Schedule migration trong maintenance window, đặc biệt cho stateful workload (database connections, WebSocket).

**Multi-interface nodes:** Nếu node có nhiều network interface, đảm bảo kubelet `--node-ip` trỏ đúng interface chính. Cilium kube-proxy replacement dựa vào `node-ip` để xác định endpoint address — sai interface = Service traffic route sai.

**Cleanup iptables không được bỏ qua:** Sau khi xoá kube-proxy, iptables KUBE-\* rules vẫn active cho đến khi bị xoá thủ công hoặc node reboot. Rules cũ có thể conflict với eBPF datapath — gây duplicate DNAT hoặc packet loop.

**Không mix kube-proxy mode giữa các node quá lâu:** Trong gradual migration, một số node chạy kube-proxy, một số chạy Cilium eBPF. Hai cơ chế hoạt động tương thích trong thời gian chuyển tiếp, nhưng không nên duy trì trạng thái hybrid lâu — khó debug khi có vấn đề.

---

## Tham khảo

- [Kubernetes Without kube-proxy — Cilium Docs](https://docs.cilium.io/en/stable/network/kubernetes/kubeproxy-free/)
- [Installation using kubeadm — Cilium Docs](https://docs.cilium.io/en/latest/installation/k8s-install-kubeadm/)
- [Per-node Configuration (Gradual Migration) — Cilium Docs](https://docs.cilium.io/en/stable/configuration/per-node-config/)
- [Installation using Helm — Cilium Docs](https://docs.cilium.io/en/stable/installation/k8s-install-helm/)
- [Cilium CNI](./README.md)
- [CNI trong Kubernetes](../README.md)
