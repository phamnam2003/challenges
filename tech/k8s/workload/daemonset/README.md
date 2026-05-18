# DaemonSet in Kubernetes

## What is a DaemonSet?

**DaemonSet** is a workload controller that ensures **every node** (or a selected subset of nodes) in the cluster runs **exactly one copy** of a Pod. When a new node joins the cluster — a Pod is automatically created on it. When a node is removed — the Pod is garbage collected. Deleting the DaemonSet cleans up all the Pods it created.

Unlike a Deployment (which declares "I want N Pods somewhere in the cluster"), a DaemonSet declares "I want **exactly 1 Pod on every node**". There is no `replicas` field — the Pod count is derived from the number of eligible nodes.

DaemonSets are used for **node-local facilities** — things that must run on each node for the cluster to function: log collection, resource monitoring, network management, storage management.

---

## Problem It Solves

Some workloads must be **present on every node**, not just running "somewhere" in the cluster:

- **Log collectors** (Fluentd, Filebeat) need to mount `/var/log` on **each node** to collect logs from all containers. A Deployment might place 3 Pods on the same node and miss others entirely
- **Monitoring agents** (node-exporter, Datadog) need to read hardware/OS metrics (CPU, RAM, disk) **directly on the node**. Running on a different node yields no useful data
- **Network plugins** (Calico, Cilium) need to install network rules on **every node** for Pods to communicate. Missing one node = that node loses networking
- **Storage daemons** (Longhorn, Ceph) need to manage disks **on each node**. You cannot manage node A's disks from node B

Deployment cannot solve this because: it doesn't guarantee every node gets a Pod, it doesn't auto-deploy when a new node joins, and when a node is removed it reschedules the Pod elsewhere (wrong behavior — DaemonSet simply garbage collects it).

---

## How DaemonSet Works

```
New node joins cluster
  → DaemonSet controller detects it
    → Creates a Pod with nodeAffinity targeting that node
      → Default scheduler binds the Pod to the node
```

1. **DaemonSet controller** (inside kube-controller-manager) watches node events (added, removed, label changes)
2. For each eligible node, the controller checks whether a matching Pod already exists — if not, it creates one
3. If a node becomes ineligible (label changed, drained), the controller deletes the Pod on that node

**Automatic toleration:** the DaemonSet controller automatically adds the toleration `node.kubernetes.io/unschedulable:NoSchedule` to its Pods. This prevents a deadlock: the network plugin can't start because the node isn't ready, and the node isn't ready because the network plugin isn't running.

---

## Manifest Structure

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: node-exporter
  namespace: monitoring
  labels:
    app: node-exporter
spec:
  selector:
    matchLabels:
      app: node-exporter

  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 0

  template:
    metadata:
      labels:
        app: node-exporter
    spec:
      tolerations:                                    # to run on control plane nodes
        - key: node-role.kubernetes.io/control-plane
          operator: Exists
          effect: NoSchedule

      containers:
        - name: node-exporter
          image: prom/node-exporter:v1.8.1
          ports:
            - containerPort: 9100
              hostPort: 9100                          # binds directly to the node IP
          resources:
            requests:
              cpu: 100m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 128Mi
          volumeMounts:
            - name: proc
              mountPath: /host/proc
              readOnly: true
            - name: sys
              mountPath: /host/sys
              readOnly: true

      volumes:
        - name: proc
          hostPath:
            path: /proc
        - name: sys
          hostPath:
            path: /sys
```

This deploys Prometheus node-exporter — one of the most common DaemonSets. It mounts `/proc` and `/sys` to read system metrics and uses `hostPort` so Prometheus can scrape directly via `<NodeIP>:9100`.

---

## Node Selection

By default, a DaemonSet creates a Pod on **every node**. This can be narrowed down with:

### nodeSelector (simple)

```yaml
spec:
  template:
    spec:
      nodeSelector:
        disk: ssd
```

Only creates Pods on nodes labeled `disk=ssd`.

### nodeAffinity (more flexible)

```yaml
spec:
  template:
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: node-type
                    operator: In
                    values: ["worker", "gpu"]
