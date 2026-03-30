#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
#  Deploy Platform — Start Script (WSL2)
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

GREEN='\033[0;32m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()    { echo -e "${CYAN}[info]${NC}  $*"; }
success() { echo -e "${GREEN}[ ok ]${NC}  $*"; }
warn()    { echo -e "${YELLOW}[warn]${NC}  $*"; }
die()     { echo -e "${RED}[err ]${NC}  $*" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT"

export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

echo ""
echo "  Deploy Platform — Starting"
echo "  ──────────────────────────"
echo ""

# ── Preflight ─────────────────────────────────────────────────────────
for cmd in docker node; do
  command -v "$cmd" &>/dev/null || die "'$cmd' not found. Run: ./scripts/install.sh"
done

# ── Docker daemon ─────────────────────────────────────────────────────
if ! docker info &>/dev/null 2>&1; then
  info "Starting Docker daemon..."
  sudo service docker start || sudo systemctl start docker
  sleep 4
  docker info &>/dev/null 2>&1 || die "Docker failed to start. Try manually: sudo service docker start"
fi
success "Docker running"

# ── k3s ───────────────────────────────────────────────────────────────
K3S_RUNNING=false
if pgrep -x k3s &>/dev/null; then
  K3S_RUNNING=true
elif sudo systemctl is-active --quiet k3s 2>/dev/null; then
  K3S_RUNNING=true
fi

if [ "$K3S_RUNNING" = false ]; then
  info "Starting k3s..."
  sudo systemctl start k3s 2>/dev/null || \
    sudo nohup k3s server --write-kubeconfig-mode=644 >/tmp/k3s.log 2>&1 &
  sleep 10
fi

# Wait for k3s API server
info "Waiting for Kubernetes API (up to 60s)..."
ELAPSED=0
until kubectl get nodes &>/dev/null 2>&1 || [ $ELAPSED -ge 60 ]; do
  sleep 3; ELAPSED=$((ELAPSED+3)); printf "."
done
echo ""

if kubectl get nodes &>/dev/null 2>&1; then
  success "k3s ready ($(kubectl get nodes --no-headers 2>/dev/null | wc -l) node(s))"
  # Apply manifests idempotently
  kubectl apply -f "$ROOT/k8s-manifests/namespace.yaml" --output=name 2>/dev/null || true
  kubectl apply -f "$ROOT/k8s-manifests/registry.yaml"  --output=name 2>/dev/null || true
  success "k8s manifests applied"
else
  warn "k3s API not ready yet — Kubernetes deploy step will fail until k3s is up"
  warn "Check: sudo journalctl -u k3s -n 20"
fi

# ── Frontend deps ─────────────────────────────────────────────────────
if [ ! -d "$ROOT/frontend/node_modules" ]; then
  info "Installing frontend npm packages (first run only)..."
  (cd "$ROOT/frontend" && npm install)
  success "npm install done"
fi

# ── Docker Compose (registry + API + nginx) ───────────────────────────
info "Starting backend services (registry, API, nginx)..."
docker compose up --build -d
success "Backend services started"

# ── Frontend (native Vite — hot reload) ──────────────────────────────
FRONTEND_LOG="/tmp/dp-frontend.log"
info "Starting React frontend on :5173..."
(cd "$ROOT/frontend" && npm run dev > "$FRONTEND_LOG" 2>&1) &
FRONTEND_PID=$!
echo $FRONTEND_PID > /tmp/dp-frontend.pid

# Wait for Vite to be ready
ELAPSED=0
until curl -sf http://localhost:5173 &>/dev/null || [ $ELAPSED -ge 30 ]; do
  sleep 2; ELAPSED=$((ELAPSED+2))
done

if curl -sf http://localhost:5173 &>/dev/null; then
  success "Frontend ready at http://localhost:5173"
else
  warn "Frontend not responding yet — check log: tail -f $FRONTEND_LOG"
fi

# ── API health ────────────────────────────────────────────────────────
ELAPSED=0
until curl -sf http://localhost:8080/health &>/dev/null || [ $ELAPSED -ge 60 ]; do
  sleep 2; ELAPSED=$((ELAPSED+2))
done
curl -sf http://localhost:8080/health &>/dev/null && success "API healthy at http://localhost:8080"

echo ""
echo -e "${GREEN}  ╔══════════════════════════════════════════╗${NC}"
echo -e "${GREEN}  ║   Deploy Platform is live!               ║${NC}"
echo -e "${GREEN}  ╚══════════════════════════════════════════╝${NC}"
echo ""
echo "  Dashboard   →  http://localhost:5173"
echo "  API         →  http://localhost:8080/health"
echo "  Registry    →  http://localhost:5000/v2/"
echo ""
echo "  Test a deploy:"
echo "    ./scripts/deploy-cli.sh hello-world ./sample-app"
echo ""
echo "  Useful commands:"
echo "    docker compose logs -f api          # API logs"
echo "    tail -f /tmp/dp-frontend.log        # Frontend logs"
echo "    kubectl get pods -n deployments     # Running apps"
echo "    docker compose down                 # Stop backend"
echo "    kill \$(cat /tmp/dp-frontend.pid)   # Stop frontend"
echo ""
