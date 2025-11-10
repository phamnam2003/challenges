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

- Recommend to use [Helm](https://helm.sh/) to install the Ingress Controller.

### Nginx Ingress Controller
