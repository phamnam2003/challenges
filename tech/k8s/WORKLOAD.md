# WORKLOADS

- A `workload` is an *application* running on `Kubernetes`. Whether your workload is a single component or several that work together, on `Kubernetes` you run it inside a set of [*pods*](https://kubernetes.io/docs/concepts/workloads/pods/). In `Kubernetes`, a Pod represents a set of running *containers* on your cluster.
- `Kubernetes pods` have a [defined lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/).
- `Kubernetes` provides several built-in workload resources
  - [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) and [ReplicaSet](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/): Deployment is a good fit for managing a stateless application workload on your cluster, where any Pod in the Deployment is interchangeable and can be replaced if needed.
  - [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/): lets you run one or more related Pods that do track state somehow, it often works with [PersistentVolumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) to provide stable storage for each Pod.
  - [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/): defines Pods that provide facilities that are local to nodes. Every time you add a node to your cluster that matches the specification in a `DaemonSet`, the control plane schedules a Pod for that `DaemonSet` onto the new node
  - [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) and [CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/): provide different ways to define tasks that run to completion and then stop.

## Pods

### What is a Pos?

- `Pod`s are the *smallest deployable units* of computing that you can *create and manage* in `Kubernetes`.
- A `Pod` (as in a pod of whales or pea pod) is a *group of one or more containers*, with *shared storage and network resources*, and a specification for how to run the containers.
- A `Pod` is similar to a set of containers with *shared namespaces and shared filesystem volumes*. `Pods` in a `Kubernetes cluster` are used in two main ways:
  - **Pods that run a single container**: the most common use case; in this case, you can think a `Pod` as a wrapper around a single container.
  - **Pods that run multiple containers that need to work together**: A Pod can encapsulate an application composed of multiple co-located containers that are tightly coupled and need to share resources. These co-located containers form a single cohesive unit.

### Using Pods

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
  - name: nginx
    image: nginx:1.14.2
    ports:
    - containerPort: 80
```

- To create the Pod shown above, run the following command:

```bash
kubectl apply -f <yaml_config_file>
```

### Resource sharing and communication

- Storage in Pods: A Pod can specify a set of shared storage volumes. All containers in the Pod can access the shared volumes, allowing those containers to share data. Volumes also allow persistent data in a Pod to survive in case one of the containers within needs to be restarted
- See [Storage](https://kubernetes.io/docs/concepts/storage/) for more information on how `Kubernetes` implements shared storage and makes it available to Pods.

### Pod networking

- Each `Pod` is assigned a *unique IP address* for each address family. Every *container* in a Pod shares the network namespace, including the IP address and network ports. Inside a `Pod` (and only then), the containers that belong to the `Pod` can communicate with one another using `localhost`
