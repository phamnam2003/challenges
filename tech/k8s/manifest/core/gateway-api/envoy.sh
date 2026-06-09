#!/usr/bin/env bash
# sed -i 's/\r$//' envoy.sh
set -euo pipefail

EG_VERSION="${EG_VERSION:-v1.8.1}"
EG_NAMESPACE="${EG_NAMESPACE:-envoy-gateway-system}"
EG_RELEASE="${EG_RELEASE:-eg}"
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
  --version VER        Envoy Gateway chart version (default: $EG_VERSION)
  --namespace NS       Target namespace (default: $EG_NAMESPACE)
  --release NAME       Helm release name (default: $EG_RELEASE)
  --dry-run            Print commands without executing
  -h, --help           Show this help message

Environment variables:
  EG_VERSION, EG_NAMESPACE, EG_RELEASE
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)   EG_VERSION="$2";   shift 2 ;;
        --namespace) EG_NAMESPACE="$2"; shift 2 ;;
        --release)   EG_RELEASE="$2";   shift 2 ;;
        --dry-run)   DRY_RUN=true;      shift   ;;
        -h|--help)   usage ;;
        *) err "Unknown option: $1"; usage ;;
    esac
done

# --- Envoy Gateway ---

log "Installing Envoy Gateway ${EG_VERSION}"
log "  Release   : $EG_RELEASE"
log "  Namespace : $EG_NAMESPACE"

run helm install "$EG_RELEASE" oci://docker.io/envoyproxy/gateway-helm \
    --version "$EG_VERSION" \
    --namespace "$EG_NAMESPACE" \
    --create-namespace

# --- Verification ---

log "Verification"
printf "  %-20s %s\n" "helm release:" "$(helm status "$EG_RELEASE" -n "$EG_NAMESPACE" --no-headers 2>/dev/null | head -1 || echo 'not found')"

echo ""
log "Envoy Gateway installation completed successfully!"