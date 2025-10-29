# WORKLOADS

- A `workload` is an *application* running on `Kubernetes`. Whether your workload is a single component or several that work together, on `Kubernetes` you run it inside a set of [*pods*](https://kubernetes.io/docs/concepts/workloads/pods/). In `Kubernetes`, a Pod represents a set of running *containers* on your cluster.
- `Kubernetes pods` have a [defined lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/).
- `Kubernetes` provides several built-in workload resources
  - [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) and [ReplicaSet](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/): Deployment is a good fit for managing a stateless application workload on your cluster, where any Pod in the Deployment is interchangeable and can be replaced if needed.
  - [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/): lets you run one or more related Pods that do track state somehow, it often works with [PersistentVolumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) to provide stable storage for each Pod.
  - [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/): defines Pods that provide facilities that are local to nodes. Every time you add a node to your cluster that matches the specification in a `DaemonSet`, the control plane schedules a Pod for that `DaemonSet` onto the new node
  - [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/) and [CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/): provide different ways to define tasks that run to completion and then stop.
