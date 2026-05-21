# CronJob in Kubernetes

## What is a CronJob?

**CronJob** is a workload controller that runs tasks **on a recurring schedule** — similar to cron on Linux but running inside a Kubernetes cluster. Each time the schedule fires, CronJob creates a **Job**, which in turn creates a Pod and runs the task to completion.

CronJob manages the schedule; the actual execution is handled by **Job**. Understanding Job means understanding 80% of CronJob — CronJob simply adds a scheduling layer on top.

---

## Problem It Solves

Some tasks need to run **repeatedly on a fixed schedule** without manual intervention:

- **Database backup** — run every night at 2 AM, dump the DB and upload to S3
- **Email notifications** — send a weekly summary report to users every Monday morning
- **Data cleanup** — delete expired logs/records once a week
- **Data synchronization** — pull data from an external API every 15 minutes

Without CronJob, you'd need to set up a cron job on a separate VM outside the cluster — harder to monitor, unable to leverage cluster resources, and it creates an additional failure point outside Kubernetes.

---

## How CronJob Works

```
Schedule fires
  → CronJob controller creates a Job object
    → Job controller creates a Pod
      → Pod runs the container until exit 0
        → Job = Complete → CronJob records history
```

1. **CronJob controller** (inside kube-controller-manager) evaluates the schedule every 10 seconds
2. When the schedule fires, the controller creates a new Job object from `jobTemplate`
3. The Job manages the Pod lifecycle — retries, timeout, completion — according to the template config
4. After the Job finishes, CronJob retains success/failure history up to the configured limits

**Note:** the CronJob controller does not guarantee second-level precision. Jobs may start a few seconds after the scheduled time.

---

## Schedule Syntax

CronJob uses the standard 5-field cron syntax:

```
┌───────────── minute (0–59)
│ ┌───────────── hour (0–23)
│ │ ┌───────────── day of month (1–31)
│ │ │ ┌───────────── month (1–12)
│ │ │ │ ┌───────────── day of week (0–6, 0 = Sunday)
│ │ │ │ │
* * * * *
```

| Schedule | Meaning |
|----------|---------|
| `0 2 * * *` | Every day at 2:00 AM |
| `*/15 * * * *` | Every 15 minutes |
| `0 9 * * 1` | Every Monday at 9:00 AM |
| `0 0 1 * *` | First day of every month at midnight |
| `0 8-18 * * 1-5` | Every hour during business hours (8 AM–6 PM, Mon–Fri) |

The schedule uses the timezone of kube-controller-manager (typically UTC). Use the `timeZone` field to specify a timezone explicitly.

---

## Manifest Structure

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: db-backup
  namespace: default
spec:
  schedule: "0 2 * * *"          # run every day at 2:00 UTC
  timeZone: "Asia/Ho_Chi_Minh"   # timezone applied to the schedule

  concurrencyPolicy: Forbid       # skip new run if previous job is still running
  startingDeadlineSeconds: 300    # skip if more than 5 minutes late
  successfulJobsHistoryLimit: 3   # keep the 3 most recent successful Jobs
  failedJobsHistoryLimit: 1       # keep the 1 most recent failed Job
  suspend: false                  # set true to pause without deleting

  jobTemplate:
    spec:
      backoffLimit: 2
      activeDeadlineSeconds: 3600
      ttlSecondsAfterFinished: 86400

      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: backup
              image: myapp-backup:1.0.0
              command: ["sh", "-c", "pg_dump $DB_URL | gzip | aws s3 cp - s3://backups/$(date +%F).sql.gz"]
              env:
                - name: DB_URL
                  valueFrom:
                    secretKeyRef:
                      name: db-secret
                      key: url
              resources:
                requests:
                  cpu: 200m
                  memory: 256Mi
                limits:
                  cpu: 500m
                  memory: 512Mi
```

---

## Key Parameters

### concurrencyPolicy

Controls what happens when the next scheduled run fires while the previous Job is still running:

| Value | Behavior |
|-------|----------|
| `Allow` (default) | Create the new Job — both run in parallel |
| `Forbid` | Skip this run, wait for the previous Job to finish |
| `Replace` | Delete the running Job, create a new one in its place |

Use `Forbid` for tasks that cannot run in parallel (backup, migration). Use `Allow` when each run is completely independent.

### startingDeadlineSeconds

If a CronJob misses its scheduled time (controller restarted, cluster overloaded), this is the maximum allowed delay. After this window, the missed run is skipped entirely.

```yaml
startingDeadlineSeconds: 300  # allow up to 5 minutes late
```

Without this — when the controller restarts after a long downtime, it may create a burst of catch-up Jobs for all missed runs at once.

### successfulJobsHistoryLimit and failedJobsHistoryLimit

Number of completed Jobs (successful and failed) retained for log inspection and status review. Defaults are `3` and `1`.

Set to `0` to delete immediately after completion — no history available, but fewer objects in the cluster.

### suspend

```bash
# Pause the CronJob (stops creating new Jobs; running Jobs are unaffected)
kubectl patch cronjob db-backup -p '{"spec":{"suspend":true}}'

# Resume
kubectl patch cronjob db-backup -p '{"spec":{"suspend":false}}'
```

Use during maintenance windows or release freezes when you need to disable the recurring task without deleting the CronJob.

---

## Triggering a Manual Run

Create a Job from the CronJob immediately without waiting for the schedule:

```bash
kubectl create job db-backup-manual --from=cronjob/db-backup
```

Useful for testing a CronJob after deployment, or running a catch-up when an important scheduled run was missed.

---

## Monitoring and Debugging

```bash
# Check CronJob status and last schedule time
kubectl get cronjob db-backup

# List Jobs created by the CronJob
kubectl get jobs -l job-name

# View logs of the most recent Job
kubectl logs job/db-backup-28123456

# View full run history and events
kubectl describe cronjob db-backup
```

---

## Common Pitfalls

**Missing `concurrencyPolicy: Forbid` for long-running tasks → overlapping Jobs.** If a backup takes 90 minutes but the schedule is every hour, the second Job starts before the first finishes. Two tasks reading/writing the same data source → race condition or duplicate output.

**Missing `startingDeadlineSeconds` → burst of catch-up Jobs after downtime.** When the cluster restarts after several hours, the CronJob controller may create dozens of Jobs to compensate for all missed runs. Always set `startingDeadlineSeconds` to cap the allowed catch-up window.

**Missing `activeDeadlineSeconds` in jobTemplate → Job hangs indefinitely.** If the container deadlocks or waits on I/O forever, the Job never terminates — and blocks the next scheduled run under `Forbid`. Always set a timeout on Jobs in production.

**Schedule in UTC mistaken for local time.** `"0 2 * * *"` is 2 AM UTC, which is 9 AM UTC+7. Always set `timeZone` explicitly to avoid confusion.

**`successfulJobsHistoryLimit: 0` → no logs when something goes wrong.** If you need to investigate a failed run, the Job has already been deleted. Keep at least `failedJobsHistoryLimit: 1` in production.

**Non-idempotent task + retry → duplicate data.** CronJob retries via the Job's `backoffLimit`. If the task processed halfway then crashed, the retry starts from the beginning. Design tasks to be idempotent, or set `backoffLimit: 0` to fail immediately on the first error.

---

## References

- [CronJob - Kubernetes Docs](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
- [Running Automated Tasks with a CronJob](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/)
- [Jobs - Kubernetes Docs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
