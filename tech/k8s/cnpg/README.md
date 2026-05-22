# CloudNativePG trên Kubernetes

## CloudNativePG là gì?

**CloudNativePG (CNPG)** là một Kubernetes Operator quản lý toàn bộ lifecycle của PostgreSQL cluster — từ khởi tạo, replication, failover đến backup — thông qua một CRD duy nhất: `Cluster`. Thay vì tự viết StatefulSet + init container + sidecar script để vận hành PostgreSQL, bạn khai báo một resource `Cluster` và operator xử lý toàn bộ logic vận hành.

Bên dưới, CNPG vẫn tạo Pod, PVC, Service. Điểm khác biệt so với [StatefulSet thuần](../workload/statefulset/README.md#bare-statefulset-vs-operator): CNPG **hiểu PostgreSQL** — nó biết cách cấu hình streaming replication, phát hiện primary chết, promote replica, và fence node lỗi. StatefulSet chỉ biết tạo Pod theo thứ tự.

Một điểm quan trọng: CNPG **không dùng StatefulSet**. Nó quản lý Pod trực tiếp — điều này cho phép operator toàn quyền quyết định tạo/xóa/promote Pod nào mà không bị ràng buộc bởi thứ tự tuần tự của StatefulSet.

---

## Bài toán giải quyết

Chạy PostgreSQL trên Kubernetes bằng StatefulSet thuần yêu cầu tự giải quyết mọi vấn đề vận hành:

- **Replication** — phải tự viết init container chạy `pg_basebackup`, cấu hình `primary_conninfo`, quản lý replication slot. CNPG tự động làm khi `instances > 1`
- **Failover** — primary chết thì không có gì xảy ra. Phải tự phát hiện lỗi, chọn replica tốt nhất (ít replication lag nhất), promote nó, reconfigure các replica còn lại trỏ về primary mới. CNPG làm trong vài giây
- **Fencing (chống split-brain)** — nếu node mất mạng nhưng Pod vẫn chạy, có thể xuất hiện hai primary ghi vào hai disk khác nhau. CNPG fence primary cũ bằng cách sửa `pg_hba.conf` từ chối mọi connection trước khi promote replica
- **Backup** — phải tự setup CronJob cho `pg_basebackup` hoặc `pg_dump`, quản lý retention, verify restore. CNPG khai báo backup schedule và destination ngay trong spec của `Cluster`

---

## Kiến trúc

```
                    ┌─────────────────────────┐
                    │   CNPG Operator (Pod)    │
                    │   watches Cluster CRD    │
                    └────────────┬────────────┘
                                 │ creates & manages
              ┌──────────────────┼──────────────────┐
              ▼                  ▼                   ▼
        ┌──────────┐      ┌──────────┐        ┌──────────┐
        │ Pod-0    │      │ Pod-1    │        │ Pod-2    │
        │ PRIMARY  │─WAL─▶│ REPLICA  │   WAL─▶│ REPLICA  │
        └────┬─────┘      └────┬─────┘        └────┬─────┘
             │                 │                    │
        ┌────┴─────┐     ┌────┴─────┐        ┌────┴─────┐
        │ PVC data │     │ PVC data │        │ PVC data │
        │ PVC wal  │     │ PVC wal  │        │ PVC wal  │
        └──────────┘     └──────────┘        └──────────┘
```

Mỗi Pod có **hai PVC**: một cho data (`PGDATA`), một cho WAL. Tách WAL ra volume riêng giúp tránh WAL bloat làm đầy data disk và giảm contention I/O (WAL ghi tuần tự, data I/O ngẫu nhiên).

Operator tự động tạo 3 Service:

| Service | Trỏ đến |
|---------|---------|
| `{cluster}-rw` | Luôn trỏ đến primary — dùng cho write |
| `{cluster}-ro` | Load-balance giữa các replica — dùng cho read |
| `{cluster}-r` | Tất cả instance (primary + replica) |

---

## Tại sao dùng Longhorn cho CNPG?

PostgreSQL đã tự replicate data ở tầng application qua streaming replication (`instances: 3`). Nếu Longhorn cũng replicate ở tầng storage thì là **double replication** — tốn disk và I/O mà không có lợi ích thêm.

Giải pháp: dùng Longhorn với `dataLocality: strict-local` và `numberOfReplicas: "1"` — data nằm trên cùng node với Pod, không replicate qua node khác. Kết quả:

- **Không có network I/O** khi đọc/ghi database — critical cho PostgreSQL performance
- **Dynamic provisioning** — CNPG khai báo `storageClass`, Longhorn tự tạo PV
- **Volume expansion** — mở rộng PVC không cần downtime

Trade-off: nếu node chết, Longhorn volume trên node đó mất. Nhưng PostgreSQL đã có 2 replica khác trên 2 node khác — operator tự promote replica và tạo lại instance mới.

---

## Cấu trúc Manifest

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: cnpg-cluster
---
apiVersion: v1
kind: Secret
metadata:
  name: cnpg-cluster-credentials
  namespace: cnpg-cluster
type: kubernetes.io/basic-auth          # yêu cầu đúng 2 field: username + password
stringData:
  username: appuser
  password: "<your-password>"
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: longhorn-cnpg
provisioner: driver.longhorn.io
allowVolumeExpansion: true
parameters:
  numberOfReplicas: "1"                 # không replicate — PostgreSQL đã tự replicate
  dataLocality: "strict-local"          # data nằm trên cùng node với Pod
  staleReplicaTimeout: "2880"
  fsType: "ext4"
reclaimPolicy: Retain                   # xóa PVC không xóa PV — bảo vệ data
volumeBindingMode: WaitForFirstConsumer # tạo PV sau khi Pod được schedule — bắt buộc với strict-local
---
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: cnpg-cluster-postgresql
  namespace: cnpg-cluster
spec:
  description: "CloudNativePG Production PostgreSQL Cluster"
  instances: 3                                    # 1 primary + 2 replica
  imageName: ghcr.io/cloudnative-pg/postgresql:18.4

  storage:
    storageClass: longhorn-cnpg
    size: 25Gi
  walStorage:                                     # tách WAL ra PVC riêng
    storageClass: longhorn-cnpg
    size: 5Gi                                     # thường 10-20% data volume

  bootstrap:
    initdb:                                       # chỉ chạy lần đầu khởi tạo cluster
      database: bootstrap-cluster-db
      owner: appuser
      secret:
        name: cnpg-cluster-credentials            # Secret phải tồn tại trước khi apply Cluster

  primaryUpdateStrategy: unsupervised             # operator tự switchover khi update

  affinity:
    enablePodAntiAffinity: true
    topologyKey: kubernetes.io/hostname
    podAntiAffinityType: preferred                # spread Pod ra nhiều node, không block scheduling
  topologySpreadConstraints:
    - maxSkew: 1
      topologyKey: kubernetes.io/hostname
      whenUnsatisfiable: ScheduleAnyway
      labelSelector:
        matchLabels:
          cnpg.io/cluster: cnpg-cluster-postgresql

  monitoring:
    enablePodMonitor: false                       # bật nếu đã cài Prometheus Operator
```

---

## Các trường quan trọng

### `bootstrap`

Định nghĩa cách cluster được khởi tạo **lần đầu tiên**. Sau khi bootstrap xong, primary chạy và các replica tự clone từ primary qua streaming replication.

| Method | Khi nào dùng |
|--------|-------------|
| `initdb` | Cluster mới — chạy `initdb` tạo PostgreSQL instance từ đầu |
| `recovery` | Restore từ backup (object store hoặc Volume Snapshot) |
| `pg_basebackup` | Clone từ một PostgreSQL server có sẵn (kể cả không phải CNPG) |

---

### `walStorage`

WAL (Write-Ahead Log) là cơ chế durability của PostgreSQL — mọi thay đổi được ghi vào WAL trước khi apply vào data file. Tách WAL ra PVC riêng:

- **Ngăn data disk full** — WAL bloat (từ long-running transaction hoặc replication lag) không ảnh hưởng data volume
- **Giảm contention I/O** — WAL ghi tuần tự, data I/O ngẫu nhiên, tách disk tránh tranh chấp

---

### `primaryUpdateStrategy`

| Strategy | Hành vi |
|----------|---------|
| `unsupervised` | Operator xử lý toàn bộ: update replica trước, rồi switchover (replica lên primary, primary cũ restart với image mới). Downtime ngắn trong lúc switchover |
| `supervised` | Operator chỉ update replica. Phải tự trigger switchover bằng `kubectl cnpg promote` |

---

### `affinity` và `topologySpreadConstraints`

Kết hợp `podAntiAffinity` với `topologySpreadConstraints` để phân bổ Pod ra nhiều node. Nếu một node chết, chỉ mất một instance — hai instance còn lại tiếp tục phục vụ.

`podAntiAffinityType: preferred` (không phải `required`): "spread nếu có thể, nhưng nếu chỉ có 2 node mà 3 instance thì vẫn schedule instance thứ 3 lên node đã có." Với `required`, Pod thứ 3 sẽ Pending mãi mãi.

---

### `volumeBindingMode: WaitForFirstConsumer`

Bắt buộc khi dùng `strict-local`. Nếu không có, Longhorn có thể tạo volume trên node-1 trong khi scheduler đặt Pod lên node-2 — Pod sẽ Pending vì không access được volume.

---

## Common Pitfalls

**`strict-local` mà không có `WaitForFirstConsumer`:** Longhorn tạo volume trên node ngẫu nhiên, Pod được schedule lên node khác → Pending mãi. Luôn dùng cặp `strict-local` + `WaitForFirstConsumer`.

**Không tách WAL storage:** WAL và data chung PVC. Long-running transaction hoặc replication slot bloat WAL, đầy disk, PostgreSQL crash. Production luôn dùng `walStorage`.

**Secret chưa tồn tại khi apply Cluster:** CNPG bootstrap thất bại, cluster vào trạng thái lỗi. Apply Secret trước Cluster.

**`podAntiAffinityType: required` trên cluster nhỏ:** 3 instance, 2 node → Pod thứ 3 Pending mãi. Dùng `preferred` trừ khi chắc chắn có đủ node.

**`enablePodMonitor: true` mà chưa cài Prometheus Operator:** PodMonitor CRD không tồn tại, CNPG log error. Chỉ bật khi đã cài [kube-prometheus-stack](../monitor/README.md).

**Không cấu hình backup:** Replication bảo vệ khỏi node failure. Backup bảo vệ khỏi xóa nhầm, corruption, ransomware. Hai thứ khác nhau — cấu hình `backup` trong Cluster spec hoặc dùng giải pháp backup ngoài.

---

## References

- [CloudNativePG Documentation](https://cloudnative-pg.io/documentation/current/)
- [CloudNativePG Helm Chart](https://github.com/cloudnative-pg/charts)
- [Longhorn Best Practices](https://longhorn.io/docs/latest/best-practices/)
- [StatefulSet — Bare vs Operator](../workload/statefulset/README.md#bare-statefulset-vs-operator)
- [StorageClass](../storage/storage_classes/README.md)
- [PersistentVolume](../storage/persistent_volume/README.md)
