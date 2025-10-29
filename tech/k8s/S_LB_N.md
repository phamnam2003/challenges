# Services, Load Balancing, and Networking

## The Kubernetes network model

- *Each* [pod](https://kubernetes.io/docs/concepts/workloads/pods/) in a cluster gets its *own unique cluster-wide IP address*: A pod has its *own private network namespace* which is *shared* by all of the *containers* within the pod. Processes running in *different containers* in the *same pod* **can communicate with each other** over `localhost`.
- The *pod network* (also called a *cluster network*) handles communication between pods. It ensures that (barring intentional network segmentation):
  - All pods can communicate with all other pods, whether they are on the same node or on different nodes. Pods can communicate with each other directly, without the use of proxies or address translation (NAT).
  - Agents on a node (such as system daemons, or kubelet) can communicate with all pods on that node.
- The [Service](https://kubernetes.io/docs/concepts/services-networking/service/) API lets you provide a *stable (long lived) IP address* or *hostname* for a service implemented by one or more backend pods, where the individual pods making up the service can change over time.
- The [Gateway](https://kubernetes.io/docs/concepts/services-networking/gateway/) API (or its predecessor, [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)) allows you to make Services accessible to clients that are *outside the cluster*.
- [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) is a built-in `Kubernetes API` that allows you to *control traffic between pods*, or *between pods* and the *outside world*.
