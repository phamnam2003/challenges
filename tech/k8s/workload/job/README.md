# Job in Kubernetes

## What is a Job?

**Job** is a workload controller for running a task **to completion** — unlike Deployment or DaemonSet which keep Pods running indefinitely. A Job creates one or more Pods, tracks them until enough successful completions are recorded, then stops.

When a Job's Pod exits with code 0 — the Job records one successful completion. When the total reaches `completions` — the Job is marked `Complete`. If Pods fail more than `backoffLimit` times — the Job is marked `Failed`.

---

## Problem It Solves

Deployment is not suited for tasks with a defined endpoint — if a Pod exits, Deployment treats it as a failure and immediately restarts it. Jobs handle:

- **Database migrations** — run once, stop on success
- **Batch processing** — queue processing, video rendering, report exports
- **One-time setup** — seed data, create admin accounts, warm caches
- **Parallel workloads** — split a large task into many smaller tasks running concurrently

---

## Manifest

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration
  namespace: default
spec:
  completions: 1              # number of successful completions required (default: 1)
  parallelism: 1              # number of Pods running simultaneously (default: 1)
  backoffLimit: 3             # max retries before Job is marked Failed (default: 6)
  activeDeadlineSeconds: 300  # Job is terminated if it runs longer than this
  ttlSecondsAfterFinished: 3600  # auto-delete Job this many seconds after it finishes

  template:
    spec:
      restartPolicy: Never    # Never | OnFailure — Always is not allowed
      containers:
        - name: migration
          image: my-app:latest
          command: ["python", "manage.py", "migrate"]
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: url
```

---

## Key Fields

### `completions` and `parallelism`

These two fields together control how the Job runs:

| completions | parallelism | Behavior |
|-------------|-------------|----------|
| 1 | 1 | Run 1 Pod, stop when done (default) |
| 5 | 1 | Run 5 Pods sequentially, one at a time |
| 5 | 3 | Run up to 3 Pods in parallel until 5 successful completions |
| unset | 3 | Run 3 Pods in parallel indefinitely until the Job is terminated (work queue pattern) |

### `backoffLimit`

The number of times Pods are allowed to fail before the Job is marked `Failed`. Between retries, Kubernetes waits progressively longer using exponential backoff (10s, 20s, 40s...). Default is `6`.

With `restartPolicy: OnFailure`, retries restart the container inside the same Pod. With `restartPolicy: Never`, each retry creates a new Pod.

### `activeDeadlineSeconds`

The maximum time a Job is allowed to run from when it starts. Once exceeded, all running Pods are terminated and the Job is marked `Failed` with reason `DeadlineExceeded` — regardless of remaining `backoffLimit`.

### `ttlSecondsAfterFinished`

The Job and its Pods are automatically deleted this many seconds after the Job finishes (both `Complete` and `Failed`). Without this, finished Jobs accumulate indefinitely until manually deleted.

### `restartPolicy`

Jobs only accept `Never` or `OnFailure` — `Always` is reserved for Deployment and will be rejected.

| Value | Behavior on Pod failure |
|-------|------------------------|
| `Never` | Pod is marked Failed; Job creates a new Pod to retry |
| `OnFailure` | Container inside the Pod is restarted in place |

`Never` is better for most cases — each retry gets its own Pod and its own logs, making debugging easier. `OnFailure` is useful when the filesystem needs to be preserved between retries.

---

## Completion Mode

### NonIndexed (default)

Every successful Pod completion counts equally — order doesn't matter. The Job only tracks the total number of completions.

### Indexed

Each Pod receives a unique index from `0` to `completions - 1`, injected via the environment variable `JOB_COMPLETION_INDEX`. Useful for splitting work by index — Pod 0 processes the first batch, Pod 1 the next, and so on.

```yaml
spec:
  completions: 4
  parallelism: 2
  completionMode: Indexed
  template:
    spec:
      containers:
        - name: worker
          image: my-worker
          env:
            - name: SHARD_INDEX
              valueFrom:
                fieldRef:
                  fieldPath: metadata.annotations['batch.kubernetes.io/job-completion-index']
```

---

## Job vs CronJob

A Job runs once. **CronJob** is a controller that creates Jobs on a recurring schedule (cron syntax). The relationship: CronJob → creates Job → creates Pod.

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: daily-report
spec:
  schedule: "0 2 * * *"        # runs at 2:00 AM every day
  concurrencyPolicy: Forbid     # Allow | Forbid | Replace
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:                       # this is the Job spec, not the Pod spec
      backoffLimit: 2
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: reporter
              image: my-reporter:latest
```

`concurrencyPolicy` controls what happens when the previous Job is still running when the next scheduled run is due:

| Value | Behavior |
|-------|---------|
| `Allow` | Both Jobs run in parallel |
| `Forbid` | Skip the current run if the previous Job is still running |
| `Replace` | Delete the old Job and create a new one |

---

## Monitoring Jobs

```bash
kubectl get jobs
kubectl describe job db-migration

# list Pods created by the Job
kubectl get pods -l job-name=db-migration

# stream logs from Job Pods
kubectl logs -l job-name=db-migration --tail=100
```

---

## Common Pitfalls

**`restartPolicy: Always` is rejected:** Jobs do not accept this value — the manifest will fail validation. Use `Never` or `OnFailure`.

**Failed Pods accumulate with `restartPolicy: Never`:** each retry creates a new Pod; previous failed Pods remain in `Failed` state until the Job is deleted. With a high `backoffLimit` and many Job runs, stale Pods pile up quickly. Set `ttlSecondsAfterFinished` to clean up automatically.

**No `activeDeadlineSeconds`:** a Job stuck due to a deadlock or connection timeout will run forever. Always set a realistic time limit.

**CronJob misses runs during cluster downtime:** if the cluster is down for more than one cycle, CronJob may skip some scheduled runs. For critical tasks, implement missed-run detection and handling logic.

**Large `completions` without setting `parallelism`:** the default is `parallelism: 1` — 100 completions run sequentially and can be very slow. Set `parallelism` appropriate to available cluster resources.

**No `resources.requests`:** the scheduler cannot calculate capacity accurately → cluster gets overcommitted → OOM kills or mid-run evictions interrupt the task. Always set resource requests on Job Pods.

---

## References

- [Jobs - Kubernetes Docs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [CronJob - Kubernetes Docs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
- [Job Patterns](https://kubernetes.io/docs/concepts/workloads/controllers/job/#job-patterns)
- [Indexed Job for Parallel Processing](https://kubernetes.io/docs/tasks/job/indexed-parallel-processing-static/)
