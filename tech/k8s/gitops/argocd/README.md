# Argo CD on Kubernetes

## What is Argo CD?

**Argo CD** is a declarative, GitOps continuous delivery tool for Kubernetes. It continuously compares the desired state declared in a Git repository against the actual state running in a cluster and reconciles any drift. The key difference from a traditional push-based CI/CD pipeline (GitHub Actions, Jenkins pushing `kubectl apply`): with Argo CD, the cluster *pulls* from Git — the Git repository is the single source of truth, not the pipeline runner. Argo CD is a CNCF graduated project.

---

## Problem It Solves

- **No audit trail for cluster changes** — with `kubectl apply` in CI pipelines, who applied what and when is scattered across pipeline logs. Argo CD traces every cluster state back to a Git commit, giving a full, immutable audit trail
- **Config drift is silent** — a `kubectl edit` or a crashed controller mutating resources creates drift that nobody notices until something breaks. Argo CD detects OutOfSync conditions continuously and can revert drift automatically via `selfHeal`
- **Rollback is manual and error-prone** — reverting a bad deploy means finding the previous pipeline run and re-triggering it. Argo CD rollback is `argocd app rollback <app> <revision>` — it applies the exact rendered manifests from a previous Git revision
- **Multi-cluster delivery requires custom scripts** — deploying the same app to 10 clusters needs 10 pipeline jobs or a loop. Argo CD's `ApplicationSet` generates `Application` resources from a template + generator, covering hundreds of clusters with a single definition

---

## Architecture

