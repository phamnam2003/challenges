# Services, Load Balancing, and Networking

## The Kubernetes network model

- *Each* [pod](https://kubernetes.io/docs/concepts/workloads/pods/) in a cluster gets its *own unique cluster-wide IP address*: A pod has its *own private network namespace* which is *shared* by all of the *containers* within the pod. Processes running in *different containers* in the *same pod* **can communicate with each other** over `localhost`.
- The *pod network* (also called a *cluster network*) handles communication between pods. It ensures that (barring intentional network segmentation):
  - All pods can communicate with all other pods, whether they are on the same node or on different nodes. Pods can communicate with each other directly, without the use of proxies or address translation (NAT).
  - Agents on a node (such as system daemons, or kubelet) can communicate with all pods on that node.
- The [Service](https://kubernetes.io/docs/concepts/services-networking/service/) API lets you provide a *stable (long lived) IP address* or *hostname* for a service implemented by one or more backend pods, where the individual pods making up the service can change over time.
- The [Gateway](https://kubernetes.io/docs/concepts/services-networking/gateway/) API (or its predecessor, [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)) allows you to make Services accessible to clients that are *outside the cluster*.
- [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) is a built-in `Kubernetes API` that allows you to *control traffic between pods*, or *between pods* and the *outside world*.

## Services

- *Expose* an application running in your cluster behind a single outward-facing endpoint, even when the workload is split across multiple backends.
- The *Service API*, part of `Kubernetes`, is an abstraction to help you *expose groups of Pods* over a network. *Each Service object* defines a logical set of endpoints (usually these endpoints are Pods) along with a policy about how to make those pods accessible.

### Defining a Service

- A `Service` is *an object* (the same way that a `Pod` or a `ConfigMap` is an object). You can create, view or modify `Service` definitions using the `Kubernetes API`. Usually you use a tool such as `kubectl` to make those API calls for you.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app.kubernetes.io/name: MyApp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
```

### Port definitions

- `Port definitions` in `Pods` have *names*, and you can reference these names in the `targetPort` attribute of a `Service`. For example, we can bind the `targetPort` of the Service to the Pod port in the following way:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    app.kubernetes.io/name: proxy
spec:
  containers:
  - name: nginx
    image: nginx:stable
    ports:
      - containerPort: 80
        name: http-web-svc
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app.kubernetes.io/name: proxy
  ports:
  - name: name-of-service-port
    protocol: TCP
    port: 80
    targetPort: http-web-svc
```

- This works even if there is a mixture of `Pods` in the `Service` using a single configured name, with the same network protocol available via different port numbers. This offers a lot of flexibility for *deploying* and *evolving* your `Services`.
- Multi-port Services: For some Services, you need to expose more than one port. Kubernetes lets you configure multiple port definitions on a Service object. When using multiple ports for a Service, you must give all of your ports names so that these are unambiguous

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app.kubernetes.io/name: MyApp
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: 9376
    - name: https
      protocol: TCP
      port: 443
      targetPort: 9377
```

### Service types

- The available type values and their behaviors are:
  - [**ClusterIP**](https://kubernetes.io/docs/concepts/services-networking/service/#type-clusterip): Exposes the Service on a cluster-internal IP. Choosing this value makes the Service only reachable from within the cluster. This is the default that is used if you don't explicitly specify a type for a Service. You can expose the Service to the public internet using an [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) or a [Gateway](https://gateway-api.sigs.k8s.io/).
  - [**NodePort**](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport): Exposes the Service on each Node's IP at a static port (the `NodePort`). To make the node port available, `Kubernetes` sets up a cluster IP address, the same as if you had requested a Service of type: `ClusterIP`. the Kubernetes control plane allocates a port from a range specified by `--service-node-port-range` flag (default: `30000-32767`)
  - [**LoadBalancer**](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer): Exposes the Service externally using an external load balancer. Kubernetes does not directly offer a load balancing component; you must provide one, or you can integrate your Kubernetes cluster with a cloud provider.
  - [**ExternalName**](https://kubernetes.io/docs/concepts/services-networking/service/#externalname): Maps the Service to the contents of the externalName field (for example, to the *hostname* `api.foo.bar.example`). The mapping configures your cluster's DNS server to return a `CNAME` record with that external `hostname` value. No proxying of any kind is set up.

## Ingress

- Make your HTTP (or HTTPS) network service available using *a protocol-aware configuration mechanism*, that understands web concepts like `URIs`, `hostnames`, `paths`, and **more**. The `Ingress` concept lets you *map traffic to different backends* based on rules you define via the `Kubernetes API`.
- [Ingress](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#ingress-v1-networking-k8s-io) exposes HTTP and HTTPS routes from outside the cluster to services within the cluster. Traffic routing is controlled by rules defined on the Ingress resource.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: minimal-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx-example
  rules:
  - http:
      paths:
      - path: /testpath
        pathType: Prefix
        backend:
          service:
            name: test
            port:
              number: 80
```
