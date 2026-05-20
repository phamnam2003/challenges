# PersistentVolumeClaim

**PVC** is a storage request from a workload — the developer declares *how much* and *what access mode* is needed, and Kubernetes finds or creates a matching PV and binds it. PVC is namespace-scoped: only Pods in the same namespace can use it.

Without PVC, Pods would embed storage backend details directly in their spec — developers would need to know the infrastructure, and any storage change would require updating every workload. PVC creates an abstraction layer: developers declare what they need, admins handle the provisioning.

---

## Lifecycle

```
PVC created → K8s finds matching PV → Bound → Pod uses it → PVC deleted → PV reclaimed
```

| Phase | Meaning |
|-------|---------|
| `Pending` | No matching PV found yet |
| `Bound` | Successfully bound to a PV |
| `Terminating` | Deletion requested but a Pod is still mounting it |

Kubernetes will not delete a PVC while a Pod is mounting it — the finalizer `kubernetes.io/pvc-protection` holds the PVC in `Terminating` until the Pod releases it.

---

## Manifest

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem       # Filesystem (default) | Block
  storageClassName: longhorn
  resources:
    requests:
      storage: 10Gi
```

---

## Key Fields

### `accessModes`

| Mode | Short | Scope |
|------|-------|-------|
| `ReadWriteOnce` | RWO | One **node** read-write |
| `ReadOnlyMany` | ROX | Many nodes read-only |
| `ReadWriteMany` | RWX | Many nodes read-write |
| `ReadWriteOncePod` | RWOP | One **Pod** read-write |

RWO and RWOP are easy to confuse: RWO is per-node (multiple Pods on the same node can all mount it), RWOP is per-Pod (exactly one Pod, regardless of node).

### `storageClassName`

| Value | Behavior |
|-------|---------|
| Class name with provisioner (e.g. `longhorn`) | Dynamic provisioning — PV created automatically |
| Class name without provisioner (e.g. `local-storage`) | K8s looks for a static PV with the same class |
| `""` (empty string) | Only binds to PVs with no storageClass |
| Omitted | Uses the cluster's default StorageClass |

### `resources.requests.storage`

Minimum capacity required. PVC only binds to a PV whose capacity is **equal to or greater than** this value.

### `volumeName` and `selector`

```yaml
spec:
  volumeName: local-pv-node1       # bind directly to a specific PV, skips auto-matching

  selector:
    matchLabels:
      tier: ssd                    # filter PVs by label when multiple PVs share the same class
```

---

## Examples by Provisioner

### Local Storage (Static)

PV must be created by an admin first. PVC stays `Pending` until a Pod is scheduled because local storage uses `volumeBindingMode: WaitForFirstConsumer`.

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

### Longhorn (Dynamic)

No PV needs to be created in advance — the CSI driver creates the volume and binds it when the PVC is applied.

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

To control replica count, disk type, etc. — create a dedicated StorageClass with the corresponding `parameters`.

---

## Using PVC in a Deployment

```yaml
volumes:
  - name: storage
    persistentVolumeClaim:
      claimName: longhorn-pvc
```

The Deployment and PVC must be in the **same namespace**. When `replicas > 1`, `accessModes` becomes the deciding factor:

| AccessMode | Behavior with multiple replicas |
|------------|---------------------------------|
| RWO | Pods on the same node: works fine. Pods on different nodes: stuck at `ContainerCreating` — `Multi-Attach error: volume already exclusively attached to one node` |
| RWOP | Second Pod fails regardless of whether it shares a node |
| RWX | All Pods from any node can mount simultaneously |

Longhorn (RWO) and NFS (RWX) solve different problems:

| | Longhorn | NFS |
|---|---|---|
| AccessMode | RWO | RWX |
| Strengths | Automatic replication, auto-failover, snapshots | Multiple Pods across nodes mounting simultaneously |
| Best for | Single replica, databases, stateful apps needing HA | Multi-replica Deployments sharing a volume |

---

## Volume Expansion

If the StorageClass has `allowVolumeExpansion: true`, increasing `resources.requests.storage` is enough — the CSI driver resizes the underlying volume automatically. Expansion only, no shrinking.

---

## PVC — PV — StorageClass

```
PVC (workload declares its need)
  └── storageClassName: longhorn
      └── StorageClass longhorn
          └── provisioner: driver.longhorn.io
              └── Creates PV + Longhorn volume
                  └── Binds PVC ↔ PV → Pod mounts /data
```

---

## References

- [Persistent Volumes - Kubernetes Docs](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims)
- [PersistentVolumeClaim API Reference](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/persistent-volume-claim-v1/)
- [PersistentVolume](../persistent_volume/README.md)
- [StorageClass](../storage_classes/README.md)
- [Longhorn: Create Volumes](https://longhorn.io/docs/1.11.2/nodes-and-volumes/volumes/create-volumes/)
