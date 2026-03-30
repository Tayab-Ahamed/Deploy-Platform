#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
#  Deploy Platform — WSL2 Installer
#  Installs: Docker, Go 1.22, k3s
#  Safe to re-run — each step checks if already installed.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
info()    { echo -e "${CYAN}[info]${NC}  $*"; }
success() { echo -e "${GREEN}[ ok ]${NC}  $*"; }
warn()    { echo -e "${YELLOW}[warn]${NC}  $*"; }
die()     { echo -e "${RED}[err ]${NC}  $*" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo ""
echo "  Deploy Platform — WSL2 Installer"
echo "  ─────────────────────────────────"
echo ""

# ── 1. Docker ─────────────────────────────────────────────────────────
if command -v docker &>/dev/null && docker info &>/dev/null 2>&1; then
  success "Docker already running: $(docker --version | cut -d' ' -f3 | tr -d ',')"
else
  info "Installing Docker..."
  sudo apt-get update -qq
  sudo apt-get install -y -qq ca-certificates curl gnupg lsb-release

  sudo install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
    | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null
  sudo chmod a+r /etc/apt/keyrings/docker.gpg

  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
    https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
    | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

  sudo apt-get update -qq
  sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin

  sudo usermod -aG docker "$USER"
  success "Docker installed — you may need to run: newgrp docker"
fi

# ── 1b. Allow insecure local registry ────────────────────────────────
DAEMON_JSON=/etc/docker/daemon.json
if ! sudo grep -q '"insecure-registries"' "$DAEMON_JSON" 2>/dev/null; then
  info "Allowing insecure local registry at localhost:5000..."
  if sudo test -f "$DAEMON_JSON"; then
    # Merge into existing file using Python (always available on Ubuntu)
    sudo python3 -c "
import json
with open('$DAEMON_JSON') as f: d = json.load(f)
d.setdefault('insecure-registries', [])
if 'localhost:5000' not in d['insecure-registries']:
    d['insecure-registries'].append('localhost:5000')
with open('$DAEMON_JSON', 'w') as f: json.dump(d, f, indent=2)
print('merged')
"
  else
    echo '{"insecure-registries":["localhost:5000"]}' | sudo tee "$DAEMON_JSON" >/dev/null
  fi
  sudo service docker restart || sudo systemctl restart docker || true
  success "Registry config written to $DAEMON_JSON"
fi

# ── 2. Go ─────────────────────────────────────────────────────────────
GO_VERSION="1.22.5"
INSTALLED_GO="$(go version 2>/dev/null | grep -oP 'go\K[0-9]+\.[0-9]+' || true)"

if [[ "$INSTALLED_GO" == "1.22" ]]; then
  success "Go already installed: $(go version)"
else
  info "Installing Go $GO_VERSION..."
  TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
  curl -fsSL "https://go.dev/dl/${TARBALL}" -o "/tmp/${TARBALL}"
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "/tmp/${TARBALL}"
  rm "/tmp/${TARBALL}"

  for RC in "$HOME/.bashrc" "$HOME/.profile"; do
    grep -q '/usr/local/go/bin' "$RC" 2>/dev/null || \
      echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> "$RC"
  done
  export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
  success "Go $GO_VERSION installed"
fi

# ── 3. k3s ────────────────────────────────────────────────────────────
if command -v k3s &>/dev/null; then
  success "k3s already installed: $(k3s --version 2>/dev/null | head -1)"
else
  info "Installing k3s (lightweight Kubernetes)..."

  # WSL2 cgroup fix
  if grep -qi microsoft /proc/version 2>/dev/null; then
    if ! grep -q "SystemdCgroup" /etc/containerd/config.toml 2>/dev/null; then
      warn "Detected WSL2. Applying cgroup fix..."
    fi
  fi

  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--write-kubeconfig-mode=644" sh -
  sleep 8
  success "k3s installed"
fi

# ── 3b. Kubeconfig access ─────────────────────────────────────────────
if [ -f /etc/rancher/k3s/k3s.yaml ]; then
  sudo chmod 644 /etc/rancher/k3s/k3s.yaml
  for RC in "$HOME/.bashrc" "$HOME/.profile"; do
    grep -q 'KUBECONFIG' "$RC" 2>/dev/null || \
      echo 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' >> "$RC"
  done
  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
  success "KUBECONFIG set to /etc/rancher/k3s/k3s.yaml"
fi

# ── 3c. kubectl alias ─────────────────────────────────────────────────
if ! command -v kubectl &>/dev/null; then
  sudo ln -sf /usr/local/bin/k3s /usr/local/bin/kubectl
  success "kubectl linked → k3s"
fi

# ── 4. Apply k8s manifests ────────────────────────────────────────────
info "Waiting for k3s API to be ready..."
TIMEOUT=60
ELAPSED=0
until kubectl get nodes &>/dev/null 2>&1 || [ $ELAPSED -ge $TIMEOUT ]; do
  sleep 3; ELAPSED=$((ELAPSED+3))
done

if kubectl get nodes &>/dev/null 2>&1; then
  kubectl apply -f "$ROOT/k8s-manifests/namespace.yaml"
  kubectl apply -f "$ROOT/k8s-manifests/registry.yaml"
  success "Kubernetes manifests applied"
else
  warn "k3s API not ready yet — run 'kubectl apply -f k8s-manifests/' manually after it starts"
fi

echo ""
echo -e "${GREEN}  ✓ Installation complete!${NC}"
echo ""
echo "  Next steps:"
echo "    newgrp docker          # pick up Docker group (or re-open WSL2)"
echo "    source ~/.bashrc       # pick up Go + KUBECONFIG in PATH"
echo "    ./scripts/start.sh     # start the platform"
echo ""
