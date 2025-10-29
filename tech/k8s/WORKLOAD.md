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

### Pod lifecycle

- Whilst a `Pod` is running, the **`kubelet`** is able to *restart containers* to *handle some kind of faults*. Within a `Pod`, `Kubernetes` tracks different container states and determines what action to take to make the Pod healthy again.
- In the `Kubernetes API`, `Pods` have both a specification and an actual status.
- Pod phases:
  - Pending: The Pod has been accepted by the `Kubernetes cluster`, but one or more of the containers has not been set up and made ready to run
  - Running: The Pod has been bound to a node, and all of the containers have been created. At least one container is still running, or is in the process of starting or restarting.
  - Succeeded: All containers in the Pod have terminated in success, and will not be restarted.
  - Failed: All containers in the Pod have terminated, and at least one container has terminated in failure
  - Unknown: For some reason the state of the Pod could not be obtained. This phase typically occurs due to an error in communicating with the node where the Pod should be running.
- Container States:
  - Waiting: If a container is not in either the `Running` or `Terminated` state, it is `Waiting`
  - Running: The Running status indicates that a container is executing without issues.
  - Terminated: A container in the Terminated state began execution and then either ran to completion or failed for some reason.
- How Pods handle problems with containers:
  - **Initial crash**: `Kubernetes` attempts an immediate restart based on the Pod `restartPolicy`.
  - **Repeated crashes**: After the initial crash `Kubernetes` applies an *exponential backoff* delay for subsequent restarts, described in `restartPolicy`. This prevents rapid, repeated restart attempts from overloading the system.
  - **CrashLoopBackOff state**: This indicates that the backoff delay mechanism is currently in effect for a given container that is in a crash loop, failing and restarting repeatedly.
  - **Backoff reset**: If a container runs successfully for a certain duration (e.g., 10 minutes), `Kubernetes` resets the backoff delay, treating any new crash as the first one.

```yaml
piVersion: v1
kind: Pod
metadata:
  name: on-failure-pod
spec:
  restartPolicy: OnFailure
  containers:
  - name: try-once-container    # This container will run only once because the restartPolicy is Never.
    image: docker.io/library/busybox:1.28
    command: ['sh', '-c', 'echo "Only running once" && sleep 10 && exit 1']
    restartPolicy: Never
  - name: on-failure-container  # This container will be restarted on failure.
    image: docker.io/library/busybox:1.28
    command: ['sh', '-c', 'echo "Keep restarting" && sleep 1800 && exit 1']
```

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: fail-pod-if-init-fails
spec:
  restartPolicy: Always
  initContainers:
  - name: init-once      # This init container will only try once. If it fails, the pod will fail.
    image: docker.io/library/busybox:1.28
    command: ['sh', '-c', 'echo "Failing initialization" && sleep 10 && exit 1']
    restartPolicy: Never
```

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: restart-on-exit-codes
spec:
  restartPolicy: Never
  containers:
  - name: restart-on-exit-codes
    image: docker.io/library/busybox:1.28
    command: ['sh', '-c', 'sleep 60 && exit 0']
    restartPolicy: Never     # Container restart policy must be specified if rules are specified
    restartPolicyRules:      # Only restart the container if it exits with code 42
    - action: Restart
      exitCodes:
        operator: In
        values: [42]
```
