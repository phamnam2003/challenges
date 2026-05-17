# PersistentVolumeClaim trong Kubernetes

## PersistentVolumeClaim là gì?

**PersistentVolumeClaim (PVC)** là yêu cầu storage từ phía workload — tương tự như Pod yêu cầu CPU/RAM nhưng dành cho persistent storage. PVC mô tả *cần bao nhiêu dung lượng* và *cần access theo kiểu nào*, còn Kubernetes lo việc tìm (hoặc tạo) PV phù hợp và bind lại.

PVC là namespace-scoped: chỉ Pod trong cùng namespace mới dùng được PVC đó.

---

## Bài toán giải quyết

Không có PVC, Pod phải tự khai báo chi tiết storage backend trong spec của mình (NFS server address, local path...). Điều này buộc dev phải biết rõ hạ tầng storage, và mỗi khi storage thay đổi phải sửa lại toàn bộ workload.

PVC tạo ra một lớp tách biệt:
- **Dev** chỉ cần khai báo nhu cầu: "tôi cần 10Gi, RWO, dùng StorageClass longhorn"
- **Admin** lo phần cung cấp storage phía dưới (qua PV hoặc StorageClass)

---

## Vòng đời của PVC

```
PVC tạo → K8s tìm PV phù hợp → Bind → Pod dùng → PVC bị xóa → PV reclaim
```

### Trạng thái PVC

| Phase | Ý nghĩa |
|-------|---------|
| `Pending` | Chưa tìm được PV phù hợp, đang chờ |
| `Bound` | Đã bind thành công với một PV |
| `Terminating` | Đang bị xóa nhưng vẫn còn Pod đang dùng |

### Cơ chế bảo vệ

Kubernetes không cho phép xóa PVC khi vẫn còn Pod đang mount nó. PVC sẽ ở trạng thái `Terminating` với finalizer `kubernetes.io/pvc-protection` cho đến khi Pod release — tránh mất dữ liệu đột ngột.

---

## Cấu trúc manifest

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  storageClassName: longhorn
  resources:
    requests:
      storage: 10Gi
```

---

## Các trường quan trọng

### `accessModes`

Phải khớp với `accessModes` của PV được bind:

| Mode | Viết tắt | Ý nghĩa |
|------|----------|---------|
| `ReadWriteOnce` | RWO | Một node mount đọc/ghi |
| `ReadOnlyMany` | ROX | Nhiều node mount chỉ đọc |
| `ReadWriteMany` | RWX | Nhiều node mount đọc/ghi |
| `ReadWriteOncePod` | RWOP | Chỉ một Pod duy nhất mount đọc/ghi |

---

### `storageClassName`

Xác định StorageClass nào xử lý PVC này:

| Giá trị | Hành vi |
|---------|---------|
| Tên class (vd `longhorn`) | Dynamic provisioning — provisioner tự tạo PV |
| Tên class không có provisioner (vd `local-storage`) | K8s tìm PV tĩnh có cùng `storageClassName` |
| `""` (chuỗi rỗng) | Tắt dynamic provisioning, chỉ bind với PV không có storageClass |
| Không khai báo | Dùng default StorageClass của cluster |

---

### `resources.requests.storage`

Dung lượng tối thiểu cần thiết. PVC chỉ bind với PV có dung lượng **bằng hoặc lớn hơn** giá trị này.

---

### `volumeMode`

| Giá trị | Hành vi |
|---------|---------|
| `Filesystem` (mặc định) | Mount như directory |
| `Block` | Expose như raw block device |

---

### `volumeName`

Bind trực tiếp với một PV cụ thể theo tên, bỏ qua quá trình matching tự động:

```yaml
spec:
  volumeName: local-pv-node1
```

---

### `selector`

Lọc PV theo label, dùng khi có nhiều PV cùng class và cần chọn cái phù hợp:

```yaml
spec:
  selector:
    matchLabels:
      tier: ssd
```

---

## Ví dụ theo provisioner

### Local Storage (Static Provisioning)

PV phải được admin tạo trước. PVC bind dựa trên `storageClassName` và capacity.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: local-pvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: local-storage
  resources:
    requests:
      storage: 100Gi
```

