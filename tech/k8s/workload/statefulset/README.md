# StatefulSet in Kubernetes

## What is a StatefulSet?

**StatefulSet** is a workload controller for managing **stateful** applications — applications whose instances are not interchangeable and need to retain their own identity and data throughout their lifecycle. StatefulSet guarantees each Pod a **fixed name**, **dedicated DNS**, and **dedicated storage** — even if the Pod is deleted, crashes, or gets rescheduled to a different node.

StatefulSet works similarly to Deployment in that both manage Pods from the same `spec.template`. The difference: Deployment treats every Pod as **interchangeable** (any Pod can be deleted and replaced, new Pods get random names), while StatefulSet assigns each Pod a **sticky identity** that never changes.

---

## Problem It Solves

Stateful applications (databases, message queues, distributed systems) have requirements that Deployment cannot fulfill:

- **Database clusters** need to distinguish primary from replicas — the primary (pod-0) must initialize first, then replicas connect to it. Deployment starts Pods in parallel with no ordering guarantees
- **Cluster members** need to discover each other through stable addresses — Kafka brokers must know the exact DNS of every other broker. Deployment assigns random names that change on restart
- **Each instance needs its own data** — a PostgreSQL primary and its replicas cannot share the same data directory. Deployment only offers shared or ephemeral storage

StatefulSet solves all of this through 3 guarantees: **stable network identity**, **stable storage**, and **ordered lifecycle**.

---

## 3 Core Guarantees

### 1. Stable Network Identity

Each Pod gets a fixed name following the pattern `{statefulset-name}-{ordinal}`: `postgres-0`, `postgres-1`, `postgres-2`.

StatefulSet **requires** a Headless Service (`clusterIP: None`). A regular Service load-balances to a random Pod — a Headless Service returns the IP of **each individual Pod**, allowing direct connections to a specific Pod via DNS.

With StatefulSet `postgres`, Headless Service `postgres`, namespace `default`:

```
postgres-0.postgres.default.svc.cluster.local  →  IP of postgres-0
postgres-1.postgres.default.svc.cluster.local  →  IP of postgres-1
postgres.default.svc.cluster.local             →  returns all Pod IPs
```

When `postgres-1` crashes and gets rescheduled to another node — it is still `postgres-1`, with the same DNS name and the same PVC. **IP may change, DNS does not** — always use DNS names instead of IPs.

### 2. Stable Storage (volumeClaimTemplates)

Each Pod automatically gets its own [PVC](../../../storage/persistent_volume_claims/README.md), named following the pattern `{template-name}-{statefulset}-{ordinal}`:

```
data-postgres-0    →  dedicated PVC for postgres-0
data-postgres-1    →  dedicated PVC for postgres-1
```

**The key point:** scaling down does not delete PVCs. Scale from 3 to 1, and `data-postgres-1` and `data-postgres-2` still exist. Scale back up → the Pod reattaches its old PVC, data intact. This is by design to protect data.

This behavior can be customized:

```yaml
persistentVolumeClaimRetentionPolicy:
  whenDeleted: Retain      # when StatefulSet is deleted: Retain (default) | Delete
  whenScaledDown: Retain   # when scaling down:           Retain (default) | Delete
```

**Note:** `volumeClaimTemplates` cannot be modified while a StatefulSet exists. You must delete the StatefulSet with `--cascade=orphan` (keeps Pods running), modify the PVCs, then recreate.

### 3. Ordered Lifecycle

Default behavior (`podManagementPolicy: OrderedReady`):

- **Scale up:** creates sequentially 0 → 1 → 2, each Pod must be Running + Ready before the next is created
- **Scale down:** deletes in reverse 2 → 1 → 0, each Pod must fully terminate before the next is deleted
- **Update:** updates in reverse ordinal order (highest → lowest)

This ordering matters because many stateful systems have a role hierarchy: a PostgreSQL primary (pod-0) must be ready before replicas (pod-1, pod-2) can connect for streaming replication.

If all nodes are peers (Cassandra, Elasticsearch) — ordering is unnecessary:

```yaml
podManagementPolicy: Parallel   # creates/deletes all Pods simultaneously
```

`Parallel` only affects scaling, **not updates** — rolling updates still proceed in order.

---

## Manifest Structure

```yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  clusterIP: None                         # Headless — required for StatefulSet
  selector:
    app: postgres
  ports:
    - port: 5432
      name: postgres
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  serviceName: "postgres"                 # references the Headless Service above
  replicas: 3
  selector:
    matchLabels:
      app: postgres

  podManagementPolicy: OrderedReady       # OrderedReady | Parallel
  updateStrategy:
    type: RollingUpdate                   # RollingUpdate | OnDelete
    rollingUpdate:
      partition: 0                        # only updates Pods with ordinal >= this value
      maxUnavailable: 1

  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Retain
    whenScaledDown: Retain

  template:
    metadata:
      labels:
        app: postgres
    spec:
      terminationGracePeriodSeconds: 60   # databases need time to flush WAL — don't set this low
      containers:
        - name: postgres
          image: postgres:17
          ports:
            - containerPort: 5432
          env:
            - name: PGDATA
              value: /var/lib/postgresql/data/pgdata
          volumeMounts:
            - name: data
              mountPath: /var/lib/postgresql/data

  volumeClaimTemplates:                   # auto-creates PVCs: data-postgres-0, data-postgres-1...
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: "longhorn"
        resources:
          requests:
            storage: 10Gi
```

---

## Update Strategy

### RollingUpdate (default)

Updates Pods in reverse ordinal order. Each Pod must be Ready before the next one is updated.

**`partition`** — canary deployment for StatefulSet. Set `partition: 2` with 3 replicas → only `postgres-2` gets updated, `postgres-0` and `postgres-1` stay on the old template. Gradually decrease `partition` to roll out completely.

### OnDelete

Kubernetes does not auto-update Pods. You must **manually delete** each Pod — the controller recreates it with the new config. Useful when you need to run migrations or manual steps between updates (e.g., schema migration on primary first, then update replicas).

---

## Common Pitfalls

**Node failure — Pods don't self-heal:** when a node loses contact, the Pod is stuck in `Terminating` indefinitely. Kubernetes **will not create a replacement Pod** because StatefulSet guarantees no two Pods with the same identity run simultaneously (to prevent split-brain). You must force-delete the Pod (`kubectl delete pod postgres-1 --force --grace-period=0`) or remove the node from the cluster.

**`terminationGracePeriodSeconds` too low:** databases need to flush buffers/WAL before shutting down. Setting this too low (or `0`) → data corruption. Production should be 60s+.

**Missing `resources.requests`:** an OOM kill on a database Pod can corrupt data. Always set requests and limits.

**PVCs are not automatically backed up:** PVCs only hold data on disk, they are not backups. You need a separate backup strategy (velero, volume snapshots, pg_dump...).

**HPA + StatefulSet:** works, but be careful — automatically scaling down database replicas can cause data loss without proper drain logic.

---

## Bare StatefulSet vs Operator

StatefulSet only handles the **infrastructure**: creating Pods in order, attaching the right PVC, maintaining the right DNS. But it **does not understand the application** — it doesn't know how to initialize a PostgreSQL replica, when to trigger failover, or how to run backups. All of that logic must be manually written through init containers, sidecar scripts, or manual operations.

An **Operator** is a pattern **within Kubernetes** — it uses Custom Resource Definitions (CRDs) and custom controllers running inside the cluster to extend K8s capabilities. An Operator understands the **domain logic** of a specific application. Under the hood it still creates StatefulSets/Pods/PVCs/Services, but adds an operational logic layer that bare StatefulSets don't have.

For example, the CloudNativePG operator for PostgreSQL:
- Automatically configures streaming replication between primary and replicas
- Primary dies → automatically promotes a replica to primary
- Runs backups (pg_basebackup) on a defined schedule
- Handles database-safe rolling updates

| | Bare StatefulSet | Operator |
|---|---|---|
| Who manages Pods | StatefulSet controller | Operator controller (still creates StatefulSet underneath) |
| Init replicas | Write your own init container / script | Operator handles it |
| Failover | Manual or write your own logic | Automatic |
| Backup | Self-setup (cronjob + pg_dump) | Declare a schedule, operator handles the rest |
| Best for | Learning, dev/staging, simple apps | Production databases |

For production databases, use Operators (CloudNativePG, Percona Operator, MongoDB Community Operator) rather than writing StatefulSets from scratch.

---

## References

- [StatefulSets - Kubernetes Docs](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [StatefulSet Basics Tutorial](https://kubernetes.io/docs/tutorials/stateful-application/basic-stateful-set/)
- [Run a Replicated Stateful Application](https://kubernetes.io/docs/tasks/run-application/run-replicated-stateful-application/)
- [PersistentVolume](../../../storage/persistent_volume/README.md)
- [PersistentVolumeClaim](../../../storage/persistent_volume_claims/README.md)
