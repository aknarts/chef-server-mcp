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
CHEF_SERVER_URL=https://chef.example.com \
CHEF_DEFAULT_ORG=acme \
./mcp-chef
```
A host (IDE/agent) should speak MCP over stdin/stdout.

## Tools
Current MCP tools (input -> output summary):

| Tool | Input Fields | Output Fields | Description |
|------|--------------|---------------|-------------|
| listNodes | organization (optional) | nodes[], organization | List node names from specified or default organization. |
| getNode | name, organization (optional) | node, organization | Fetch a single node object by name from specified or default organization. |
| listRoles | organization (optional) | roles[], organization | List role names from specified or default organization. |
| getRole | name, organization (optional) | role, organization | Fetch a single role definition from specified or default organization. |
| listUsers | organization (optional) | users[], organization | List user names from specified or default organization. |
| getUser | name, organization (optional) | user, organization | Fetch a single user object from specified or default organization. |
| search | index, query, organization (optional) | total, start, rows[], organization | Execute a Chef search returning decoded rows from specified or default organization. |
| searchJSON | index, query, organization (optional) | total, start, rows[] (raw JSON), organization | Execute a Chef search returning raw JSON rows from specified or default organization. |
| getOrganization | organization (optional) | organization, orgName | Get organization details from specified or default organization. |
| listCookbooks | organization (optional) | cookbooks, organization | List cookbook names and versions from specified or default organization. |
| getCookbook | name, version (optional, defaults to _latest), organization (optional) | cookbook, organization | Get a specific cookbook version from specified or default organization. |
| listDataBags | organization (optional) | dataBags[], organization | List data bag names from specified or default organization. |
| listDataBagItems | name, organization (optional) | items[], dataBag, organization | List items in a specific data bag from specified or default organization. |
| getDataBagItem | bagName, itemName, organization (optional) | item, dataBag, organization | Get a specific data bag item from specified or default organization. |
| listEnvironments | organization (optional) | environments[], organization | List environment names from specified or default organization. |
| getEnvironment | name, organization (optional) | environment, organization | Get a specific environment definition from specified or default organization. |

Notes:
- All tools require valid Chef API credentials.
- Inline PEM key contents are supported via CHEF_KEY_PATH env var.
- The `organization` parameter is optional in all tools. If omitted, `CHEF_DEFAULT_ORG` is used.
- Organization aliases are resolved automatically (e.g., "qa" → "qa1", "prod" → "fireamp_classic").

## Features
- Multiple MCP tools for nodes, roles, users, search, cookbooks, data bags, environments, and organizations
- Multi-organization support with aliases
- Cookbook version management with `_latest` default support
- Complete data bag management (list bags, list items, get specific items)
- Environment configuration access
- API-only (knife fallback removed)
- Inline PEM key support
- Graceful shutdown on SIGINT/SIGTERM

## Configuration
Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| CHEF_USER | Chef API client/user name | (empty) |
| CHEF_KEY_PATH | Path to client PEM key OR the inline PEM contents themselves | (empty) |
| CHEF_SERVER_URL | Base Chef server URL (e.g. https://chef.example.com) | (empty) |
| CHEF_DEFAULT_ORG | Default organization to use when none specified in requests | (empty) |
| CHEF_ORG_ALIASES | Organization aliases in JSON format or key=value pairs | (empty) |

Behavior:
- CHEF_USER, CHEF_KEY_PATH, and CHEF_SERVER_URL are required; process exits if any are missing.
- CHEF_DEFAULT_ORG is optional but recommended for easier usage.
- CHEF_ORG_ALIASES allows you to define shortcuts for organization names.

### Usage Examples

Basic usage with default organization:
```bash
export CHEF_USER=myuser
export CHEF_SERVER_URL=https://chef.example.com
export CHEF_DEFAULT_ORG=acme
export CHEF_KEY_PATH="$(cat $HOME/.chef/myuser.pem)"
./mcp-chef
```

With organization aliases:
```bash
export CHEF_USER=myuser
export CHEF_SERVER_URL=https://chef.example.com
export CHEF_DEFAULT_ORG=qa1
export CHEF_ORG_ALIASES='{"qa":"qa1","prod":"fireamp_classic"}'
export CHEF_KEY_PATH="$(cat $HOME/.chef/myuser.pem)"
./mcp-chef
```

Alternative alias format:
```bash
export CHEF_ORG_ALIASES="qa=qa1,prod=fireamp_classic"
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