```

### Tolerations for control plane

Control plane nodes are tainted with `node-role.kubernetes.io/control-plane:NoSchedule` — DaemonSet Pods **will not run there** unless a toleration is added:

```yaml
tolerations:
  - key: node-role.kubernetes.io/control-plane
    operator: Exists
    effect: NoSchedule
```

Most infrastructure DaemonSets (CNI, logging, monitoring) should tolerate control plane nodes to ensure full cluster coverage.

---

## Update Strategy

### RollingUpdate (default)

Old Pods are deleted and new Pods are created sequentially, controlled by 2 parameters:

```yaml
updateStrategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: 1    # max Pods unavailable at the same time (default: 1)
    maxSurge: 0          # max nodes that can run both old and new Pod simultaneously (default: 0)
```

**Zero-downtime update** — creates the new Pod first, waits for Ready, then deletes the old one:

```yaml
rollingUpdate:
  maxSurge: 1
  maxUnavailable: 0
```

### OnDelete

Kubernetes does not auto-update. You must **manually delete** each Pod — the controller recreates it with the new config. Use for extremely sensitive DaemonSets (CNI plugins, storage drivers) that require manual verification at each node.

---

## Communicating with DaemonSet Pods

| Pattern | How it works | Best for |
|---------|-------------|----------|
| **Push** | Pod pushes data outbound (e.g., Fluentd → Elasticsearch) | No inbound traffic needed |
| **hostPort** | Binds a port on the node IP, clients connect via `<NodeIP>:<port>` | Clients know node IPs (e.g., Prometheus scraping) |
| **Headless Service** | DNS returns individual Pod IPs | Need to discover all Pods |
| **ClusterIP Service** | Load-balances via a single virtual IP | Clients don't need to target a specific node |

---

## Common Use Cases

### Log collection
Fluentd, Fluent Bit, Filebeat — mount `hostPath` `/var/log` to collect container and node logs, forward to Elasticsearch/Loki.

### Monitoring
Prometheus node-exporter — exposes hardware/OS metrics (CPU, memory, disk, network) via `/metrics`. Datadog agent — full-stack monitoring including APM and log collection.

### Network plugins (CNI)
Calico, Cilium, kube-proxy — install network rules (iptables/eBPF) on each node. These are the **most critical** DaemonSets — without them, Pods cannot communicate.

### Storage daemons
Longhorn, Ceph (rook-ceph) — manage disks on each node, providing distributed storage for the cluster.

---

## Common Pitfalls

**Resources multiply with node count.** A DaemonSet requesting 500Mi memory × 10 nodes = 5Gi. × 100 nodes = 50Gi. Always calculate cluster-wide cost before deploying — especially as the cluster scales.

**Missing tolerations = Pod not scheduled.** The #1 cause of "DaemonSet not running on all nodes". When adding new node pools with taints (GPU, maintenance), you must update tolerations on essential DaemonSets (CNI, CSI, logging).

**Missing `resources.requests`.** The scheduler cannot calculate capacity accurately → nodes get overcommitted → OOM kills. Infrastructure DaemonSets (CNI, kube-proxy) should set `requests == limits` for **Guaranteed** QoS, preventing eviction when the node runs out of resources.

**`hostPort` conflict.** Two DaemonSets using the same `hostPort` on 1 node → only 1 Pod can bind, the other fails. Each port can only be used once per node.

**Selector is immutable.** `.spec.selector` is **immutable** after creation. To change it, you must delete and recreate the DaemonSet.

**Missing `priorityClassName`.** When a node runs out of resources, DaemonSet Pods can be evicted if they don't have high priority. Infrastructure DaemonSets should use `system-node-critical` or `system-cluster-critical`.

**Running on control plane without resource limits.** A memory-hungry DaemonSet Pod on the control plane can OOM kill etcd or kube-apiserver — affecting the entire cluster. Always set limits when tolerating control plane nodes.

---

## References

- [DaemonSet - Kubernetes Docs](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)
- [Perform a Rolling Update on a DaemonSet](https://kubernetes.io/docs/tasks/manage-daemon/update-daemon-set/)
- [Taints and Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)
