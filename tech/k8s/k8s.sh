set -e

# Error handling and cleanup
trap 'echo "[ERROR] Script failed at line $LINENO"; exit 1' ERR
trap 'echo "[INFO] Cleanup on exit"; exit 0' EXIT

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

log "Starting Kubernetes environment setup"

# update packages in apt package manager
log "Updating package manager"
sudo apt update

# install containerd using the apt package manager
# containerd is lightweight, reliable and fast (CRI native)
log "Installing containerd"
sudo apt-get install -y containerd

# create /etc/containerd directory for containerd configuration
sudo mkdir -p /etc/containerd

# Backup existing config if it exists
if [ -f /etc/containerd/config.toml ]; then
    log "Backing up existing containerd config"
    sudo cp /etc/containerd/config.toml /etc/containerd/config.toml.bak
fi

# Generate the default containerd configuration
# Change the pause container to version 3.10 (pause container holds the linux ns for Kubernetes namespaces)
# Set `SystemdCgroup` to true to use same cgroup drive as kubelet
containerd config default |
  sed 's/SystemdCgroup = false/SystemdCgroup = true/' |
  sed 's|sandbox_image = ".*"|sandbox_image = "registry.k8s.io/pause:3.10"|' |
  sudo tee /etc/containerd/config.toml >/dev/null

# Restart containerd to apply the configuration changes
log "Restarting containerd"
sudo systemctl restart containerd
sudo systemctl enable containerd

# Kubernetes doesn't support swap unless explicitly configured under cgroup v2
log "Disabling swap"
sudo swapoff -a

# Load required kernel modules for Kubernetes networking
log "Loading kernel modules for networking"
sudo modprobe overlay
sudo modprobe br_netfilter

# Persist kernel modules across reboots
sudo bash -c 'echo "overlay" >> /etc/modules-load.d/k8s.conf'
sudo bash -c 'echo "br_netfilter" >> /etc/modules-load.d/k8s.conf'

# update packages
log "Updating package manager (again)"
sudo apt update

# install apt-transport-https ca-certificates curl and gpg packages using
# apt package manager in order to fetch Kubernetes packages from
# external HTTPS repositories
log "Installing apt dependencies"
sudo apt-get install -y apt-transport-https ca-certificates curl gpg

# create a secure directory for storing GPG keyring files
# used by APT to verify trusted repositories.
# This is part of a newer, more secure APT repository layout that
# keeps trusted keys isolated from system-wide GPG configurations
sudo mkdir -p -m 755 /etc/apt/keyrings

# download the k8s release gpg key FOR 1.33
log "Downloading Kubernetes GPG key"
sudo curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.33/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# Download and convert the Kubernetes APT repository's GPG public key into
# a binary format (`.gpg`) that APT can use to verify the integrity
# and authenticity of Kubernetes packages during installation.
# This overwrites any existing configuration in
# /etc/apt/sources.list.d/kubernetes.list FOR 1.33
# (`tee` without `-a` (append) will **replace** the contents of the file)
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.33/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

# update packages in apt
log "Updating package manager with Kubernetes repository"
sudo apt-get update

log "Checking available Kubernetes versions"
apt-cache madison kubelet
apt-cache madison kubectl
apt-cache madison kubeadm

KUBE_VERSION="1.33.2-1.1"

# install kubelet, kubeadm, and kubectl at version 1.33.2-1.1
log "Installing Kubernetes tools (kubelet, kubeadm, kubectl) version $KUBE_VERSION"
sudo apt-get install -y kubelet=$KUBE_VERSION kubeadm=$KUBE_VERSION kubectl=$KUBE_VERSION

# hold these packages at version
log "Holding Kubernetes packages at version $KUBE_VERSION"
sudo apt-mark hold kubelet kubeadm kubectl

# Enable kubelet service
log "Enabling kubelet service"
sudo systemctl enable kubelet

# Configure sysctl parameters for Kubernetes networking
log "Configuring sysctl parameters for Kubernetes"
sudo bash -c 'cat > /etc/sysctl.d/99-k8s.conf <<EOF
# Enable IP packet forwarding
net.ipv4.ip_forward = 1

# Enable netfilter bridge calls for iptables
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1

# Allow bridge to process IPv4 traffic
net.ipv4.conf.all.forwarding = 1
EOF'

# Apply sysctl settings
sudo sysctl --system

# Setting node IP address for kubelet, this very important in multi-homed systems or virtual machines
log "Configuring kubelet with node IP address"
echo 'KUBELET_EXTRA_ARGS="--node-ip=192.168.1.122"' | sudo tee /etc/default/kubelet
sudo systemctl daemon-reload
sudo systemctl restart kubelet

# Verify installation
log "Verifying Kubernetes installation"
echo ""

log "Installation Verification"

echo -n "containerd version: "
containerd --version 2>/dev/null || echo "Not available"

echo "Expected kubelet version: $KUBE_VERSION"
echo -n "Actual kubelet version: "
kubelet --version 2>/dev/null || echo "Not installed"

KUBEADM_VER=$(kubeadm version -o short 2>/dev/null || echo "Not available")
echo "kubeadm version: $KUBEADM_VER"

KUBECTL_VER=$(kubectl version --client --short 2>/dev/null || echo "Not available")
echo "kubectl version: $KUBECTL_VER"

log "Service Status"
echo -n "containerd service: "
sudo systemctl is-active containerd 2>/dev/null || echo "FAILED"

echo -n "kubelet service: "
sudo systemctl is-active kubelet 2>/dev/null || echo "FAILED"

log "Kernel Modules"
echo -n "overlay module: "
ls /sys/module/overlay >/dev/null 2>&1 && echo "LOADED" || echo "NOT LOADED"

echo -n "br_netfilter module: "
ls /sys/module/br_netfilter >/dev/null 2>&1 && echo "LOADED" || echo "NOT LOADED"

echo ""
log "Kubernetes environment setup completed successfully!"
