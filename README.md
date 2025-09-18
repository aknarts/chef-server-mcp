# Chef Server MCP

A Model Context Protocol (MCP) server that provides AI assistants with read-only access to Chef Server data including nodes, roles, cookbooks, data bags, environments, and more.

[![CI](https://github.com/aknarts/chef-server-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/aknarts/chef-server-mcp/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Read-only Chef Server access** via Chef API
- **Multi-organization support** with aliases and defaults
- **Comprehensive Chef data access**: nodes, roles, users, cookbooks, data bags, environments
- **Search capabilities** with both decoded and raw JSON results
- **Docker-based deployment** for easy installation
- **MCP protocol compliance** for seamless AI assistant integration

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `CHEF_USER` | Yes | Chef username for API authentication |
| `CHEF_KEY_PATH` | Yes | Path to Chef private key file (.pem) |
| `CHEF_SERVER_URL` | Yes | Chef Server base URL (without organization path) |
| `CHEF_DEFAULT_ORG` | No | Default organization to use when none specified |
| `CHEF_ORG_ALIASES` | No | Organization aliases in JSON or key=value format |

### Organization Support

The server supports multiple organizations through:
- **Default organization**: Set via `CHEF_DEFAULT_ORG`
- **Organization aliases**: Set via `CHEF_ORG_ALIASES` (e.g., `"qa=qa1,prod=fireamp_classic"`)
- **Per-request organization**: Specify in individual MCP tool calls

## IDE Integration

### Visual Studio Code

Add to your VSCode settings (`.vscode/settings.json` or global settings):

```json
{
  "mcpServers": {
    "chef-server": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "--env", "CHEF_USER=your-username",
        "--env", "CHEF_KEY_PATH=/chef/your-username.pem",
        "--env", "CHEF_SERVER_URL=https://your-chef-server.com/",
        "--env", "CHEF_ORG_ALIASES=qa=qa1,prod=production",
        "--env", "CHEF_DEFAULT_ORG=qa1",
        "--volume", "/path/to/.chef:/chef:ro",
        "ghcr.io/aknarts/chef-server-mcp:latest"
      ]
    }
  }
}
```

### GitHub Copilot

Add to your Copilot configuration:

```json
{
  "servers": {
    "chef-server": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "--env", "CHEF_USER=your-username",
        "--env", "CHEF_KEY_PATH=/chef/your-username.pem",
        "--env", "CHEF_SERVER_URL=https://your-chef-server.com/",
        "--env", "CHEF_ORG_ALIASES=qa=qa1,prod=production",
        "--env", "CHEF_DEFAULT_ORG=qa1",
        "--volume", "/path/to/your/.chef:/chef:ro",
        "ghcr.io/aknarts/chef-server-mcp:latest"
      ]
    }
  }
}
```

**Important Notes**:
- The `--volume` flag maps your local `.chef` directory to `/chef` inside the container
- The `CHEF_KEY_PATH` must point to the file path **inside** the container (e.g., `/chef/your-username.pem`)
- Make sure the volume path matches your actual local `.chef` directory location


## Available Tools

| Tool | Description |
|------|-------------|
| `listNodes` | List all node names |
| `getNode` | Get detailed node information |
| `listRoles` | List all role names |
| `getRole` | Get role definition and run lists |
| `listUsers` | List all user names |
| `getUser` | Get user details |
| `search` | Execute Chef search queries (decoded results) |
| `searchJSON` | Execute Chef search queries (raw JSON results) |
| `getOrganization` | Get organization details |
| `listCookbooks` | List cookbooks and their versions |
| `getCookbook` | Get cookbook metadata and files |
| `listDataBags` | List all data bag names |
| `listDataBagItems` | List items in a data bag |
| `getDataBagItem` | Get specific data bag item |
| `listEnvironments` | List all environments |
| `getEnvironment` | Get environment configuration |

All tools support optional `organization` parameter for multi-org setups.

## Development

For development instructions, building from source, and contributing guidelines, see [DEVELOPMENT.md](DEVELOPMENT.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
