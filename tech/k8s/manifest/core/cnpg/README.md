# CloudNativePG - Cluster Manifests (bare-metal on-prem)

Bộ manifest cài đặt một PostgreSQL cluster (1 primary + 2 replica) bằng
CloudNativePG trên cluster bare-metal, dùng Longhorn làm storage local.

Chi tiết khái niệm (kiến trúc, failover, fencing, vì sao Longhorn strict-local)
xem [tech/k8s/cnpg/README.md](../../../cnpg/README.md). File này chỉ là runbook cài đặt.

## Yêu cầu trước

- Longhorn đã cài (provisioner `driver.longhorn.io`) - xem [core/storage/longhorn.sh](../storage/longhorn.sh).
- `helm` và `kubectl` trỏ đúng cluster.
- Cluster >= 3 node để rải 3 instance mỗi node một cái.

## Thứ tự cài đặt

Dependency chặt nên file được đánh số, apply đúng thứ tự:

```bash
# 1. Cài operator (chạy 1 chỗ ở ns cnpg-system, watch Cluster CR mọi namespace)
./cnpg.sh

# 2. Namespace chứa cluster
kubectl apply -f 01-namespace.yaml

# 3. StorageClass (cluster-scoped) + Secret credentials
kubectl apply -f 02-storage.yaml
#    Tạo Secret bằng lệnh thay vì commit mật khẩu (khuyến nghị):
kubectl -n cnpg-cluster create secret generic cnpg-cluster-credentials \
  --type=kubernetes.io/basic-auth \
  --from-literal=username=appuser \
  --from-literal=password="$(openssl rand -base64 24)"
#    Hoặc sửa placeholder rồi: kubectl apply -f 03-secret.yaml

# 4. Cluster - operator bắt đầu bootstrap sau bước này
kubectl apply -f 04-cluster.yaml
```

Secret (bước 3) PHẢI có trước Cluster (bước 4); thiếu thì bootstrap lỗi.

## Kiểm tra

```bash
# operator sẵn sàng
kubectl -n cnpg-system get deploy -l app.kubernetes.io/name=cloudnative-pg

# cluster + instance (chờ tới khi Cluster in healthy state)
kubectl -n cnpg-cluster get cluster,pods
kubectl cnpg status cnpg-cluster-postgresql -n cnpg-cluster   # cần plugin kubectl-cnpg
```

## Kết nối

Operator tự sinh 3 Service trong ns `cnpg-cluster`:

| Service | Trỏ tới | Dùng cho |
|---------|---------|----------|
| `cnpg-cluster-postgresql-rw` | primary | ghi (read-write) |
| `cnpg-cluster-postgresql-ro` | các replica (cân tải) | đọc (read-only) |
| `cnpg-cluster-postgresql-r`  | mọi instance | đọc bất kỳ |

App ghi qua `cnpg-cluster-postgresql-rw.cnpg-cluster.svc:5432`.

## Ghi chú

- **Backup**: replication chỉ chống mất node, KHÔNG chống xóa nhầm/corruption.
  Production cần backup riêng qua Barman Cloud Plugin (in-tree `barmanObjectStore`
  deprecated từ 1.26) - hướng dẫn ở cuối [04-cluster.yaml](04-cluster.yaml).
- **Monitoring**: `enablePodMonitor: false`; chỉ bật khi đã cài Prometheus Operator.
- **Gỡ**: xóa Cluster trước, PVC giữ lại do `reclaimPolicy: Retain` (dọn thủ công).
