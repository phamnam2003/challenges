# PersistentVolume in Kubernetes

## What is a PersistentVolume?

**PersistentVolume (PV)** is a cluster-level resource that represents a piece of physical storage — a local disk, NFS share, or any other storage backend. A PV has a lifecycle independent of any Pod: the Pod can be deleted but the data in the PV persists.

PVs act as an abstraction layer: Pods do not need to know what the underlying storage is — they just declare their needs via a PVC and Kubernetes handles the rest.

---

## Problem It Solves

Without PVs, each Pod would mount directly to a storage backend (NFS path, host directory, etc.), which causes:

- **Tight coupling** between Pod specs and storage infrastructure — changing storage means updating every workload
- **No config sharing** — multiple Pods using the same NFS share must copy-paste the same configuration
- **No data protection** — nothing prevents a Pod deletion from taking the volume with it

PVs cleanly separate *storage provisioning* (the admin's responsibility) from *storage consumption* (the workload's responsibility).

---

## PV Lifecycle

```
Provisioning → Binding → Using → Reclaiming
```

### 1. Provisioning

**Static:** Admin manually creates PVs with full storage backend details.

**Dynamic:** Kubernetes automatically creates PVs when a PVC requests a `storageClassName` whose provisioner supports dynamic provisioning (see [StorageClass](../storage_classes/README.md)).

### 2. Binding

The control plane finds a PV matching the PVC's `storageClassName`, `accessModes`, and `capacity`. The bind relationship is **1-to-1**: one PV can only be bound to one PVC at a time.

If no matching PV exists, the PVC stays in `Pending` until a suitable PV becomes available.

### 3. Using

The Pod mounts the PVC like any other volume. The PV belongs to that workload until the PVC is deleted.

Kubernetes enforces **Storage Object in Use Protection**: a PVC actively used by a Pod will not actually be deleted until the Pod releases it — preventing accidental data loss.

### 4. Reclaiming

When a PVC is deleted, the PV moves to `Released` state. What happens next depends on `reclaimPolicy`:

| Policy | Behavior |
|--------|----------|
| `Retain` | PV and data are kept; admin must clean up manually |
| `Delete` | PV and the underlying physical storage are deleted automatically |
| `Recycle` | Deprecated — do not use |

---

## PV Phase (Status)

| Phase | Meaning |
|-------|---------|
| `Available` | PV is ready, not yet bound to any PVC |
| `Bound` | Bound to a PVC |
| `Released` | PVC was deleted; PV is awaiting reclamation |
| `Failed` | Reclamation failed |

---

## Manifest Structure

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-example
spec:
  capacity:
    storage: 50Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt/disks/ssd1
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - node1
```

---

## Key Fields

### `capacity`

Storage size of the PV:

```yaml
capacity:
  storage: 100Gi
```

A PVC can only bind to a PV whose capacity is **equal to or greater than** the requested amount.

---

### `accessModes`

Controls how the volume can be mounted:

| Mode | Short | Meaning |
|------|-------|---------|
| `ReadWriteOnce` | RWO | One node can mount read/write |
| `ReadOnlyMany` | ROX | Multiple nodes can mount read-only |
| `ReadWriteMany` | RWX | Multiple nodes can mount read/write |
| `ReadWriteOncePod` | RWOP | Only one Pod can mount read/write |

Not every storage backend supports all modes. Local disk only supports RWO; NFS supports RWX.

---

### `volumeMode`

| Value | Behavior |
|-------|----------|
| `Filesystem` (default) | Mounted as a regular directory |
| `Block` | Exposed as a raw block device — for databases that manage their own I/O |

---

### `persistentVolumeReclaimPolicy`

See the Reclaiming table above.

---

### `storageClassName`

Links the PV to a StorageClass. The PVC must declare the same `storageClassName` to match. If left empty, the PV only binds with PVCs that also have no `storageClassName`.

---

### `nodeAffinity`

**Required** for local storage. Specifies which node physically holds this volume. Kubernetes ensures any Pod using the PVC is scheduled onto that node.

```yaml
nodeAffinity:
  required:
    nodeSelectorTerms:
      - matchExpressions:
          - key: kubernetes.io/hostname
            operator: In
            values:
              - node1
```

---

## Volume Types on Bare Metal

### Local

Direct access to a disk, partition, or directory on a node. Highest I/O performance since there is no network hop. `nodeAffinity` is mandatory.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv-node1
spec:
  capacity:
    storage: 200Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt/disks/ssd1
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - node1
```

---

### NFS

Mounts from an NFS server over the network. Supports `ReadWriteMany` — multiple Pods across multiple nodes can mount the same volume simultaneously.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
spec:
  capacity:
    storage: 500Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  storageClassName: nfs-storage
  nfs:
    server: 192.168.1.100
    path: /exported/k8s
```

No `nodeAffinity` needed — NFS is accessible from every node in the cluster.

---

### Longhorn

Longhorn primarily uses **dynamic provisioning** — creating a PVC with `storageClassName: longhorn` is all that's needed; Longhorn creates the PV automatically. This is the recommended flow for all normal workloads (see [StorageClass](../storage_classes/README.md)).

There are two cases where a static PV must be created manually, as documented by Longhorn:

**1. Binding to an existing Longhorn volume via Longhorn UI**

The Longhorn UI can create a PV/PVC for an existing detached volume. The default StorageClass for this flow is `longhorn-static`. No manual PV YAML is needed — the UI handles it.

**2. Restoring from backup via CLI (Custom Resource)**

When restoring by manually creating a Volume CR, a matching PV must also be created. Per Longhorn docs: the `volumeHandle` in the PV **must exactly match** the `metadata.name` of the Volume CR.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: restored-pv
spec:
  capacity:
    storage: 50Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: longhorn-static
  csi:
    driver: driver.longhorn.io
    fsType: ext4
    volumeHandle: restored-volume-name  # must match metadata.name of the Volume CR
```

---

## PV — PVC — Pod Relationship

```
Admin creates PV (or StorageClass provisions it automatically)
  └── Dev creates PVC (declares storageClassName, accessModes, storage size)
      └── K8s binds PVC ↔ PV
          └── Pod declares a volume referencing the PVC
              └── Container mounts the volume at a path inside the container
```

Example Pod using a PVC:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-pod
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
        claimName: my-pvc
```

---

## Static vs Dynamic Provisioning on Bare Metal

| | Static | Dynamic |
|---|---|---|
| Who creates the PV | Admin, manually | Provisioner, automatically |
| Best for | Local storage, self-managed NFS | Longhorn, NFS CSI driver |
| Control | High — admin defines each PV | Lower — StorageClass decides |
| Maintenance overhead | High as cluster scales | Fully automated |

With **local storage**, static provisioning is required because `kubernetes.io/no-provisioner` does not create PVs automatically. With **Longhorn** and the **NFS CSI driver**, dynamic provisioning works out of the box.

---

## References

- [Persistent Volumes - Kubernetes Docs](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Storage Classes](../storage_classes/README.md)
- [Local Persistent Volumes](https://kubernetes.io/docs/concepts/storage/volumes/#local)
- [Longhorn: Create Volumes](https://longhorn.io/docs/1.11.2/nodes-and-volumes/volumes/create-volumes/)
- [Longhorn: Restore from Backup](https://longhorn.io/docs/1.11.2/snapshots-and-backups/backup-and-restore/restore-from-a-backup/)
