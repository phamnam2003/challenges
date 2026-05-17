# StorageClass in Kubernetes

## What is a StorageClass?

**StorageClass** is a Kubernetes object that allows cluster admins to describe different tiers or classes of storage available in the cluster. It acts as a template so Kubernetes can automatically create PersistentVolumes (PVs) in response to PersistentVolumeClaim (PVC) requests — this mechanism is called **dynamic provisioning**.

Before StorageClass existed, admins had to manually create PVs before a PVC could bind to them. StorageClass eliminates that manual step entirely.

---

## Problem It Solves

### The Problem with Static Provisioning

Without StorageClass, the storage allocation workflow looks like this:

```
Admin manually creates PV → Dev creates PVC → K8s binds PVC to a matching PV
```

This breaks down when:
- The cluster has many applications needing storage with different sizes and configurations
- Admins must pre-create a large pool of PVs or continuously add more on demand
- There is no systematic way to distinguish "fast" (SSD), "slow" (HDD), or "replicated" storage

### Solution: Dynamic Provisioning with StorageClass

```
Dev creates PVC (declares storageClassName) → K8s calls provisioner → PV created automatically → bind
```

StorageClass enables:
- **Automation** of PV creation — no manual admin intervention required
- **Storage classification** by quality and policy — devs just pick the right class
- **Standardization** of storage configuration across the entire cluster

---

## Manifest Structure

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: driver.longhorn.io
reclaimPolicy: Delete
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
parameters:
  numberOfReplicas: "3"
  fsType: "ext4"
```

---

## Key Fields

### `provisioner` (required)

Determines which plugin performs the actual volume creation.

---

### `reclaimPolicy`

Controls what happens to the PV when its PVC is deleted:

| Value | Behavior |
|-------|----------|
| `Delete` (default) | Deletes both the PV and the underlying physical storage when PVC is deleted |
| `Retain` | Keeps the PV and data intact — requires manual admin cleanup |

Use `Retain` for critical data that needs audit trails or recovery after accidental PVC deletion.

---

### `volumeBindingMode`

Controls **when** the PV is created and bound:

**`Immediate`** (default):
- PV is created and bound as soon as the PVC is created, regardless of whether any Pod uses it yet
- Risk with local storage: PV gets provisioned on node A but the Pod gets scheduled to node B

**`WaitForFirstConsumer`**:
- Delays PV creation until a Pod that actually needs this PVC is scheduled
- K8s knows which node the Pod lands on → creates the PV with the correct topology
- **Required** for local storage; recommended for NFS in multi-node clusters

---

### `allowVolumeExpansion`

```yaml
allowVolumeExpansion: true
```

Allows resizing the PV by editing `resources.requests.storage` in the PVC. Volumes can only be expanded, not shrunk.

---

### `mountOptions`

Mount options passed down when the PV is mounted into a Pod. Kubernetes does not validate these — if the driver does not support an option, the mount will fail.

---

### `parameters`

Provisioner-specific configuration passed directly to the plugin. Each provisioner has its own set of parameters.

---

## Default StorageClass

```yaml
annotations:
  storageclass.kubernetes.io/is-default-class: "true"
```

PVCs that do not declare a `storageClassName` will automatically use the default. A cluster should have only **one** default StorageClass.

---

## Provisioners on Bare Metal

### 1. Local Storage

Uses a disk, partition, or directory directly on a node. There is no dynamic provisioning — admins still create PVs manually. However, using `no-provisioner` ensures binding is delayed until the Pod is scheduled, preventing node mismatch.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
```

The corresponding PV must declare `nodeAffinity` to specify which node holds the volume:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv-node1
spec:
  capacity:
    storage: 100Gi
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

**Best for:** databases requiring high I/O, workloads tightly coupled to a specific node.  
**Limitation:** no replication — if the node goes down, the data becomes inaccessible.

---

### 2. NFS

Uses a shared NFS server, allowing multiple Pods across multiple nodes to mount the same volume (`ReadWriteMany`). Requires installing the [NFS CSI driver](https://github.com/kubernetes-csi/csi-driver-nfs) or [nfs-subdir-external-provisioner](https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner).

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: nfs-storage
provisioner: nfs.csi.k8s.io
reclaimPolicy: Delete
allowVolumeExpansion: true
volumeBindingMode: Immediate
mountOptions:
  - nfsvers=4.1
  - nconnect=8
parameters:
  server: 192.168.1.100
  share: /exported/k8s
```

**Best for:** shared config, static assets, workloads requiring `ReadWriteMany`.  
**Limitation:** the NFS server is a single point of failure without HA; I/O is lower than local disk.

---

### 3. Longhorn

Longhorn is a distributed block storage system that runs inside the cluster and manages replication across nodes automatically. Provisioner: `driver.longhorn.io`.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: longhorn
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: driver.longhorn.io
allowVolumeExpansion: true
reclaimPolicy: Delete
volumeBindingMode: Immediate
parameters:
  numberOfReplicas: "3"        # number of data copies
  staleReplicaTimeout: "2880"  # minutes before a failed replica is removed
  fsType: "ext4"
  dataLocality: "best-effort"  # prefer placing a replica on the same node as the Pod
```

Notable `parameters`:

| Parameter | Description |
|-----------|-------------|
| `numberOfReplicas` | Number of data replicas, default `3` |
| `dataLocality` | `disabled` / `best-effort` / `strict-local` — controls replica placement relative to the Pod |
| `diskSelector` | Restrict provisioning to disks with matching tags (e.g. `ssd`) |
| `nodeSelector` | Restrict provisioning to nodes with matching tags |
| `encrypted` | Enable volume encryption |
| `fsType` | Filesystem type: `ext4` or `xfs` |

**Best for:** stateful workloads (databases, message queues) requiring HA and automatic failover.  
**Limitation:** consumes extra CPU/RAM/disk for replication; needs at least 3 nodes for `numberOfReplicas: 3` to be effective.

---

## Quick Comparison

| | Local Storage | NFS | Longhorn |
|---|---|---|---|
| Dynamic provisioning | No | Yes | Yes |
| Access mode | RWO | RWX | RWO (RWX via NFS mode) |
| Replication | No | No (depends on NFS server) | Yes (built-in) |
| Automatic failover | No | No | Yes |
| I/O performance | Highest | Medium | Good (depends on replica count) |
| Operational complexity | Low | Low–Medium | Medium–High |
| Best for | High I/O, node-pinned workloads | Shared files, RWX | Stateful apps needing HA |

---

## How Dynamic Provisioning Works

```
Dev creates PVC (storageClassName: "longhorn")
  └── K8s finds StorageClass "longhorn"
      └── volumeBindingMode: Immediate → proceed right away
          └── Calls Longhorn CSI driver to create PV with 3 replicas
              └── Bind PVC ↔ PV
                  └── Pod mounts the volume
```

---

## References

- [Storage Classes - Kubernetes Docs](https://kubernetes.io/docs/concepts/storage/storage-classes/)
- [Local Persistent Volumes](https://kubernetes.io/docs/concepts/storage/volumes/#local)
- [NFS CSI Driver](https://github.com/kubernetes-csi/csi-driver-nfs)
- [Longhorn StorageClass Parameters](https://longhorn.io/docs/latest/references/storage-class-parameters/)
