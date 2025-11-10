# Ingress in Kubernetes

Make your `HTTP` (or `HTTPS`) network service available using a protocol-aware configuration mechanism, that understands web concepts like URIs, hostnames, paths, and more. The Ingress concept *lets you map traffic* to different backends based on rules you define via the Kubernetes API.

## Prerequisites

- Must have [Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) to satisfy the Ingress.
- You may need tp deploy an Ingress Controller such as [ingress-nginx](https://kubernetes.github.io/ingress-nginx/deploy/).

## Type of Ingress

### Ingress backend by a single Service

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress
spec:
  defaultBackend:
    service:
      name: test
      port:
        number: 80
```

### Fanout

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: simple-fanout-example
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          service:
            name: service1
            port:
              number: 4200
      - path: /bar
        pathType: Prefix
        backend:
          service:
            name: service2
            port:
              number: 8080
```

![Ingress fanout](../../../images/ingressFanOut.svg)

## Installation

- Recommend to use [Helm](https://helm.sh/) to install the Ingress Controller. Install Helm in [here](https://helm.sh/docs/intro/install)

### Nginx Ingress Controller

- Add repo Ingress Nginx by Helm

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx/
helm repo update
helm search repo ingress-nginx
```

- Pull helm chart to use `values.yaml` to customize the installation, we often change some field in that file before apply into k8s cluster:

```bash
helm pull ingress-nginx/ingress-nginx --untar
```

- If you use Cloud to implentation k8s cluster you can use `LoadBalancer` service type, otherwise you can use `NodePort` or `ClusterIP` service type. And change value `http` and `https` port if you want (http and https of a service field).
- Create new namespace to apply Ingress Controller. After that install Ingress Nginx by Helm:

```bash
helm -n <ingress_namespace> install <release_name> -f ingress-nginx/values.yaml ingress-nginx/ingress-nginx
```

- You need one server made Load Balancer, and install `nginx` or `haproxy` to forward traffic from port `80` and `443` to the Ingress Controller service.

```nginx
upstream k8s_servers {
  server 192.168.79.11:30080;
  server 192.168.79.12:30080;
}
server {
  listen 80;
  location / {
    proxy_pass http://k8s_servers;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 90;
  }
}
```
