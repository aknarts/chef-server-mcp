# chef-server-mcp

A lightweight Go HTTP service that exposes a limited management/control plane in front of a Chef Server. It uses the Chef API where possible (via github.com/go-chef/chef) and transparently falls back to the `knife` CLI when an API path is unavailable or fails (when fallback is enabled).

Repository: https://github.com/aknarts/chef-server-mcp

## Features (current)
- Endpoints:
  - `GET /healthz` – liveness check
  - `GET /version` – build/version info
  - `GET /nodes` – list node names (Chef API first; optional knife fallback)
- Graceful shutdown on SIGINT/SIGTERM
- Environment-based configuration
- Optional knife fallback path

## Roadmap (suggested)
- Additional resources: cookbooks, environments, roles, data bags
- Filtering & pagination
- AuthN/AuthZ layer (token or mTLS)
- Metrics endpoint (Prometheus)
- Structured logging upgrade (zap/slog)
- OpenAPI specification

## Configuration
Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | HTTP listen port | 8080 |
| CHEF_USER | Chef API client/user name | (empty) |
| CHEF_KEY_PATH | Path to client PEM key | (empty) |
| CHEF_SERVER_URL | Base URL of Chef server (e.g. https://chef.example.com/organizations/org) | (empty) |
| KNIFE_FALLBACK | Enable knife fallback (true/false) | true |

Behavior matrix:
- If Chef creds complete and valid → API used; knife only if API call fails.
- If creds missing and KNIFE_FALLBACK=true → knife only.
- If creds missing and KNIFE_FALLBACK=false → startup fails.

## Installation
### 0. Initialize local repo (if you cloned without remote yet)
```bash
git remote add origin git@github.com:aknarts/chef-server-mcp.git
git branch -M main
git push -u origin main
```

### 1. From Source (development clone)
```bash
git clone https://github.com/aknarts/chef-server-mcp.git
cd chef-server-mcp
go build -o chef-mcp ./cmd/chef-mcp
./chef-mcp
```

### 2. Using go install (no repo clone needed)
(Produces a binary in $GOBIN or $GOPATH/bin)
```bash
go install github.com/aknarts/chef-server-mcp/cmd/chef-mcp@latest
chef-mcp
```
Set env vars before running (see Configuration).

### 3. Download Release Binary (after releases are published)
```bash
# Example for Linux amd64
VERSION=v0.1.0
curl -L -o chef-mcp.tar.gz \
  https://github.com/aknarts/chef-server-mcp/releases/download/$VERSION/chef-mcp_${VERSION}_linux_amd64.tar.gz
tar -xzf chef-mcp.tar.gz
./chef-mcp
```
Verify checksum/signature if provided.

### 4. Docker (GitHub Container Registry)
Authenticate (only needed for private repo or rate limits):
```bash
echo "$GHCR_TOKEN" | docker login ghcr.io -u aknarts --password-stdin
```
Pull & run:
```bash
docker pull ghcr.io/aknarts/chef-server-mcp:latest

docker run --rm -p 8080:8080 \
  -e CHEF_USER=$CHEF_USER \
  -e CHEF_KEY_PATH=/secrets/client.pem \
  -e CHEF_SERVER_URL=$CHEF_SERVER_URL \
  -e KNIFE_FALLBACK=true \
  -v $HOME/.chef/myuser.pem:/secrets/client.pem:ro \
  ghcr.io/aknarts/chef-server-mcp:latest
```
If knife fallback is required inside the container, extend the image to add `knife` (Chef Workstation) or mount it from a shared volume.

### 5. Minimal Derived Dockerfile (adding knife)
```Dockerfile
FROM ghcr.io/aknarts/chef-server-mcp:latest AS base
# Example layer (adjust for real Chef Workstation installation steps)
# RUN curl -L https://omnitruck.chef.io/install.sh | bash -s -- -P chef-workstation
# Ensure knife in PATH and configure /etc/chef/knife.rb or mount at runtime.
```

## Editor / IDE Integration
### VSCode
#### Debug (launch.json)
Create .vscode/launch.json:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Run chef-mcp",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/chef-mcp",
      "env": {
        "CHEF_USER": "myuser",
        "CHEF_KEY_PATH": "${env:HOME}/.chef/myuser.pem",
        "CHEF_SERVER_URL": "https://chef.example.com/organizations/acme",
        "KNIFE_FALLBACK": "true"
      }
    }
  ]
}
```
#### Task (optional build)
.vscode/tasks.json:
```json
{
  "version": "2.0.0",
  "tasks": [
    { "label": "build", "type": "shell", "command": "go build -o bin/chef-mcp ./cmd/chef-mcp" }
  ]
}
```
#### Dev Container (optional)
.devcontainer/devcontainer.json (snippet):
```json
{
  "name": "chef-mcp-dev",
  "image": "mcr.microsoft.com/devcontainers/go:1-1.22-bullseye",
  "features": {},
  "postCreateCommand": "go mod download",
  "forwardPorts": [8080],
  "containerEnv": {
    "KNIFE_FALLBACK": "true"
  },
  "mounts": [
    "source=${env:HOME}/.chef,target=/workspaces/.chef,type=bind,consistency=cached"
  ]
}
```
Add CHEF_* env vars in a local override or use a secret mount.

### JetBrains (GoLand / IntelliJ with Go plugin)
1. Open the project root (go.mod recognized automatically).
2. New Run Configuration → Go Build:
   - Name: chef-mcp
   - Package path: ./cmd/chef-mcp
   - Working dir: project root
   - Environment:
     - CHEF_USER=myuser
     - CHEF_KEY_PATH=/home/you/.chef/myuser.pem
     - CHEF_SERVER_URL=https://chef.example.com/organizations/acme
     - KNIFE_FALLBACK=true
3. Run or Debug (breakpoints in internal/server or chefapi).
4. For Docker:
   - Add a Docker run configuration pointing at image ghcr.io/aknarts/chef-server-mcp:latest
   - Map port 8080
   - Add environment variables under Container settings
   - Mount /home/you/.chef as read-only if using knife fallback.

## Local Development
```bash
# 1. Clone & enter
git clone https://github.com/aknarts/chef-server-mcp.git && cd chef-server-mcp

