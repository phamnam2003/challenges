# Autoscaling in Kubernetes

Kubernetes provides three main autoscaling mechanisms, each operating at a different layer:

| Mechanism | Layer | Goal |
|---|---|---|
| **HPA** | Pod | Change the number of Pod replicas |
| **VPA** | Container | Adjust CPU/Memory requests of containers |
| **Cluster Autoscaler** | Node | Add/remove Nodes in the cluster |

---

## Prerequisites — Metrics Server

HPA requires **Metrics Server** to be running in the cluster to read CPU and memory usage from nodes and pods. Without it, HPA cannot function for resource metrics.

### Install with Helm

```bash
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm repo update
```

```bash
helm install metrics-server metrics-server/metrics-server \
  --namespace kube-system
```

For **local clusters** (kind, minikube, kubeadm with self-signed certs) the kubelet TLS certificate is not trusted by default — add `--kubelet-insecure-tls`:

```bash
helm install metrics-server metrics-server/metrics-server \
  --namespace kube-system \
  --set args={--kubelet-insecure-tls}
```

### Verify

```bash
kubectl get deployment metrics-server -n kube-system
kubectl top nodes
kubectl top pods -A
```

`kubectl top nodes` returning data confirms Metrics Server is working. HPA will start functioning within one scrape interval (~15 seconds).

---

## HPA — Horizontal Pod Autoscaler

**Official docs:** https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/

### How it works

- HPA is a **control loop** that runs periodically (default every 15 seconds), continuously comparing actual metrics against the target.
- When a metric exceeds the threshold → increase replicas; when below → decrease replicas.
- Formula for desired replica count:

```
desiredReplicas = ceil[currentReplicas × (currentMetricValue / desiredMetricValue)]
```

- HPA only works with workloads that have a `scale` subresource: `Deployment`, `StatefulSet`, `ReplicaSet`.

### Metrics

HPA supports 3 metric types:

- **Resource metrics** (`cpu`, `memory`): sourced from Metrics Server (must be installed separately). Can use `Utilization` (% of request) or `AverageValue` (absolute value).
- **Custom metrics**: custom metrics from Prometheus or other adapters, accessed via the Custom Metrics API.
- **External metrics**: metrics from systems outside the cluster (queue length, HTTP RPS...).

> Requirement: containers **must declare `resources.requests`** for HPA to calculate `Utilization`.

### Configuration

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70   # scale when average CPU exceeds 70%
  - type: Resource
    resource:
      name: memory
      target:
        type: AverageValue
        averageValue: 500Mi
```

### Scale behavior (preventing flapping)

From `autoscaling/v2`, you can control scale-up and scale-down speed to avoid continuous oscillation:

```yaml
spec:
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300   # wait 5 minutes before scaling down
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60              # reduce at most 10% of replicas per 60 seconds
    scaleUp:
      stabilizationWindowSeconds: 0    # scale up immediately
      policies:
      - type: Pods
        value: 4
        periodSeconds: 60              # add at most 4 pods per 60 seconds
```

- **`stabilizationWindowSeconds`**: the time window HPA looks back into metric history before deciding to scale — prevents premature scaling on temporary metric spikes.
- `scaleDown` defaults to `stabilizationWindowSeconds: 300`; `scaleUp` defaults to `0`.

### Important notes

- HPA and the `replicas` field in a Deployment manifest can conflict when using GitOps — remove `replicas` from the Deployment manifest or use the `argocd.argoproj.io/managed-fields` annotation.
- Do not use HPA and VPA (mode `Auto`/`Recreate`) on the same metric (CPU/Memory) — the two controllers will conflict.
- HPA cannot scale to 0 (use KEDA for that).

---

## VPA — Vertical Pod Autoscaler

**Official docs:** https://kubernetes.io/docs/concepts/workloads/autoscaling/vertical-pod-autoscale/

> VPA is not built into Kubernetes — it must be installed separately from `kubernetes/autoscaler`.

### How it works

VPA consists of 3 components:

- **Recommender**: continuously observes container CPU/Memory usage and computes recommended `request`/`limit` values.
- **Admission Controller (Webhook)**: when a Pod is created, automatically adjusts `request`/`limit` according to the recommendation before the Pod starts.
- **Updater**: detects Pods running with outdated requests and evicts them so they are recreated with the new values.

### Update modes

| Mode | Behavior |
|---|---|
| `Off` | Only computes recommendations, never applies them automatically |
| `Initial` | Applies recommendations only when a Pod is first created, never evicts running Pods |
| `Recreate` | Evicts and recreates Pods when recommendations change significantly |
| `Auto` | Same as `Recreate` (default, may change in the future) |

### Configuration

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: "*"
      minAllowed:
        cpu: 100m
        memory: 50Mi
      maxAllowed:
        cpu: 2
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
```