```
  Git Repository
  (desired state)
        │
        │  pull / webhook
        ▼
┌───────────────────────────────────────────────────────────────┐
│                        Argo CD (namespace: argocd)            │
│                                                               │
│  ┌──────────────┐   rendered    ┌────────────────────┐        │
│  │  Repo Server │◀─manifests───▶│  Application       │        │
│  │  (git clone, │               │  Controller        │        │
│  │  helm/kustomize│              │  (reconcile loop)  │        │
│  │  render)     │               └─────────┬──────────┘        │
│  └──────────────┘                         │ compare & sync    │
│                                           ▼                   │
│  ┌──────────────┐               ┌────────────────────┐        │
│  │  API Server  │               │  Target Cluster(s) │        │
│  │  (UI, CLI,   │               │  (live state)      │        │
│  │  gRPC/REST)  │               └────────────────────┘        │
│  └──────────────┘                                             │
│  ┌──────────────┐  ┌──────────┐  ┌────────────────────────┐  │
│  │     Dex      │  │  Redis   │  │  ApplicationSet        │  │
│  │  (SSO/OIDC)  │  │  (cache) │  │  Controller            │  │
│  └──────────────┘  └──────────┘  └────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

| Component | Role |
|-----------|------|
| **API Server** | gRPC/REST gateway for Web UI, CLI, and CI/CD webhooks. Manages CRUD for Applications, triggers sync/rollback, stores repo and cluster credentials, enforces RBAC |
| **Repo Server** | Clones Git repos locally and renders manifests (raw YAML, Helm, Kustomize, Jsonnet, plugin). Accepts `(repoURL, revision, path)` and returns Kubernetes objects |
| **Application Controller** | Continuously compares desired state (Repo Server output) against live cluster state. Marks apps OutOfSync, executes sync lifecycle hooks, drives automated sync. Reconciles every 3 minutes by default |
| **ApplicationSet Controller** | Generates `Application` resources from templates + generators (cluster, git, list, matrix, pull-request, etc.) for multi-cluster and multi-app management at scale |
| **Dex** | Optional SSO proxy — delegates authentication to external OIDC/SAML/LDAP providers. Not needed for local user auth |
| **Redis** | Caches Application state and rendered manifests. Not persistent — data loss is tolerable; state is rebuilt from Git and the cluster on next reconcile |

---

## Manifest Structure

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: argocd
---
# Application CRD — the core Argo CD resource
# One Application = one Helm release / Kustomize overlay / manifest directory
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd                              # must be in the argocd namespace
  finalizers:
    - resources-finalizer.argocd.argoproj.io    # cascading delete: removing this Application
                                                 # also deletes all managed cluster resources
spec:
  project: default                               # ArgoCD Project for RBAC and source/dest restrictions

  source:
    repoURL: https://github.com/my-org/my-repo
    targetRevision: v1.2.3                       # pin to a tag or SHA — never use a mutable branch in prod
    path: deploy/overlays/production             # path inside the repo to the manifest directory

  destination:
    server: https://kubernetes.default.svc       # in-cluster target; use registered cluster URL for remote clusters
    namespace: my-app                            # target namespace for namespace-scoped resources

  syncPolicy:
    automated:
      prune: true                                # delete cluster resources absent from Git during sync
      selfHeal: true                             # revert drift caused by out-of-band kubectl edits
      allowEmpty: false                          # safety guard: block sync if it would delete all resources
    syncOptions:
      - CreateNamespace=true                     # auto-create destination.namespace if it doesn't exist
      - ApplyOutOfSyncOnly=true                  # only apply resources that are actually out-of-sync
    retry:
      limit: 3
      backoff:
        duration: 10s
        factor: 2
        maxDuration: 3m
---
# ApplicationSet — generate multiple Application resources from a template
# Use this instead of hand-crafting one Application per cluster/environment
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: my-app-all-clusters
  namespace: argocd
spec:
  generators:
    - clusters:
        selector:
          matchLabels:
            environment: production             # target all clusters labeled environment=production
  template:
    metadata:
      name: "{{name}}-my-app"                  # {{name}} is the cluster name from the generator
    spec:
      project: default
      source:
        repoURL: https://github.com/my-org/my-repo
        targetRevision: v1.2.3
        path: "deploy/overlays/{{metadata.labels.environment}}"
      destination:
        server: "{{server}}"
        namespace: my-app
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

---

## Key Fields

### `spec.source.targetRevision`

Determines which Git commit is the desired state. This is the most critical field for reproducibility.

| Value | Behavior |
|-------|----------|
| `v1.2.3` (tag) | Immutable — recommended for production. Rollback is trivial |
| `abc1234` (SHA) | Fully immutable — safest option |
| `main` (branch) | Mutable — any push changes desired state immediately. Acceptable only for dev environments |
| `HEAD` | Latest commit on the default branch — same risk as a branch name |

For Helm chart sources (`chart` + `repoURL` pointing to an OCI/Helm registry), `targetRevision` is the chart version and must be an exact semver (e.g., `5.21.0`), never a range.

---

### `spec.syncPolicy.automated`

Controls whether Argo CD syncs automatically or requires a manual trigger.

| Field | Default | Effect when `true` |
|-------|---------|-------------------|
| `prune` | `false` | Resources removed from Git are deleted from the cluster on next sync |
| `selfHeal` | `false` | Out-of-band changes (manual `kubectl edit`, controller mutations) are reverted automatically |
| `allowEmpty` | `false` | Allows sync to proceed even if it would delete every managed resource |

Without `selfHeal: true`, the GitOps guarantee breaks — manual changes to the cluster are never reverted. The recommended pattern is to enable both `prune` and `selfHeal` in all environments and use Git branch protection as the safety gate instead.

---

### `spec.syncPolicy.syncOptions`

Feature flags that modify sync behavior per Application.

| Option | When to use |
|--------|------------|
| `CreateNamespace=true` | App owns a namespace that may not exist yet |
| `ApplyOutOfSyncOnly=true` | Large apps where applying all resources on every sync overloads the API server |
| `PruneLast=true` | Resources must be created before old ones are removed (e.g., replacing a Service) |
| `Replace=true` | Resources have immutable fields that `kubectl apply` cannot patch (requires replace instead) |
| `Validate=false` | CRDs that aren't yet registered in the cluster cause schema validation errors |
| `RespectIgnoreDifferences=true` | Apply `spec.ignoreDifferences` rules during sync, not just during diff display |

---

### `spec.ignoreDifferences`

Fields that Argo CD should exclude from the OutOfSync calculation. Essential when Kubernetes controllers mutate fields after apply (HPA managing `spec.replicas`, controllers injecting annotations, etc.).

```yaml
ignoreDifferences:
  - group: apps
    kind: Deployment
    jsonPointers:
      - /spec/replicas          # HPA manages this — ignore drift caused by autoscaling
  - group: ""
    kind: Secret
    jsonPointers:
      - /data                   # externally managed secrets — ArgoCD should not own the data