# 2. (Optional) export configuration
export CHEF_USER=myuser
export CHEF_KEY_PATH=$HOME/.chef/myuser.pem
export CHEF_SERVER_URL=https://chef.example.com/organizations/acme
export KNIFE_FALLBACK=true

# 3. Run
go run ./cmd/chef-mcp

# 4. Test endpoints
curl -s localhost:8080/healthz
curl -s localhost:8080/version
curl -s localhost:8080/nodes | jq
```

### Hot Reload (optional)
```bash
go install github.com/air-verse/air@latest
air
```

### Tests
```bash
go test ./...
```

## Docker
Build multi-stage image:
```bash
docker build -t chef-mcp:dev .

docker run --rm -p 8080:8080 \
  -e CHEF_USER=$CHEF_USER \
  -e CHEF_KEY_PATH=/secrets/client.pem \
  -e CHEF_SERVER_URL=$CHEF_SERVER_URL \
  -e KNIFE_FALLBACK=true \
  -v $HOME/.chef/myuser.pem:/secrets/client.pem:ro \
  chef-mcp:dev
```
If using knife fallback inside container, ensure `knife` binary & config are in the image or mounted. (Current Dockerfile does not yet install knife—add workstation layer or mount host binary if needed.)

## Version Injection
At build time:
```bash
go build -ldflags "-X github.com/aknarts/chef-server-mcp/internal/version.Version=$(git describe --tags --always --dirty)" ./cmd/chef-mcp
```

## GitHub Actions
A sample workflow is included in `.github/workflows/ci.yml` for build + test + (placeholder) docker build. Tag-based releases can extend this.

## Deployment Options
- Systemd service on VM
- Kubernetes Deployment + Service + Ingress (mount key secret)
- Container on ECS/Fargate/Nomad

Ensure secret management (Chef key) uses a secrets manager or sealed secret. If knife fallback enabled, Chef workstation or minimal runtime with `knife` must be present.

## Security Notes
- Run as non-root in production containers
- Limit network egress to Chef server if possible
- Consider removing knife fallback in high security contexts
- Add auth (header token / mTLS) before exposing externally

## Contributing
1. Fork
2. Branch
3. Make changes + tests
4. PR with concise description

## License
(Choose and add a LICENSE file – MIT, Apache-2.0, etc.)

## Disclaimer
This is an early-stage tool; validate outputs in non-production first.
