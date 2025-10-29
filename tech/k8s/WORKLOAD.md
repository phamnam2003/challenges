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

## Workload Management

- `Kubernetes` provides several *built-in APIs* for declarative management of your workloads and the components of those workloads.

### Deployments

- A `Deployment` manages a set of Pods to run an application workload, usually one that doesn't maintain state.
- A `Deployment` provides declarative updates for *Pods* and *ReplicaSets*.
- You describe a desired state in a Deployment, and the Deployment Controller changes the actual state to the desired state at a controlled rate. You can define Deployments to create new ReplicaSets, or to remove existing Deployments and adopt all their resources with new Deployments.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
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

```bash
# Apply the Deployment configuration
kubectl apply -f <yaml_config_file>
# Show deployments status 
kubectl get deployments
```

### ReplicaSet

- A `ReplicaSet's` purpose is to maintain a stable set of replica `Pods` running at any given time. Usually, you define a `Deployment` and let that Deployment manage `ReplicaSets` automatically.
- A `ReplicaSet` is defined with *fields*, including a *selector* that specifies how to identify `Pods` it can acquire, a *number of replicas indicating how many Pods* it should be maintaining, and a pod template specifying the data of new `Pods` it should create to meet the number of replicas criteria.
- A `ReplicaSet` ensures that a specified number of pod replicas are running at any given time

```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: frontend
  labels:
    app: guestbook
    tier: frontend
spec:
  # modify replicas according to your case
  replicas: 3
  selector:
    matchLabels:
      tier: frontend
  template:
    metadata:
      labels:
        tier: frontend
    spec:
      containers:
      - name: php-redis
        image: us-docker.pkg.dev/google-samples/containers/gke/gb-frontend:v5
```

- You can then get the current ReplicaSets deployed and describe that:

```bash
kubectl get rs
# NAME       DESIRED   CURRENT   READY   AGE
# frontend   3         3         3       6s
kubectl describe rs/frontend
# Name:         frontend
# Namespace:    default
# Selector:     tier=frontend
# Labels:       app=guestbook
#               tier=frontend
# Annotations:  <none>
# Replicas:     3 current / 3 desired
# Pods Status:  3 Running / 0 Waiting / 0 Succeeded / 0 Failed
# Pod Template:
#   Labels:  tier=frontend
#   Containers:
#    php-redis:
#     Image:        us-docker.pkg.dev/google-samples/containers/gke/gb-frontend:v5
#     Port:         <none>
#     Host Port:    <none>
#     Environment:  <none>
#     Mounts:       <none>
#   Volumes:        <none>
# Events:
#   Type    Reason            Age   From                   Message
#   ----    ------            ----  ----                   -------
#   Normal  SuccessfulCreate  13s   replicaset-controller  Created pod: frontend-gbgfx
#   Normal  SuccessfulCreate  13s   replicaset-controller  Created pod: frontend-rwz57
#   Normal  SuccessfulCreate  13s   replicaset-controller  Created pod: frontend-wkl7w
```

- And lastly you can check for the Pods brought up:

```bash
kubectl get pods
# NAME             READY   STATUS    RESTARTS   AGE
# frontend-gbgfx   1/1     Running   0          10m
# frontend-rwz57   1/1     Running   0          10m
# frontend-wkl7w   1/1     Running   0          10m
```

### StatefulSets

- A `StatefulSet` runs *a group of Pods*, and maintains a sticky identity for each of those Pods. This is useful for managing applications that need persistent storage or a stable, unique network identity.
- `StatefulSet` is the workload API object used to manage stateful applications.
- StatefulSets are valuable for applications that require one or more of the following
  - Stable, unique network identifiers.
  - Stable, persistent storage.
  - Ordered, graceful deployment and scaling.
  - Ordered, automated rolling updates.
- Components:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 80
    name: web
  clusterIP: None
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  selector:
    matchLabels:
      app: nginx # has to match .spec.template.metadata.labels
  serviceName: "nginx"
  replicas: 3 # by default is 1
  minReadySeconds: 10 # by default is 0
  template:
    metadata:
      labels:
        app: nginx # has to match .spec.selector.matchLabels
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: nginx
        image: registry.k8s.io/nginx-slim:0.24
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "my-storage-class"
      resources:
        requests:
          storage: 1Gi
```

### DaemonSet

- A `DaemonSet` defines Pods that provide *node-local facilities*. These might be fundamental to the operation of your cluster, such as a networking helper tool, or be part of an add-on.
- A `DaemonSet` ensures that all (or some) Nodes run a copy of a Pod. As nodes are added to the cluster, Pods are added to them. Deleting a `DaemonSet` will clean up the Pods it created. Some typical uses of a `DaemonSet` are:
  - running a cluster storage daemon on every node
  - running a logs collection daemon on every node
  - running a node monitoring daemon on every node

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-elasticsearch
  namespace: kube-system
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: fluentd-elasticsearch
  template:
    metadata:
      labels:
        name: fluentd-elasticsearch
    spec:
      tolerations:
      # these tolerations are to have the daemonset runnable on control plane nodes
      # remove them if your control plane nodes should not run pods
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      containers:
      - name: fluentd-elasticsearch
        image: quay.io/fluentd_elasticsearch/fluentd:v5.0.1
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 100m
            memory: 200Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
      # it may be desirable to set a high priority class to ensure that a DaemonSet Pod
      # preempts running Pods
      # priorityClassName: important
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
```

- `DaemonSets` are similar to Deployments in that they both create Pods, and those Pods have processes which are not expected to terminate (e.g. web servers, storage servers).
