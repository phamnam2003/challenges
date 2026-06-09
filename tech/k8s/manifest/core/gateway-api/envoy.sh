#!/usr/bin/env bash
# sed -i 's/\r$//' envoy.sh
set -euo pipefail

EG_VERSION="${EG_VERSION:-v1.8.1}"
EG_NAMESPACE="${EG_NAMESPACE:-envoy-gateway-system}"
EG_RELEASE="${EG_RELEASE:-eg}"
GATEWAY_API_CHANNEL="${GATEWAY_API_CHANNEL:-standard}"
DRY_RUN=false
UPGRADE=false

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
  --version VER              Envoy Gateway chart version (default: $EG_VERSION)
  --namespace NS             Target namespace (default: $EG_NAMESPACE)
  --release NAME             Helm release name (default: $EG_RELEASE)
  --gateway-api-channel CH   Gateway API channel: standard|experimental (default: $GATEWAY_API_CHANNEL)
  --upgrade                  Upgrade existing installation instead of fresh install
  --dry-run                  Print commands without executing
  -h, --help                 Show this help message

Environment variables:
  EG_VERSION, EG_NAMESPACE, EG_RELEASE, GATEWAY_API_CHANNEL
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)               EG_VERSION="$2";           shift 2 ;;
        --namespace)             EG_NAMESPACE="$2";         shift 2 ;;
        --release)               EG_RELEASE="$2";           shift 2 ;;
        --gateway-api-channel)   GATEWAY_API_CHANNEL="$2";  shift 2 ;;
        --upgrade)               UPGRADE=true;              shift   ;;
        --dry-run)               DRY_RUN=true;              shift   ;;
        -h|--help)               usage ;;
        *) err "Unknown option: $1"; usage ;;
    esac
done

# --- Step 1: CRDs via gateway-crds-helm ---
#
# Helm cannot update CRDs placed in the /crds folder on `helm upgrade`, and
# `helm install --skip-crds` skips ALL CRDs including Envoy's own. The official
# recommendation is to render the dedicated gateway-crds-helm chart and pipe it
# through kubectl apply --server-side, which handles both initial install and
# idempotent upgrades correctly.

log "Installing CRDs (gateway-crds-helm ${EG_VERSION})"
log "  Gateway API channel : $GATEWAY_API_CHANNEL"

run helm template eg-crds oci://docker.io/envoyproxy/gateway-crds-helm \
    --version "$EG_VERSION" \
    --set crds.gatewayAPI.enabled=true \
    --set crds.gatewayAPI.channel="$GATEWAY_API_CHANNEL" \
    --set crds.envoyGateway.enabled=true \
    | run kubectl apply --server-side -f -

log "Waiting for CRDs to become established..."
run kubectl wait --for=condition=Established --timeout=60s \
    crd/gateways.gateway.networking.k8s.io \
    crd/httproutes.gateway.networking.k8s.io \
    crd/gatewayclasses.gateway.networking.k8s.io \
    crd/envoyproxies.gateway.envoyproxy.io \
    crd/envoypatchpolicies.gateway.envoyproxy.io

# --- Step 2: ValidatingAdmissionPolicy ownership migration (v1.8.1 breaking change) ---
#
# Before v1.8.1, safe-upgrades ValidatingAdmissionPolicy/Binding were part of the
# CRD bundle. They moved into gateway-helm templates in v1.8.1. On upgrades where
# these resources already exist without Helm ownership labels, `helm upgrade` will
# fail because Helm cannot adopt them. We must annotate them first so Helm owns them.

if $UPGRADE; then
    log "Checking ValidatingAdmissionPolicy ownership for v1.8.1 migration..."

    VAP="safe-upgrades.gateway.networking.k8s.io"
    if kubectl get validatingadmissionpolicy "$VAP" &>/dev/null; then
        CURRENT_MANAGER=$(kubectl get validatingadmissionpolicy "$VAP" \
            -o jsonpath='{.metadata.labels.app\.kubernetes\.io/managed-by}' 2>/dev/null || true)

        if [[ "$CURRENT_MANAGER" != "Helm" ]]; then
            log "  Adopting ValidatingAdmissionPolicy into Helm ownership..."
            run kubectl annotate validatingadmissionpolicy "$VAP" \
                "meta.helm.sh/release-name=${EG_RELEASE}" \
                "meta.helm.sh/release-namespace=${EG_NAMESPACE}" \
                --overwrite
            run kubectl label validatingadmissionpolicy "$VAP" \
                "app.kubernetes.io/managed-by=Helm" \
                --overwrite

            run kubectl annotate validatingadmissionpolicybinding "$VAP" \
                "meta.helm.sh/release-name=${EG_RELEASE}" \
                "meta.helm.sh/release-namespace=${EG_NAMESPACE}" \
                --overwrite
            run kubectl label validatingadmissionpolicybinding "$VAP" \
                "app.kubernetes.io/managed-by=Helm" \
                --overwrite
        else
            log "  ValidatingAdmissionPolicy already Helm-managed, skipping."
        fi
    fi
fi

# --- Step 3: Install/upgrade gateway-helm with --skip-crds ---
#
# CRDs are already applied above via server-side apply, so --skip-crds prevents
# Helm from attempting to re-apply them during install or upgrade.

log "Installing Envoy Gateway ${EG_VERSION}"
log "  Release   : $EG_RELEASE"
log "  Namespace : $EG_NAMESPACE"

if $UPGRADE; then
    run helm upgrade "$EG_RELEASE" oci://docker.io/envoyproxy/gateway-helm \
        --version "$EG_VERSION" \
        --namespace "$EG_NAMESPACE" \
        --skip-crds
else
    run helm install "$EG_RELEASE" oci://docker.io/envoyproxy/gateway-helm \
        --version "$EG_VERSION" \
        --namespace "$EG_NAMESPACE" \
        --create-namespace \
        --skip-crds
fi

# --- Step 4: Verify ---

log "Waiting for Envoy Gateway deployment to be ready..."
run kubectl wait --timeout=5m \
    -n "$EG_NAMESPACE" \
    deployment/envoy-gateway \
    --for=condition=Available

log "Verification"
printf "  %-24s %s\n" "helm release:" \
    "$(helm status "$EG_RELEASE" -n "$EG_NAMESPACE" 2>/dev/null | grep STATUS | awk '{print $2}' || echo 'not found')"
printf "  %-24s %s\n" "envoy-gateway pod:" \
    "$(kubectl get pods -n "$EG_NAMESPACE" -l app.kubernetes.io/name=envoy-gateway \
        --no-headers 2>/dev/null | awk '{print $1, $3}' | head -1 || echo 'not found')"

echo ""
log "Envoy Gateway ${EG_VERSION} installation completed successfully!"
