#!/usr/bin/env bash
# sed -i 's/\r$//' cnpg.sh
set -euo pipefail

# Chart 0.29.0 maps to operator 1.29.1 (ships the CVE-2026-44477 fix: metrics
# exporter runs as a dedicated pg_monitor role, not superuser). See release notes
# for the chart-to-operator version mapping.
CNPG_VERSION="${CNPG_VERSION:-0.29.0}"
CNPG_NAMESPACE="${CNPG_NAMESPACE:-cnpg-system}"
CNPG_RELEASE="${CNPG_RELEASE:-cnpg}"
DRY_RUN=false

log()  { echo "[$(date +'%Y-%m-%d %H:%M:%S')] [INFO]  $*"; }
err()  { echo "[$(date +'%Y-%m-%d %H:%M:%S')] [ERROR] $*" >&2; }

trap 'err "Script failed at line $LINENO"' ERR

run() {
    if $DRY_RUN; then
        echo "[DRY-RUN] $*"
    else
        "$@"
    fi
}

usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --version VER    CloudNativePG chart version (default: $CNPG_VERSION)
  --namespace NS   Operator namespace (default: $CNPG_NAMESPACE)
  --release NAME   Helm release name (default: $CNPG_RELEASE)
  --dry-run        Print commands without executing
  -h, --help       Show this help message

Environment variables:
  CNPG_VERSION, CNPG_NAMESPACE, CNPG_RELEASE
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)   CNPG_VERSION="$2";   shift 2 ;;
        --namespace) CNPG_NAMESPACE="$2"; shift 2 ;;
        --release)   CNPG_RELEASE="$2";   shift 2 ;;
        --dry-run)   DRY_RUN=true;        shift   ;;
        -h|--help)   usage ;;
        *) err "Unknown option: $1"; usage ;;
    esac
done

# --- Install operator via Helm ---
# The operator runs once in cnpg-system and watches Cluster CRs in every namespace.
# Install it before applying the 01..04 manifests (namespace/storage/secret/cluster).

log "Installing CloudNativePG operator (chart ${CNPG_VERSION})"
log "  Release   : $CNPG_RELEASE"
log "  Namespace : $CNPG_NAMESPACE"

run helm repo add cnpg https://cloudnative-pg.github.io/charts
run helm repo update cnpg

# upgrade --install is idempotent: re-run to upgrade, first run installs
run helm upgrade --install "$CNPG_RELEASE" cnpg/cloudnative-pg \
    --version "$CNPG_VERSION" \
    --namespace "$CNPG_NAMESPACE" \
    --create-namespace

# --- Verify ---
# Helm names the deployment <release>-cloudnative-pg; use a label selector so this
# does not depend on the exact name.

log "Waiting for operator deployment to be ready..."
run kubectl wait --timeout=5m \
    -n "$CNPG_NAMESPACE" \
    deployment \
    -l app.kubernetes.io/name=cloudnative-pg \
    --for=condition=Available

log "Verification"
printf "  %-20s %s\n" "helm release:" \
    "$(helm status "$CNPG_RELEASE" -n "$CNPG_NAMESPACE" 2>/dev/null | grep STATUS | awk '{print $2}' || echo 'not found')"
printf "  %-20s %s\n" "operator pod:" \
    "$(kubectl get pods -n "$CNPG_NAMESPACE" -l app.kubernetes.io/name=cloudnative-pg \
        --no-headers 2>/dev/null | awk '{print $1, $3}' | head -1 || echo 'not found')"

echo ""
log "CloudNativePG operator installed. Next: apply 01-namespace -> 02-storage -> 03-secret -> 04-cluster."
