# kube-prometheus-stack

**kube-prometheus-stack** is a Helm chart from `prometheus-community` that bundles the standard Kubernetes monitoring stack — Prometheus, Alertmanager, Grafana, kube-state-metrics, node-exporter, and **Prometheus Operator** — pre-configured to work together out of the box, with hundreds of dashboards and alert rules for every K8s component included.

## Bundled Components

| Component | Role |
|---|---|
| **Prometheus** | Collects and stores metrics as time-series data |
| **Alertmanager** | Receives alerts from Prometheus, routes and deduplicates before notifying |
| **Grafana** | Visualization dashboards |
| **kube-state-metrics** | Exposes K8s object state (Deployment, Pod, Node...) as metrics |
| **prometheus-node-exporter** | Exposes hardware and OS metrics from each Node |
| **Prometheus Operator** | CRD controller — lets you configure Prometheus via K8s objects instead of raw ConfigMaps |

Prometheus Operator is the most important component — it introduces CRDs so you can declare scraping and alerting as K8s objects, without editing raw ConfigMaps or restarting Prometheus.

---

## Key CRDs

### ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: my-app
  namespace: monitoring
  labels:
    release: kube-prometheus-stack    # must match the Prometheus CR's selector
spec:
  selector:
    matchLabels:
      app: my-app
  namespaceSelector:
    matchNames:
      - production
  endpoints:
    - port: metrics
      interval: 30s
      path: /metrics
```

### PrometheusRule

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: my-app-alerts
  namespace: monitoring
  labels:
    release: kube-prometheus-stack
spec:
  groups:
    - name: my-app
      rules:
        - alert: HighErrorRate
          expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: "Error rate exceeded 5% for 2 minutes"
```

**PodMonitor** — same as ServiceMonitor but scrapes Pods directly, used when a Pod has no backing Service.

---

## Installation

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --values values.yaml
```

---

## Configuration via values.yaml

```yaml
prometheus:
  prometheusSpec:
    retention: 30d
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: longhorn
          resources:
            requests:
              storage: 50Gi
    serviceMonitorSelector:
      matchLabels:
        release: kube-prometheus-stack

grafana:
  adminPassword: "your-password"
  ingress:
    enabled: true
    hosts:
      - grafana.example.com
  persistence:
    enabled: true
    size: 5Gi

alertmanager:
  config:
    route:
      group_by: ["alertname", "namespace"]
      receiver: "slack"
    receivers:
      - name: "slack"
        slack_configs:
          - api_url: "https://hooks.slack.com/..."
            channel: "#alerts"
```

### serviceMonitorSelector

The Prometheus CR only picks up `ServiceMonitor` resources whose labels match `serviceMonitorSelector`. The chart default is `release: kube-prometheus-stack` — any ServiceMonitor you create must carry this label, otherwise Prometheus silently ignores it with no error.

To scrape all ServiceMonitors regardless of labels:

```yaml
prometheus:
  prometheusSpec:
    serviceMonitorSelectorNilUsesHelmValues: false
    serviceMonitorSelector: {}
    serviceMonitorNamespaceSelector: {}
```

---

## Control Plane & kube-proxy Targets: Why Are They Down?

kube-prometheus-stack creates ServiceMonitors for system components automatically, but Prometheus gets **connection refused** when scraping — because these components bind their metrics port to `127.0.0.1` by default, not `0.0.0.0`.

| Component | Port | Config location |
|---|---|---|
| kube-proxy | 10249 | ConfigMap `kube-proxy` in `kube-system` |
| kube-scheduler | 10259 | Static pod `/etc/kubernetes/manifests/kube-scheduler.yaml` |
| kube-controller-manager | 10257 | Static pod `/etc/kubernetes/manifests/kube-controller-manager.yaml` |
| etcd | 2381 | Static pod `/etc/kubernetes/manifests/etcd.yaml` |

### Fix kube-proxy

```bash
kubectl edit configmap kube-proxy -n kube-system
# change: metricsBindAddress: "0.0.0.0:10249"

kubectl rollout restart daemonset kube-proxy -n kube-system
```

### Fix kube-scheduler and kube-controller-manager

SSH into the control plane node and edit the static pod manifests directly — kubelet will restart the Pod automatically after detecting the change:

```yaml
# /etc/kubernetes/manifests/kube-scheduler.yaml
# /etc/kubernetes/manifests/kube-controller-manager.yaml
spec:
  containers:
  - command:
    - kube-scheduler          # or kube-controller-manager
    - --bind-address=0.0.0.0
```

### Managed K8s (GKE, EKS, AKS)

The control plane is not accessible — there is no fix. Disable these targets in values:

```yaml
kubeScheduler:
  enabled: false
kubeControllerManager:
  enabled: false
kubeEtcd:
  enabled: false
kubeProxy:
  enabled: false    # GKE Dataplane V2 does not use kube-proxy
```

---

## Upgrade

Helm does **not upgrade CRDs** when running `helm upgrade`. Manually upgrade CRDs before each major version bump — check the chart changelog to know which CRDs changed.

---

## Common Pitfalls

**ServiceMonitor not being scraped:** missing label `release: kube-prometheus-stack`. Check the Targets tab in Prometheus UI — if the target does not appear, the Operator has not picked it up.

**Control plane targets down:** default bind address is `127.0.0.1`. Fix each component as described above, or disable entirely on managed K8s.

**Grafana loses dashboards after restart:** enable `grafana.persistence.enabled: true` or manage dashboards via ConfigMap with the annotation `grafana_dashboard: "1"`.

**Prometheus OOM:** large clusters with long retention need a high `resources.limits.memory`. Consider Thanos or VictoriaMetrics for long-term storage.

**Alert noise after fresh install:** the chart ships hundreds of alert rules, many of which fire immediately on managed K8s due to missing components. Audit and disable irrelevant rules before enabling real notifications.

---

## References

- [kube-prometheus-stack - Artifact Hub](https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack)
- [Prometheus Operator - Getting Started](https://prometheus-operator.dev/docs/getting-started/introduction/)
- [Prometheus Operator - CRD References](https://prometheus-operator.dev/docs/api-reference/api/)
- [kube-prometheus-stack values.yaml](https://github.com/prometheus-community/helm-charts/blob/main/charts/kube-prometheus-stack/values.yaml)