```

---

### `metadata.finalizers`

| Finalizer | Effect on Application deletion |
|-----------|-------------------------------|
| `resources-finalizer.argocd.argoproj.io` | Cascading delete — all managed cluster resources are deleted before the Application object is removed |
| `resources-finalizer.argocd.argoproj.io/background` | Same, but uses Kubernetes background propagation policy |
| *(none)* | Application object is deleted; cluster resources are left orphaned (no cleanup) |

Add the finalizer only when you want Argo CD to own the full lifecycle of the resources. For shared infrastructure (Namespaces, CRDs), omit it.

---

## Sync Waves and Hook Ordering

Argo CD applies resources in waves to control ordering within a single sync. Resources in lower waves are applied and become healthy before higher waves start.

```yaml
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "0"    # default wave; lower numbers go first
```

Hooks execute at specific points in the sync lifecycle:

| Hook | When it runs |
|------|-------------|
| `PreSync` | Before any resources are applied — database migrations, pre-checks |
| `Sync` | During the main apply phase |
| `PostSync` | After all resources are Healthy — smoke tests, notifications |
| `SyncFail` | If the sync fails — cleanup, alerts |
| `Skip` | Never applied by Argo CD (managed externally) |

Hooks are standard Kubernetes Jobs/Pods with the annotation `argocd.argoproj.io/hook: PreSync`. They are **not** Helm hooks — Argo CD renders all manifests via `helm template` and then applies the output, bypassing the Helm hook lifecycle entirely.

---

## Common Pitfalls

**App stuck OutOfSync after a successful sync:** Kustomize injects `app.kubernetes.io/instance` labels that conflict with Argo CD's own instance label. Fix: set `application.instanceLabelKey: argocd.argoproj.io/instance` in the `argocd-cm` ConfigMap, then re-sync affected apps.

**Perpetual OutOfSync from normalized resource fields:** Kubernetes normalizes resource quantities (`1000m` → `1`, `3072Mi` → `3Gi`). The Git manifest never matches what the API server stores. Fix: use `spec.ignoreDifferences` with `jsonPointers` targeting the affected fields rather than changing values in Git.

**HPA and `spec.replicas` conflict:** When HPA manages replica counts, the value in Git diverges from what HPA sets in the cluster. Argo CD detects this as drift and either reports OutOfSync forever or (with `selfHeal`) fights HPA in a loop. Fix: omit `spec.replicas` from the manifest in Git entirely, or add it to `spec.ignoreDifferences`.

**`argocd app set` parameter overrides in production:** Overrides set via CLI or UI are stored in the Application object, not in Git — invisible to code reviewers and not recoverable from Git. Every configuration change must be a Git commit. Use separate `valueFiles` per environment instead.

**`helm ls` shows no trace of ArgoCD-managed releases:** Argo CD uses `helm template` + `kubectl apply`, not `helm install`. Helm has no release record. Teams expecting standard Helm workflows (`helm rollback`, `helm history`) are surprised. Rollback goes through Argo CD, not Helm.

**Unpinned `targetRevision` in production:** Using a branch name means any merged PR immediately changes desired state in production. Pin to a tag or SHA; use a separate branch-tracked Application for dev/staging only.

**Finalizer set on shared infrastructure:** The `resources-finalizer` on a Namespace Application will delete the Namespace (and everything in it) when the Application is removed. Never add the finalizer to apps managing shared infrastructure like Namespaces, CRDs, or cluster-level RBAC.

**Application definitions not stored in Git:** Apps created via UI or CLI but not committed to Git cannot be recovered if Argo CD is reinstalled. All `Application` and `ApplicationSet` CRDs must live in a Git repository (the "app of apps" pattern).

---

## References

- [Argo CD Architecture](https://argo-cd.readthedocs.io/en/stable/operator-manual/architecture/)
- [Application Specification](https://argo-cd.readthedocs.io/en/stable/user-guide/application-specification/)
- [Installation](https://argo-cd.readthedocs.io/en/stable/operator-manual/installation/)
- [Best Practices](https://argo-cd.readthedocs.io/en/stable/user-guide/best_practices/)
- [FAQ](https://argo-cd.readthedocs.io/en/stable/faq/)
- [ApplicationSet Generators](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators/)
- [Argo CD Helm Chart](https://github.com/argoproj/argo-helm/tree/main/charts/argo-cd)
