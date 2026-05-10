#!/usr/bin/env bash
set -euo pipefail

KUBE_MINOR="${KUBE_MINOR:-1.33}"
KUBE_VERSION="${KUBE_VERSION:-1.33.2-1.1}"
NODE_IP="${NODE_IP:-}"
DRY_RUN=false

log()  { echo "[$(date +'%Y-%m-%d %H:%M:%S')] [INFO]  $*"; }
warn() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] [WARN]  $*" >&2; }
err()  { echo "[$(date +'%Y-%m-%d %H:%M:%S')] [ERROR] $*" >&2; }

trap 'err "Script failed at line $LINENO"' ERR

# Wraps commands so --dry-run can print them instead of executing.
# Pipe-based commands need their own if/else blocks since shell pipes
# can't be passed as arguments to a function.
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
  --version VERSION    Kubernetes package version (default: $KUBE_VERSION)
  --minor MINOR        Kubernetes minor version for repo URL (default: $KUBE_MINOR)
  --node-ip IP         Node IP address (default: auto-detect via hostname -I)
  --dry-run            Print commands without executing
  -h, --help           Show this help message

Environment variables:
  KUBE_MINOR, KUBE_VERSION, NODE_IP  (same as above flags)
EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version)  KUBE_VERSION="$2"; shift 2 ;;
        --minor)    KUBE_MINOR="$2";   shift 2 ;;
        --node-ip)  NODE_IP="$2";      shift 2 ;;
        --dry-run)  DRY_RUN=true;      shift   ;;
        -h|--help)  usage ;;
        *) err "Unknown option: $1"; usage ;;
    esac
done

# Fails early before any package is installed so the user doesn't end up
# with a half-configured node that's hard to clean up.
preflight() {
    log "Running pre-flight checks"

    if [[ ! -f /etc/os-release ]]; then
        err "Cannot detect OS — /etc/os-release not found"; exit 1
    fi
    # shellcheck source=/dev/null
    . /etc/os-release
    if [[ "$ID" != "ubuntu" && "$ID" != "debian" ]]; then
        err "Unsupported OS: $PRETTY_NAME (requires Ubuntu or Debian)"; exit 1
    fi
    log "OS: $PRETTY_NAME"

    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)  log "Architecture: amd64" ;;
        aarch64) log "Architecture: arm64" ;;
        *) err "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    if ! sudo -n true 2>/dev/null; then
        err "This script requires passwordless sudo (or run as root)"; exit 1
    fi
    log "Sudo: OK"

    if ! curl -fsSL --connect-timeout 5 --max-time 10 https://pkgs.k8s.io >/dev/null 2>&1; then
        err "Cannot reach https://pkgs.k8s.io — check internet connectivity"; exit 1
    fi
    log "Internet: OK"
}

# In multi-homed VMs, kubelet may advertise the wrong IP to the control plane.
# We pick the first IP reported by hostname -I, which matches the primary interface.
detect_node_ip() {
    if [[ -z "$NODE_IP" ]]; then
        NODE_IP=$(hostname -I | awk '{print $1}')
        log "Auto-detected node IP: $NODE_IP"
    else
        log "Using specified node IP: $NODE_IP"
    fi
    if [[ -z "$NODE_IP" ]]; then
        err "Could not determine node IP. Use --node-ip or set NODE_IP env var."; exit 1
    fi
}

log "Starting Kubernetes environment setup"
log "  Version : $KUBE_VERSION"
log "  Minor   : $KUBE_MINOR"

preflight
detect_node_ip

log "Installing containerd"
run sudo apt-get update -q
run sudo apt-get install -y containerd

run sudo mkdir -p /etc/containerd

if [[ -f /etc/containerd/config.toml ]]; then
    log "Backing up existing containerd config"
    run sudo cp /etc/containerd/config.toml /etc/containerd/config.toml.bak
fi

# SystemdCgroup=true makes containerd use the same cgroup driver as kubelet.
# Mismatched drivers cause kubelet to fail with cgroup errors at node join time.
# pause:3.10 is the sandbox image that holds Linux namespaces for each pod.
log "Generating containerd config"
if ! $DRY_RUN; then
    containerd config default \
        | sed 's/SystemdCgroup = false/SystemdCgroup = true/' \
        | sed 's|sandbox_image = ".*"|sandbox_image = "registry.k8s.io/pause:3.10"|' \
        | sudo tee /etc/containerd/config.toml >/dev/null
else
    echo "[DRY-RUN] containerd config default | sed ... | sudo tee /etc/containerd/config.toml"
fi

run sudo systemctl enable --now containerd

if command -v crictl &>/dev/null; then
    log "Configuring crictl to use containerd socket"
    run sudo crictl config --set runtime-endpoint=unix:///run/containerd/containerd.sock
else
    warn "crictl not found — configure it after kubeadm install completes"
fi

# swapoff -a disables swap for the current session only.
# The fstab edit prevents it from re-enabling after reboot.
# Kubernetes rejects nodes with active swap unless explicitly configured for cgroup v2.
log "Disabling swap (current session + persistent via fstab)"
run sudo swapoff -a
run sudo sed -i.bak '/\bswap\b/s/^/#/' /etc/fstab
log "Swap entries commented out in /etc/fstab (backup: /etc/fstab.bak)"

log "Loading kernel modules for networking"
run sudo modprobe overlay
run sudo modprobe br_netfilter

