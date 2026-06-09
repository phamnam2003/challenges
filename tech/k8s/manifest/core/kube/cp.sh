#!/usr/bin/env bash
# sed -i 's/\r$//' cp.sh
set -euo pipefail

CP_ENDPOINT="${CP_ENDPOINT:-192.168.79.11:6443}"
POD_CIDR="${POD_CIDR:-10.200.0.0/16}"
SVC_CIDR="${SVC_CIDR:-10.96.0.0/12}"
HELM_VERSION="${HELM_VERSION:-4.2.0}"
DRY_RUN=false

log()  { echo "[$(date +'%Y-%m-%d %H:%M:%S')] [INFO]  $*"; }
warn() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] [WARN]  $*" >&2; }
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
  --endpoint HOST:PORT   Control-plane endpoint (default: $CP_ENDPOINT)
  --pod-cidr CIDR        Pod network CIDR (default: $POD_CIDR)
  --svc-cidr CIDR        Service CIDR (default: $SVC_CIDR)
  --helm-version VER     Helm version to install (default: $HELM_VERSION)
  --dry-run              Print commands without executing
  -h, --help             Show this help message

Environment variables:
  CP_ENDPOINT, POD_CIDR, SVC_CIDR, HELM_VERSION  (same as above flags)
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --endpoint)      CP_ENDPOINT="$2"; shift 2 ;;
        --pod-cidr)      POD_CIDR="$2";    shift 2 ;;
        --svc-cidr)      SVC_CIDR="$2";    shift 2 ;;
        --helm-version)  HELM_VERSION="$2"; shift 2 ;;
        --dry-run)       DRY_RUN=true;      shift   ;;
        -h|--help)       usage ;;
        *) err "Unknown option: $1"; usage ;;
    esac
done

# --- kubeadm init ---

log "Initializing control plane"
log "  Endpoint : $CP_ENDPOINT"
log "  Pod CIDR : $POD_CIDR"
log "  Svc CIDR : $SVC_CIDR"

run sudo kubeadm init \
    --control-plane-endpoint="$CP_ENDPOINT" \
    --pod-network-cidr="$POD_CIDR" \
    --service-cidr="$SVC_CIDR" \
    --skip-phases=addon/kube-proxy

# kubeconfig cho user hiện tại để kubectl hoạt động ngay sau init
log "Configuring kubeconfig for current user"
if ! $DRY_RUN; then
    mkdir -p "$HOME/.kube"
    sudo cp /etc/kubernetes/admin.conf "$HOME/.kube/config"
    sudo chown "$(id -u):$(id -g)" "$HOME/.kube/config"
else
    echo "[DRY-RUN] Copy admin.conf to $HOME/.kube/config"
fi

# --- Helm ---

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  HELM_ARCH="linux-amd64" ;;
    aarch64) HELM_ARCH="linux-arm64" ;;
    *) err "Unsupported architecture: $ARCH"; exit 1 ;;
esac

HELM_TARBALL="helm-v${HELM_VERSION}-${HELM_ARCH}.tar.gz"
HELM_URL="https://get.helm.sh/${HELM_TARBALL}"

if command -v helm &>/dev/null; then
    CURRENT_HELM=$(helm version --short 2>/dev/null || echo "unknown")
    log "Helm already installed: $CURRENT_HELM"
    log "Upgrading to v${HELM_VERSION}"
fi

log "Installing Helm v${HELM_VERSION} (${HELM_ARCH})"
run sudo wget -q "$HELM_URL"
run sudo tar -xzf "$HELM_TARBALL"
run sudo mv "${HELM_ARCH}/helm" /usr/local/bin/helm
run sudo rm -rf "$HELM_ARCH" "$HELM_TARBALL"

# --- Verification ---

log "Verification"
printf "  %-20s %s\n" "kubeadm:"  "$(kubeadm version -o short 2>/dev/null || echo 'not found')"
printf "  %-20s %s\n" "kubectl:"  "$(kubectl version --client 2>/dev/null | head -1 || echo 'not found')"
printf "  %-20s %s\n" "helm:"     "$(helm version --short 2>/dev/null || echo 'not found')"

echo ""
log "Control plane initialized successfully!"