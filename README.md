# chef-server-mcp (MCP only)

The project now ships only an MCP (Model Context Protocol) stdio tool binary: `mcp-chef`.

Repository: https://github.com/aknarts/chef-server-mcp

## Binary
- `mcp-chef`: MCP stdio server exposing Chef introspection tools (API only, knife fallback removed).

### MCP Server (`mcp-chef`)
Build:
```bash
go build -o mcp-chef ./cmd/mcp-chef
```
Run (environment variables):
```bash
CHEF_USER=myuser \
CHEF_KEY_PATH=$HOME/.chef/myuser.pem \
CHEF_SERVER_URL=https://chef.example.com/organizations/acme \
./mcp-chef
```
A host (IDE/agent) should speak MCP over stdin/stdout.

## Tools
Current MCP tools (input -> output summary):

| Tool | Input Fields | Output Fields | Description |
|------|--------------|---------------|-------------|
| listNodes | (none) | nodes[] | List node names. |
| getNode | name | node | Fetch a single node object by name. |
| listRoles | (none) | roles[] | List role names. |
| getRole | name | role | Fetch a single role definition. |
| listUsers | (none) | users[] | List user names. |
| getUser | name | user | Fetch a single user object. |
| search | index, query | total, start, rows[] | Execute a Chef search returning decoded rows (generic objects). |
| searchJSON | index, query | total, start, rows[] (raw JSON) | Execute a Chef search returning raw JSON rows. |

Notes:
- All tools require valid Chef API credentials.
- Inline PEM key contents are supported via CHEF_KEY_PATH env var.

## Features
- Multiple MCP tools for nodes, roles, users, search
- API-only (knife fallback removed)
- Inline PEM key support
- Graceful shutdown on SIGINT/SIGTERM

## Configuration
Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| CHEF_USER | Chef API client/user name | (empty) |
| CHEF_KEY_PATH | Path to client PEM key OR the inline PEM contents themselves | (empty) |
| CHEF_SERVER_URL | Chef server URL (e.g. https://chef.example.com/organizations/org) | (empty) |

Behavior:
- All variables required; process exits if any are missing.

Inline key usage example:
```bash
export CHEF_USER=myuser
export CHEF_SERVER_URL=https://chef.example.com/organizations/acme
export CHEF_KEY_PATH="$(cat $HOME/.chef/myuser.pem)"
./mcp-chef
```

## Installation
### From Source
```bash
git clone https://github.com/aknarts/chef-server-mcp.git
cd chef-server-mcp
go build -o mcp-chef ./cmd/mcp-chef
./mcp-chef
```

### go install
```bash
go install github.com/aknarts/chef-server-mcp/cmd/mcp-chef@latest
mcp-chef
```

### Docker
```bash
docker build -t mcp-chef:latest .
```
(Use a host process to interact over stdio.)

## Development
```bash
make build-mcp
make mcp-smoke
```

## Version Injection
```bash
go build -ldflags "-X github.com/aknarts/chef-server-mcp/internal/version.Version=$(git describe --tags --always --dirty)" ./cmd/mcp-chef
```

## Roadmap
- Additional resources (cookbooks, environments, data bags) as tools
- Filtering & pagination helpers
- AuthN/AuthZ (tokens / mTLS) for any future network exposure
- Structured logging
- Metrics (latency, error counts)

## Security Notes
- Never commit private keys
- Validate / sanitize any user-supplied inputs (queries)

## License
See repository for licensing details.
