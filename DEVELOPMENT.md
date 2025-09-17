# Development Guide

This document contains information for developers working on the chef-server-mcp project.

## Building from Source

### Prerequisites
- Go 1.25.x or later
- Docker (for container builds)

### Building the Binary

Build the MCP server:
```bash
go build -o mcp-chef ./cmd/mcp-chef
```

### Running Locally

Run with environment variables:
```bash
CHEF_USER=myuser \
CHEF_KEY_PATH=$HOME/.chef/myuser.pem \
CHEF_SERVER_URL=https://chef.example.com \
CHEF_DEFAULT_ORG=acme \
./mcp-chef
```

### Development Configuration

For local development, you can use the wrapper script in `scripts/mcp-chef-wrapper.sh`.

#### VSCode Configuration (Development)
```json
{
  "mcpServers": {
    "chef-server": {
      "command": "/path/to/chef-server-mcp/scripts/mcp-chef-wrapper.sh",
      "env": {
        "CHEF_USER": "your-username",
        "CHEF_KEY_PATH": "/path/to/.chef/your-username.pem",
        "CHEF_SERVER_URL": "https://your-chef-server.com/",
        "CHEF_ORG_ALIASES": "qa=qa1,prod=production",
        "CHEF_DEFAULT_ORG": "qa1",
        "NO_REBUILD": "1"
      }
    }
  }
}
```

#### GitHub Copilot Configuration (Development)
```json
{
  "servers": {
    "chef-server": {
      "command": "/path/to/chef-server-mcp/scripts/mcp-chef-wrapper.sh",
      "env": {
        "CHEF_USER": "your-username",
        "CHEF_KEY_PATH": "/path/to/.chef/your-username.pem",
        "CHEF_SERVER_URL": "https://your-chef-server.com/",
        "CHEF_ORG_ALIASES": "qa=qa1,prod=production",
        "CHEF_DEFAULT_ORG": "qa1",
        "RESET_CHEF_CLIENT": "true"
      }
    }
  }
}
```

## Docker Development

### Building Docker Image
```bash
docker build --build-arg VERSION=dev -t local/mcp-chef:dev .
```

### Testing Docker Image
```bash
docker run --rm \
  -e CHEF_USER=myuser \
  -e CHEF_KEY_PATH=/chef/myuser.pem \
  -e CHEF_SERVER_URL=https://chef.example.com \
  -e CHEF_DEFAULT_ORG=acme \
  -v /path/to/.chef:/chef:ro \
  local/mcp-chef:dev
```

## CI/CD

The project uses GitHub Actions for continuous integration and deployment:

- **Build and Test**: Runs on every push and pull request
- **Docker Publish**: Builds and publishes multi-arch Docker images on version tags
- **Release**: Creates release artifacts on version tags

### Creating a Release

1. Create and push a version tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. GitHub Actions will automatically:
   - Build and test the code
   - Build multi-architecture Docker images
   - Push images to GitHub Container Registry
   - Create release binaries

## Project Structure

```
.
├── cmd/
│   └── mcp-chef/          # Main MCP server application
├── internal/
│   ├── chefapi/           # Chef API client
│   ├── config/            # Configuration handling
│   └── version/           # Version information
├── scripts/
│   └── mcp-chef-wrapper.sh # Development wrapper script
├── .github/
│   └── workflows/         # GitHub Actions workflows
├── Dockerfile             # Container build configuration
└── README.md             # User documentation
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Ensure CI passes
6. Submit a pull request

## Testing

Run tests:
```bash
go test -race -coverprofile=coverage.out ./...
```

Run with coverage:
```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
