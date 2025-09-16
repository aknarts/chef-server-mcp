# chef-server-mcp (MCP only)

The project now ships only an MCP (Model Context Protocol) stdio tool binary: `mcp-chef`.

Repository: https://github.com/aknarts/chef-server-mcp

## Binary
- `mcp-chef`: MCP stdio server exposing Chef introspection tools (API only, knife fallback removed, multi-organization support).

### MCP Server (`mcp-chef`)
Build:
```bash
go build -o mcp-chef ./cmd/mcp-chef
```
Run (environment variables example with aliases):
```bash
# Root Chef server URL (no /organizations/<org>)
CHEF_USER=myuser \
CHEF_KEY_PATH=$HOME/.chef/myuser.pem \
CHEF_SERVER_URL=https://chef.example.com \
CHEF_ORG_ALIASES="qa=qa1,prod=fireamp_classic" \
CHEF_DEFAULT_ORG=prod \
CHEF_DEBUG=1 \
./mcp-chef
```
If you do not configure aliases you must supply the real organization name in the `org` field for org‑scoped tools.

## Tools
Current MCP tools (input -> output summary):

| Tool | Input Fields | Output Fields | Scope | Description |
|------|--------------|---------------|-------|-------------|
| listOrgAliases | (none) | aliases{}, default | config | Show configured org alias mapping and default.|
| listOrgClients | (none) | clients{org:baseURL} | debug | Show cached initialized org clients and their base URLs. |
| listNodes | org? | nodes[] | org | List node names within an organization (alias or real org if no mapping configured). |
| getNode | name, org? | node | org | Fetch a single node object. |
| listRoles | org? | roles[] | org | List role names. |
| getRole | name, org? | role | org | Fetch a single role definition. |
| listUsers | (none) | users[] | server | List user names (server-scope). |
| getUser | name | user | server | Fetch a single user object (server-scope). |
| search | index, query, org? | total, start, rows[] | org | Execute a Chef search returning decoded rows. |
| searchJSON | index, query, org? | total, start, rows[] (raw JSON) | org | Execute a Chef search returning raw JSON rows. |

Notes:
- For org‑scoped tools, `org` is optional if `CHEF_DEFAULT_ORG` (alias) is set or only one alias defined.
- If no aliases configured, the provided `org` value is treated as the actual organization name.
- Inline PEM key contents are supported via CHEF_KEY_PATH (supply the PEM text directly).

## Features
- Multi-organization support with alias mapping & default alias.
- Multiple MCP tools for nodes, roles, users, search.
- API-only (knife fallback removed).
- Inline PEM key support.
- Graceful shutdown on SIGINT/SIGTERM.
- Verbose debug logging (enable with CHEF_DEBUG=1/true/yes) including client creation, cache hits, and API call summaries/errors.

## Configuration
Environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| CHEF_USER | Chef API client/user name | yes |
| CHEF_KEY_PATH | Path to client PEM key OR the inline PEM contents themselves | yes |
| CHEF_SERVER_URL | Root Chef server URL (no /organizations segment) | yes |
| CHEF_ORG_ALIASES | Comma-separated alias=org list (e.g. `qa=qa1,prod=fireamp_classic`) | no |
| CHEF_DEFAULT_ORG | Default org alias to use when `org` omitted | no |
| CHEF_DEBUG | Enable verbose debug logs (1 / true / yes) | no |

Behavior:
- All required variables must be set or the process exits.
- If `CHEF_ORG_ALIASES` has exactly one alias and `CHEF_DEFAULT_ORG` not set, that alias becomes the implicit default.
- If aliases not configured you must always provide `org` (actual org name) for org‑scoped tools.

Inline key usage example:
```bash
export CHEF_USER=myuser
export CHEF_SERVER_URL=https://chef.example.com
export CHEF_KEY_PATH="$(cat $HOME/.chef/myuser.pem)"
export CHEF_ORG_ALIASES="prod=acme-prod"
export CHEF_DEFAULT_ORG=prod
export CHEF_DEBUG=1
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