### Limitations

- VPA does not work well with workloads that require a fixed replica count (stateful apps).
- `Recreate`/`Auto` mode causes downtime if there are not enough replicas (requires a `PodDisruptionBudget`).
- Do not use VPA together with HPA on the same CPU/Memory metric.

---

## Cluster Autoscaler

**Official docs:** https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler

> Cluster Autoscaler depends on the cloud provider (GKE, EKS, AKS...).

### How it works

- **Scale up**: when Pods are stuck in `Pending` due to insufficient resources → Cluster Autoscaler adds a new Node.
- **Scale down**: when a Node has low utilization and all its Pods can be rescheduled elsewhere → removes the Node after `scale-down-delay` (default 10 minutes).

### Conditions for scaling down a Node

- Utilization below the threshold (default 50%).
- All Pods on the Node can be rescheduled on other Nodes.
- No Pod on the Node is blocked by a `PodDisruptionBudget`.
- No Pod has the annotation `cluster-autoscaler.kubernetes.io/safe-to-evict: "false"`.
- The Node does not host Pods with `local storage` (emptyDir, hostPath) unless explicitly configured to allow it.

### Node Groups / Node Pools

Cluster Autoscaler manages **Node Groups** (AWS: Auto Scaling Group, GCP: Node Pool). Each group has a `min` and `max` node count.

```yaml
# Annotation on a Node to prevent CA from evicting it during scale down
cluster-autoscaler.kubernetes.io/safe-to-evict: "false"
```

---

## KEDA — Kubernetes Event-Driven Autoscaling

**Official docs:** https://keda.sh/docs/

> KEDA is an add-on that extends HPA with the ability to scale based on events from many external sources and supports **scale-to-zero**.

### How it works

- KEDA installs a `ScaledObject` CRD — underneath it automatically creates and manages the corresponding HPA resource.
- KEDA connects to **scalers** (Kafka, RabbitMQ, Redis, Prometheus, AWS SQS, Azure Service Bus...) to pull metrics and feed them into the HPA pipeline.
- Supports scaling to **0 replicas** when there are no events (native HPA cannot do this).

### Basic configuration

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: my-app-keda
spec:
  scaleTargetRef:
    name: my-app
  minReplicaCount: 0    # scale-to-zero
  maxReplicaCount: 20
  cooldownPeriod: 60    # seconds to wait before scaling to 0
  triggers:
  - type: kafka
    metadata:
      bootstrapServers: kafka:9092
      topic: my-topic
      consumerGroup: my-group
      lagThreshold: "10"   # scale up when lag per replica exceeds 10 messages
```

---

## Comparison and combining strategies

| | HPA | VPA | Cluster Autoscaler | KEDA |
|---|---|---|---|---|
| Scale target | Pod count | Pod resources | Node count | Pod count |
| Metric source | CPU, Memory, Custom | CPU, Memory | Pod Pending | Event-driven |
| Scale-to-zero | No | No | No | Yes |
| Can combine with | VPA (Off/Initial mode) | HPA (different metric) | Always | Replaces HPA |

**Common strategies:**
- HPA + Cluster Autoscaler: the most common combo — HPA scales Pods, CA scales Nodes to accommodate them.
- VPA (mode `Off`) + HPA: use VPA only to *suggest* request sizes, let HPA handle scale-out.
- KEDA + Cluster Autoscaler: for event-driven workloads (consumers, batch jobs).