Lưu ý: vì local storage dùng `volumeBindingMode: WaitForFirstConsumer`, PVC sẽ ở `Pending` cho đến khi có Pod được schedule.

---

### NFS

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-pvc
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: nfs-storage
  resources:
    requests:
      storage: 50Gi
```

NFS hỗ trợ `ReadWriteMany` — nhiều Pod từ nhiều node có thể cùng mount PVC này.

---

### Longhorn (Dynamic Provisioning)

Không cần tạo PV trước. Khi PVC được apply, Longhorn CSI driver tự tạo Longhorn volume và PV tương ứng rồi bind ngay.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: longhorn-pvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: longhorn
  resources:
    requests:
      storage: 20Gi
```

Nếu muốn control số replica, disk type... thì tạo StorageClass riêng với các `parameters` tương ứng và trỏ `storageClassName` sang class đó.

---

## Dùng PVC trong Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
        - name: app
          image: nginx
          volumeMounts:
            - mountPath: /data
              name: storage
      volumes:
        - name: storage
          persistentVolumeClaim:
            claimName: longhorn-pvc
```

Deployment và PVC phải ở **cùng namespace**.

Khi `replicas > 1`, `accessModes` trở thành yếu tố quyết định:

| AccessMode | Behavior với nhiều replica | Dùng khi nào |
|------------|---------------------------|--------------|
| `ReadWriteOnce` (RWO) | Chỉ các Pod cùng node mới mount được — Pod ở node khác bị treo ở `ContainerCreating` | Deployment 1 replica, stateful app |
| `ReadWriteMany` (RWX) | Tất cả Pod từ mọi node đều mount được đồng thời | Deployment nhiều replica cần share volume |

**Longhorn vs NFS với Deployment nhiều replica:**

Longhorn mặc định là RWO — không phù hợp cho Deployment nhiều replica cần tất cả Pod cùng ghi vào một volume. Tuy nhiên đây không phải điểm yếu của Longhorn, vì bài toán Longhorn giải quyết là khác: **replication tự động và HA** — dữ liệu được replicate qua nhiều node, node chết thì tự failover. Điểm mạnh đó hoàn toàn độc lập với việc có bao nhiêu replica trong Deployment.

Longhorn và NFS giải quyết hai bài toán khác nhau:

| | Longhorn (RWO) | NFS (RWX) |
|---|---|---|
| Phù hợp | Deployment 1 replica, DB, stateful app cần HA | Deployment nhiều replica cần share volume |
| Điểm mạnh | Replication tự động, auto-failover, snapshot | Nhiều Pod từ nhiều node cùng mount |
| Deployment nhiều replica | Không phù hợp nếu các Pod cần share cùng một volume | Phù hợp |

---

## Mở rộng dung lượng (Volume Expansion)

Nếu StorageClass có `allowVolumeExpansion: true`, chỉ cần chỉnh `resources.requests.storage` lên cao hơn:

```yaml
spec:
  resources:
    requests:
      storage: 50Gi  # tăng từ 20Gi lên 50Gi
```

Kubernetes và CSI driver sẽ tự xử lý việc resize volume phía dưới. Chỉ mở rộng được, không thu hẹp.

---

## Tóm tắt: PVC — PV — StorageClass

```
PVC (yêu cầu của workload)
  └── storageClassName: longhorn
      └── StorageClass longhorn
          └── provisioner: driver.longhorn.io
              └── Tự tạo PV + Longhorn volume
                  └── Bind PVC ↔ PV
                      └── Pod mount được /data
```

---

## Tài liệu tham khảo

- [Persistent Volumes - Kubernetes Docs](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims)
- [PersistentVolumeClaim API Reference](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/persistent-volume-claim-v1/)
- [PersistentVolume](../persistent_volume/README.md)
- [StorageClass](../storage_classes/README.md)
- [Longhorn: Create Volumes](https://longhorn.io/docs/1.11.2/nodes-and-volumes/volumes/create-volumes/)
