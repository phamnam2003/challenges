# CloudNativePG on Kubernetes

## What is CloudNativePG?

**CloudNativePG (CNPG)** is a Kubernetes Operator that manages the full lifecycle of a PostgreSQL cluster — from initialization, replication, and failover to backup — through a single CRD: `Cluster`. Instead of writing StatefulSets + init containers + sidecar scripts to operate PostgreSQL yourself, you declare a `Cluster` resource and the operator handles all operational logic.

Under the hood, CNPG still creates Pods, PVCs, and Services. The difference from a [bare StatefulSet](../workload/statefulset/README.md#bare-statefulset-vs-operator): CNPG **understands PostgreSQL** — it knows how to configure streaming replication, detect a dead primary, promote a replica, and fence a failed node. A StatefulSet only knows how to create Pods in order.

One important point: CNPG **does not use StatefulSets**. It manages Pods directly — this gives the operator full control over which Pods to create, delete, or promote without being constrained by StatefulSet's sequential ordering.

---

## Problem It Solves

Running PostgreSQL on Kubernetes with a bare StatefulSet requires solving every operational concern yourself:

- **Replication** — you must write init containers running `pg_basebackup`, configure `primary_conninfo`, manage replication slots. CNPG handles this automatically when `instances > 1`
- **Failover** — when the primary dies, nothing happens by default. You must detect the failure, pick the best replica (least replication lag), promote it, and reconfigure remaining replicas to point at the new primary. CNPG does this in seconds
- **Fencing (split-brain prevention)** — if a node loses network but the Pod is still running, two primaries could write to two different disks. CNPG fences the old primary by patching `pg_hba.conf` to reject all connections before promoting a replica
- **Backup** — you must set up CronJobs for `pg_basebackup` or `pg_dump`, manage retention, and verify restores. CNPG lets you declare the backup schedule and destination directly in the `Cluster` spec

---

## Architecture

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

Each Pod has **two PVCs**: one for data (`PGDATA`) and one for WAL. Separating WAL onto its own volume prevents WAL bloat from filling the data disk and reduces I/O contention (WAL writes are sequential; data I/O is random).

The operator automatically creates 3 Services:

| Service | Points to |
|---------|-----------|
| `{cluster}-rw` | Always points to the primary — use for writes |
| `{cluster}-ro` | Load-balances across replicas — use for reads |
| `{cluster}-r` | All instances (primary + replicas) |

---

## Why Use Longhorn with CNPG?

PostgreSQL already replicates data at the application layer via streaming replication (`instances: 3`). If Longhorn also replicates at the storage layer, that's **double replication** — wasting disk and I/O with no added benefit.

The solution: use Longhorn with `dataLocality: strict-local` and `numberOfReplicas: "1"` — data stays on the same node as the Pod, without replicating across nodes. Result:

- **No network I/O** when reading/writing the database — critical for PostgreSQL performance
- **Dynamic provisioning** — CNPG declares the `storageClass`, Longhorn creates the PV automatically
- **Volume expansion** — expand PVCs without downtime

Trade-off: if a node dies, the Longhorn volume on that node is gone. But PostgreSQL already has 2 replicas on 2 other nodes — the operator promotes a replica and recreates the failed instance.

---

## Manifest Structure

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
type: kubernetes.io/basic-auth          # requires exactly 2 fields: username + password
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
  numberOfReplicas: "1"                 # no replication — PostgreSQL already replicates
  dataLocality: "strict-local"          # data stays on the same node as the Pod
  staleReplicaTimeout: "2880"
  fsType: "ext4"
reclaimPolicy: Retain                   # deleting the PVC does not delete the PV — protects data
volumeBindingMode: WaitForFirstConsumer # create PV after Pod is scheduled — required with strict-local
---
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: cnpg-cluster-postgresql
  namespace: cnpg-cluster
spec:
  description: "CloudNativePG Production PostgreSQL Cluster"
  instances: 3                                    # 1 primary + 2 replicas
  imageName: ghcr.io/cloudnative-pg/postgresql:18.4

  storage:
    storageClass: longhorn-cnpg
    size: 25Gi
  walStorage:                                     # separate WAL onto its own PVC
    storageClass: longhorn-cnpg
    size: 5Gi                                     # typically 10-20% of the data volume

  bootstrap:
    initdb:                                       # runs only on first cluster initialization
      database: bootstrap-cluster-db
      owner: appuser
      secret:
        name: cnpg-cluster-credentials            # Secret must exist before applying Cluster

  primaryUpdateStrategy: unsupervised             # operator handles switchover automatically on update

  affinity:
    enablePodAntiAffinity: true
    topologyKey: kubernetes.io/hostname
    podAntiAffinityType: preferred                # spread Pods across nodes without blocking scheduling
  topologySpreadConstraints:
    - maxSkew: 1
      topologyKey: kubernetes.io/hostname
      whenUnsatisfiable: ScheduleAnyway
      labelSelector:
        matchLabels:
          cnpg.io/cluster: cnpg-cluster-postgresql

  monitoring:
    enablePodMonitor: false                       # enable only if Prometheus Operator is installed
```

---

## Key Fields

### `bootstrap`

Defines how the cluster is initialized **for the first time**. After bootstrap completes, the primary is running and replicas clone themselves from the primary via streaming replication.

| Method | When to use |
|--------|-------------|
| `initdb` | New cluster — runs `initdb` to create a fresh PostgreSQL instance |
| `recovery` | Restore from backup (object store or Volume Snapshot) |
| `pg_basebackup` | Clone from an existing PostgreSQL server (including non-CNPG) |

---

### `walStorage`

WAL (Write-Ahead Log) is PostgreSQL's durability mechanism — every change is written to WAL before being applied to data files. Separating WAL onto its own PVC:

- **Prevents data disk from filling up** — WAL bloat (from long-running transactions or replication lag) does not affect the data volume
- **Reduces I/O contention** — WAL writes are sequential, data I/O is random; separate disks avoid competition

---

### `primaryUpdateStrategy`

| Strategy | Behavior |
|----------|----------|
| `unsupervised` | Operator handles everything: updates replicas first, then performs switchover (replica becomes primary, old primary restarts with new image). Brief downtime during switchover |
| `supervised` | Operator only updates replicas. You must manually trigger switchover with `kubectl cnpg promote` |

---

### `affinity` and `topologySpreadConstraints`

Combining `podAntiAffinity` with `topologySpreadConstraints` spreads Pods across nodes. If a node dies, only one instance is lost — the other two continue serving traffic.

`podAntiAffinityType: preferred` (not `required`): "spread if possible, but if there are only 2 nodes and 3 instances, still schedule the third instance on an already-occupied node." With `required`, the third Pod would stay Pending forever.

---

### `volumeBindingMode: WaitForFirstConsumer`

Required when using `strict-local`. Without it, Longhorn may create the volume on node-1 while the scheduler places the Pod on node-2 — the Pod will stay Pending because it cannot access the volume.

---

## Common Pitfalls

**`strict-local` without `WaitForFirstConsumer`:** Longhorn creates the volume on a random node, Pod is scheduled on a different node → Pending forever. Always pair `strict-local` with `WaitForFirstConsumer`.

**No separate WAL storage:** WAL and data share the same PVC. A long-running transaction or replication slot bloats WAL, fills the disk, PostgreSQL crashes. Always use `walStorage` in production.

**Secret does not exist when applying Cluster:** CNPG bootstrap fails, cluster enters an error state. Apply the Secret before the Cluster.

**`podAntiAffinityType: required` on a small cluster:** 3 instances, 2 nodes → third Pod Pending forever. Use `preferred` unless you can guarantee enough nodes.

**`enablePodMonitor: true` without Prometheus Operator installed:** The PodMonitor CRD does not exist, CNPG logs errors. Only enable when [kube-prometheus-stack](../monitor/README.md) is installed.

**No backup configured:** Replication protects against node failure. Backup protects against accidental deletion, corruption, and ransomware. These are different concerns — configure `backup` in the Cluster spec or use an external backup solution.

---

## References

- [CloudNativePG Documentation](https://cloudnative-pg.io/documentation/current/)
- [CloudNativePG Helm Chart](https://github.com/cloudnative-pg/charts)
- [Longhorn Best Practices](https://longhorn.io/docs/latest/best-practices/)
- [StatefulSet — Bare vs Operator](../workload/statefulset/README.md#bare-statefulset-vs-operator)
- [StorageClass](../storage/storage_classes/README.md)
- [PersistentVolume](../storage/persistent_volume/README.md)
