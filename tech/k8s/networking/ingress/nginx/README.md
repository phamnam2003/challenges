# NGINX Ingress Controller

## Prerequisites

- Helm 3.x installed
- kubectl configured against your cluster
- Bare-metal cluster (NodePort exposure)

## Installation

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=NodePort \
  --set controller.service.nodePorts.http=30080 \ 
  --set controller.service.nodePorts.https=30443
```

`NodePort` is used instead of `LoadBalancer` because bare-metal clusters have no cloud load balancer provider.

## Verify

```bash
kubectl get pods -n ingress-nginx
kubectl get svc -n ingress-nginx
```

The controller pod should reach `Running` and the service will show the assigned NodePort (e.g. `80:3xxxx/TCP`, `443:3xxxx/TCP`).

## Apply workloads

Apply in this order:

```bash
kubectl apply -f ../namespace.yaml
kubectl apply -f ../gate-controller.yaml
kubectl apply -f ../gate-unit.yaml
kubectl apply -f ingress.yaml
```

Verify the ingress was accepted:

```bash
kubectl get ingress -n go-gate-ns
```

The `ADDRESS` column showing the ingress controller's cluster IP means the ingress is active.

## Accessing services

Two approaches depending on the environment:

### Bare metal (node IP is reachable)

Use when the cluster node is on the same network as the client machine — node IP is directly pingable.

```bash
# Get node IP
kubectl get nodes -o wide

# Get NodePort
kubectl get svc ingress-nginx-controller -n ingress-nginx
```

Add to hosts file:
```
<NODE_IP>  k8s.gate.local.io
```

Access:
```
http://k8s.gate.local.io:<NODE_PORT>/controller/health-check
http://k8s.gate.local.io:<NODE_PORT>/unit/health-check
```

---

### Local / minikube on WSL2 (node IP not reachable from Windows)

Use when minikube runs inside Docker/WSL2 — the node IP (`192.168.49.x`) is isolated and not pingable from Windows.

**Dedicated terminal — keep running:**
```bash
kubectl port-forward svc/ingress-nginx-controller 80:80 -n ingress-nginx
```

Add to `C:\Windows\System32\drivers\etc\hosts` (Notepad as Admin or PowerShell as Admin):
```
127.0.0.1 k8s.gate.local.io
```

PowerShell as Admin:
```powershell
Add-Content -Path "C:\Windows\System32\drivers\etc\hosts" -Value "127.0.0.1 k8s.gate.local.io"
```

Access (standard port 80, no NodePort needed):
```
http://k8s.gate.local.io/controller/health-check
http://k8s.gate.local.io/unit/health-check
```
