#!/usr/bin/env bash
# sed -i 's/\r$//' cilium.sh
set -euo pipefail

API_SERVER_IP="${API_SERVER_IP:-192.168.79.11}"
API_SERVER_PORT="${API_SERVER_PORT:-6443}"
CILIUM_VERSION="${CILIUM_VERSION:-1.19.4}"
POD_CIDR="${POD_CIDR:-10.200.0.0/16}"
OPERATOR_REPLICAS="${OPERATOR_REPLICAS:-2}"
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
  --api-server-ip IP       API server IP (default: $API_SERVER_IP)
  --api-server-port PORT   API server port (default: $API_SERVER_PORT)
  --cilium-version VER     Cilium chart version (default: $CILIUM_VERSION)
  --pod-cidr CIDR          Pod network CIDR for native routing (default: $POD_CIDR)
  --operator-replicas N    Cilium operator replicas (default: $OPERATOR_REPLICAS)
  --dry-run                Print commands without executing
  -h, --help               Show this help message

Environment variables:
  API_SERVER_IP, API_SERVER_PORT, CILIUM_VERSION, POD_CIDR, OPERATOR_REPLICAS
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --api-server-ip)     API_SERVER_IP="$2";     shift 2 ;;
        --api-server-port)   API_SERVER_PORT="$2";   shift 2 ;;
        --cilium-version)    CILIUM_VERSION="$2";    shift 2 ;;
        --pod-cidr)          POD_CIDR="$2";          shift 2 ;;
        --operator-replicas) OPERATOR_REPLICAS="$2"; shift 2 ;;
        --dry-run)           DRY_RUN=true;           shift   ;;
        -h|--help)           usage ;;
        *) err "Unknown option: $1"; usage ;;
    esac
done

# --- Cilium CLI ---

log "Installing Cilium CLI"

CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
log "  CLI version: $CILIUM_CLI_VERSION"

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  CLI_ARCH="amd64" ;;
    aarch64) CLI_ARCH="arm64" ;;
    *) err "Unsupported architecture: $ARCH"; exit 1 ;;
esac

CLI_TARBALL="cilium-linux-${CLI_ARCH}.tar.gz"
CLI_URL="https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}"

run curl -L --fail --remote-name-all "${CLI_URL}/${CLI_TARBALL}"{,.sha256sum}
run sha256sum --check "${CLI_TARBALL}.sha256sum"
run sudo tar xzvfC "$CLI_TARBALL" /usr/local/bin
run rm -f "${CLI_TARBALL}"{,.sha256sum}

log "Cilium CLI installed: $(cilium version --client 2>/dev/null || echo 'unknown')"

# --- Cilium CNI via Helm ---

log "Installing Cilium v${CILIUM_VERSION} as CNI"
log "  API Server : ${API_SERVER_IP}:${API_SERVER_PORT}"
log "  Pod CIDR   : $POD_CIDR"
log "  Replicas   : $OPERATOR_REPLICAS"

run helm install cilium oci://quay.io/cilium/charts/cilium \
    --version "$CILIUM_VERSION" \
    --namespace kube-system \
    --create-namespace \
    --set kubeProxyReplacement=true \
    --set k8sServiceHost="$API_SERVER_IP" \
    --set k8sServicePort="$API_SERVER_PORT" \
    --set routingMode=native \
    --set autoDirectNodeRoutes=true \
    --set ipam.mode=kubernetes \
    --set bpf.masquerade=true \
    --set loadBalancer.algorithm=maglev \
    --set hubble.enabled=true \
    --set hubble.relay.enabled=true \
    --set ipv4NativeRoutingCIDR="$POD_CIDR" \
    --set bpf.hostLegacyRouting=false \
    --set socketLB.enabled=true \
    --set socketLB.hostNamespaceOnly=true \
    --set envoy.enabled=false \
    --set l7Proxy=false \
    --set operator.replicas="$OPERATOR_REPLICAS"

# --- Verification ---

log "Verification"
printf "  %-20s %s\n" "cilium-cli:" "$(cilium version --client 2>/dev/null || echo 'not found')"
printf "  %-20s %s\n" "helm release:" "$(helm status cilium -n kube-system --no-headers 2>/dev/null | head -1 || echo 'not found')"

echo ""
log "Cilium CNI installation completed successfully!"
