#!/usr/bin/env bash
# sed -i 's/\r$//' longhorn.sh
set -euo pipefail

LH_VERSION="${LH_VERSION:-1.12.0}"
LH_NAMESPACE="${LH_NAMESPACE:-longhorn-system}"
LH_RELEASE="${LH_RELEASE:-longhorn}"
LH_GATEWAY_NAME="${LH_GATEWAY_NAME:-eg}"
LH_GATEWAY_NS="${LH_GATEWAY_NS:-envoy-gateway-system}"
LH_HOSTNAME="${LH_HOSTNAME:-longhorn.local}"
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
  --version VER        Longhorn chart version (default: $LH_VERSION)
  --namespace NS       Target namespace (default: $LH_NAMESPACE)
  --release NAME       Helm release name (default: $LH_RELEASE)
  --gateway-name NAME  Gateway resource name for HTTPRoute (default: $LH_GATEWAY_NAME)
  --gateway-ns NS      Gateway resource namespace (default: $LH_GATEWAY_NS)
  --hostname HOST      Hostname to access Longhorn UI (default: $LH_HOSTNAME)
  --dry-run            Print commands without executing
  -h, --help           Show this help message

Environment variables:
  LH_VERSION, LH_NAMESPACE, LH_RELEASE, LH_GATEWAY_NAME, LH_GATEWAY_NS, LH_HOSTNAME
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)   LH_VERSION="$2";   shift 2 ;;
        --namespace) LH_NAMESPACE="$2"; shift 2 ;;
        --release)      LH_RELEASE="$2";      shift 2 ;;
        --gateway-name) LH_GATEWAY_NAME="$2"; shift 2 ;;
        --gateway-ns)   LH_GATEWAY_NS="$2";   shift 2 ;;
        --hostname)     LH_HOSTNAME="$2";     shift 2 ;;
        --dry-run)      DRY_RUN=true;          shift   ;;
        -h|--help)   usage ;;
        *) err "Unknown option: $1"; usage ;;
    esac
done

# --- Longhorn ---

log "Installing Longhorn v${LH_VERSION}"
log "  Release   : $LH_RELEASE"
log "  Namespace : $LH_NAMESPACE"
log "  Gateway   : ${LH_GATEWAY_NAME} (ns: ${LH_GATEWAY_NS})"
log "  Hostname  : $LH_HOSTNAME"

run helm repo add longhorn https://charts.longhorn.io
run helm repo update longhorn

run helm install "$LH_RELEASE" longhorn/longhorn \
    --version "$LH_VERSION" \
    --namespace "$LH_NAMESPACE" \
    --create-namespace \
    --set httproute.enabled=true \
    --set "httproute.parentRefs[0].name=$LH_GATEWAY_NAME" \
    --set "httproute.parentRefs[0].namespace=$LH_GATEWAY_NS" \
    --set "httproute.hostnames[0]=$LH_HOSTNAME"

# --- Verification ---

log "Verification"
printf "  %-20s %s\n" "helm release:" "$(helm status "$LH_RELEASE" -n "$LH_NAMESPACE" --no-headers 2>/dev/null | head -1 || echo 'not found')"
printf "  %-20s %s\n" "httproute:" "$(kubectl get httproute -n "$LH_NAMESPACE" -o name 2>/dev/null || echo 'not found')"

echo ""
log "Longhorn UI accessible at: http://$LH_HOSTNAME"
log "Longhorn installation completed successfully!"