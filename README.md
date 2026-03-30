# Deploy Platform

A self-hosted mini Heroku / Vercel clone. Push a zipped project → auto-builds a Docker container → deploys it to Kubernetes (k3s) → assigns it a subdomain URL.

## Stack

| Layer | Tech |
|---|---|
| API | Go 1.22 (zero external dependencies, pure stdlib) |
| Frontend | React 18 + Vite |
| Orchestration | k3s (lightweight Kubernetes) |
| Proxy | Nginx |
| Registry | Docker Registry v2 |
| Dev env | Docker Compose |

## Quick start (WSL2 / Linux)

```bash
# 1. Install dependencies (Docker, Go, k3s)
chmod +x scripts/install.sh && ./scripts/install.sh

# 2. Reload shell
newgrp docker && source ~/.bashrc

# 3. Start the platform
chmod +x scripts/start.sh && ./scripts/start.sh

# 4. Open dashboard
#    http://localhost:5173

# 5. Deploy the included sample app
chmod +x scripts/deploy-cli.sh
./scripts/deploy-cli.sh hello-world ./sample-app
```

## How it works

```
CLI / Dashboard
      │  ZIP + POST /deploy
      ▼
  Go API (:8080)
      │  queues job
      ▼
  Build Worker
      │  docker build → docker push → localhost:5000
      ▼
  k8s Deployer
      │  Deployment + Service + Ingress
      ▼
  k3s cluster
      │
      ▼
  http://your-app.localhost
```

## Project structure

```
deploy-platform/
├── api/                   Go REST API (zero external deps)
│   ├── main.go
│   ├── builder/           docker build + push + job queue
│   ├── handlers/          HTTP handlers (deploy, status, list)
│   ├── k8s/               Kubernetes REST client (pure stdlib)
│   └── models/            DeployJob struct
├── frontend/              React dashboard (Vite)
│   └── src/
│       └── components/    DeployForm, DeployCard, LogViewer, StatusBadge
├── nginx/                 Reverse proxy config
├── k8s-manifests/         Namespace, Registry, API deployment YAMLs
├── scripts/               install.sh, start.sh, deploy-cli.sh
├── sample-app/            Test app (Go HTTP server + Dockerfile)
└── docker-compose.yml     Local dev stack
```

## Deploying your own apps

Your project needs:
1. A `Dockerfile` at the root
2. The container must listen on port `8080`

```bash
./scripts/deploy-cli.sh <app-name> ./path/to/project
```

### Example Dockerfiles

**Node.js**
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install --production
COPY . .
EXPOSE 8080
CMD ["node", "index.js"]
```

**Python**
```dockerfile
FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 8080
CMD ["python", "app.py"]
```

## API reference

| Method | Path | Description |
|---|---|---|
| `POST` | `/deploy` | Submit a zip for deployment |
| `GET` | `/deploy` | List all deployments |
| `GET` | `/deploy/:id` | Get status + logs for a deployment |
| `GET` | `/health` | Health check |

## Troubleshooting

**Docker permission denied**
```bash
newgrp docker
```

**k3s not starting in WSL2**
Add to `/etc/wsl.conf`:
```ini
[boot]
command="mount --make-rshared /"
```
Then restart: `wsl --shutdown` in PowerShell.

**Registry push fails**
Check `/etc/docker/daemon.json` contains:
```json
{"insecure-registries": ["localhost:5000"]}
```
Then: `sudo service docker restart`

## License

MIT
