# Introduction

- [Kubernetes](https://kubernetes.io/docs/concepts/overview/), also known as `K8s`, is an open-source system for automating deployment, scaling, and management of containerized applications. `Kubernetes` is a portable, extensible, open source platform for managing containerized workloads and services, that facilitates both declarative configuration and automation. It has a large, rapidly growing ecosystem. `Kubernetes` services, support, and tools are widely available.
- Planet Scale: Designed on the same principles that allow Google to run billions of containers a week, `Kubernetes` can scale without increasing your operations team.
- Never Outgrow: Whether testing locally or running a global enterprise, `Kubernetes` flexibility grows with you to deliver your applications consistently and easily no matter how complex your need is.
- Run `K8s` Anywhere: `Kubernetes` is open source giving you the freedom to take advantage of on-premises, hybrid, or public cloud infrastructure, letting you effortlessly move workloads to where it matters to you.

# Components

![k8s components](../../images/components-of-kubernetes.svg)

## Core components

- A `Kubernetes cluster` consists of a *control plane* and *one or more worker nodes*. Here's a brief overview of the main components:

### Control Plane Components

- Manage the overall state of the cluster:
  - [**kube-apiserver**](https://kubernetes.io/docs/concepts/architecture/#kube-apiserver): The core component server that exposes the Kubernetes HTTP API.
  - [**etcd**](https://kubernetes.io/docs/concepts/architecture/#etcd): Consistent and highly-available key value store for all API server data.
  - [**kube-scheduler**](https://kubernetes.io/docs/concepts/architecture/#kube-scheduler): Looks for Pods not yet bound to a node, and assigns each Pod to a suitable node.
  - [**kube-controller-manager**](https://kubernetes.io/docs/concepts/architecture/#kube-controller-manager): Runs controllers to implement Kubernetes API behavior.
  - [**cloud-controller-manager**](https://kubernetes.io/docs/concepts/architecture/#cloud-controller-manager): Integrates with underlying cloud provider(s).

### Node Components

- Run on every node, maintaining running pods and providing the Kubernetes runtime environment:
  - [**kubelet**](https://kubernetes.io/docs/concepts/architecture/#kubelet): Ensures that Pods are running, including their containers.
  - [**kube-proxy**](https://kubernetes.io/docs/concepts/architecture/#kube-proxy): Maintains network rules on nodes to implement Services.
  - [**Container runtime**](https://kubernetes.io/docs/concepts/architecture/#container-runtime): Software responsible for running containers.

# Objects in Kubernetes

- `Kubernetes` objects are *persistent entities* in the `Kubernetes` system. `Kubernetes` uses these entities to represent the state of your cluster. Learn about the `Kubernetes` object model and how to work with these objects.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 2 # tells deployment to run 2 pods matching the template
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
```

- Required fields in a `Kubernetes` object manifest:
  - `apiVersion`: The versioned schema of the object representation.
  - `kind`: The type of `Kubernetes` object being created.
  - `metadata`: Data that helps uniquely identify the object, including a `name` string, `UID`, and optional `namespace`.
  - `spec`: The desired state of the object, as defined by the user.
- In `Kubernetes`, namespaces provide a mechanism for isolating groups of resources within a single cluster. Names of resources need to be unique within a namespace, but not across namespaces. Namespace-based scoping is applicable only for `namespaced` objects (e.g. `Deployments`, `Services`, etc.) and not for cluster-wide objects (e.g. `StorageClass`, `Nodes`, `PersistentVolumes`, etc.)