# grep before append prevents duplicate entries when the script is re-run.
MODULES_FILE=/etc/modules-load.d/k8s.conf
for mod in overlay br_netfilter; do
    if ! grep -qxF "$mod" "$MODULES_FILE" 2>/dev/null; then
        if ! $DRY_RUN; then
            echo "$mod" | sudo tee -a "$MODULES_FILE" >/dev/null
        else
            echo "[DRY-RUN] echo '$mod' | sudo tee -a $MODULES_FILE"
        fi
        log "  Added: $mod"
    else
        log "  Already present: $mod (skip)"
    fi
done

log "Configuring sysctl parameters"
run sudo bash -c 'cat > /etc/sysctl.d/99-k8s.conf <<EOF
net.ipv4.ip_forward = 1
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.conf.all.forwarding = 1
EOF'

# bridge-nf sysctl parameters only take effect after br_netfilter is loaded.
# sysctl --system silently skips unknown keys, so without the module the
# bridge firewall rules would appear to apply but actually do nothing.
if ! lsmod | grep -q "^br_netfilter"; then
    warn "br_netfilter not yet in lsmod — loading before sysctl"
    if ! $DRY_RUN; then sudo modprobe br_netfilter; fi
fi

run sudo sysctl --system

log "Installing apt transport dependencies"
run sudo apt-get install -y apt-transport-https ca-certificates curl gpg

run sudo mkdir -p -m 755 /etc/apt/keyrings

GPG_KEYRING=/etc/apt/keyrings/kubernetes-apt-keyring.gpg
K8S_REPO_URL="https://pkgs.k8s.io/core:/stable:/v${KUBE_MINOR}/deb"

# Skip re-downloading the key if it's already there so the script is idempotent.
if [[ ! -f "$GPG_KEYRING" ]]; then
    log "Downloading Kubernetes GPG key for v${KUBE_MINOR}"
    if ! $DRY_RUN; then
        sudo curl -fsSL "${K8S_REPO_URL}/Release.key" | sudo gpg --dearmor -o "$GPG_KEYRING"
    else
        echo "[DRY-RUN] Download and dearmor GPG key to $GPG_KEYRING"
    fi
else
    log "GPG keyring already exists: $GPG_KEYRING (skip download)"
fi

log "Writing Kubernetes apt source for v${KUBE_MINOR}"
if ! $DRY_RUN; then
    echo "deb [signed-by=${GPG_KEYRING}] ${K8S_REPO_URL}/ /" \
        | sudo tee /etc/apt/sources.list.d/kubernetes.list >/dev/null
else
    echo "[DRY-RUN] Write kubernetes.list for v${KUBE_MINOR}"
fi

run sudo apt-get update -q

log "Available Kubernetes versions (top 5):"
apt-cache madison kubelet 2>/dev/null | head -5 || true

log "Installing kubelet, kubeadm, kubectl @ $KUBE_VERSION"
run sudo apt-get install -y \
    kubelet="$KUBE_VERSION" \
    kubeadm="$KUBE_VERSION" \
    kubectl="$KUBE_VERSION"

# apt-mark hold prevents unattended-upgrades from bumping the version mid-cluster.
log "Holding Kubernetes packages at $KUBE_VERSION"
run sudo apt-mark hold kubelet kubeadm kubectl

run sudo systemctl enable kubelet

# --node-ip tells kubelet which interface to advertise to the control plane.
# Critical on multi-homed nodes where the default route may not be the cluster interface.
log "Configuring kubelet node IP: $NODE_IP"
if ! $DRY_RUN; then
    echo "KUBELET_EXTRA_ARGS=\"--node-ip=${NODE_IP}\"" | sudo tee /etc/default/kubelet >/dev/null
else
    echo "[DRY-RUN] Write KUBELET_EXTRA_ARGS=--node-ip=${NODE_IP} to /etc/default/kubelet"
fi
run sudo systemctl daemon-reload
run sudo systemctl restart kubelet

log "Installation Verification"
printf "  %-24s %s\n" "containerd:"          "$(containerd --version 2>/dev/null || echo 'not found')"
printf "  %-24s %s\n" "kubelet expected:"    "$KUBE_VERSION"
printf "  %-24s %s\n" "kubelet installed:"   "$(kubelet --version 2>/dev/null || echo 'not found')"
printf "  %-24s %s\n" "kubeadm:"             "$(kubeadm version -o short 2>/dev/null || echo 'not found')"
printf "  %-24s %s\n" "kubectl:"             "$(kubectl version --client 2>/dev/null | head -1 || echo 'not found')"

echo ""
log "Service Status"
printf "  %-14s %s\n" "containerd:" "$(systemctl is-active containerd 2>/dev/null || echo 'FAILED')"
printf "  %-14s %s\n" "kubelet:"    "$(systemctl is-active kubelet 2>/dev/null || echo 'FAILED')"

echo ""
log "Kernel Modules"
for mod in overlay br_netfilter; do
    if lsmod | grep -q "^${mod}"; then
        printf "  %-16s LOADED\n" "$mod:"
    else
        printf "  %-16s NOT LOADED\n" "$mod:"
    fi
done

echo ""
log "Sysctl Parameters"
sysctl net.ipv4.ip_forward \
       net.bridge.bridge-nf-call-iptables \
       net.bridge.bridge-nf-call-ip6tables 2>/dev/null | sed 's/^/  /' || true

echo ""
log "Kubernetes environment setup completed successfully!"
