# proxmox-mcp-go

A Proxmox MCP (Model Context Protocol) server written in Go. Exposes Proxmox VE management tools to AI agents over HTTP using the MCP streamable HTTP transport.

## Getting Started

### Prerequisites

- Go 1.24+
- Docker (optional)

### Configuration

Copy and edit the config file:

```yaml
bind_port: 3001
verbose: true
pve_url: "https://your-proxmox-host:8006"
pve_token_id: "root@pam!mcp"
pve_token: "your-api-token-secret"
mcp_api_key: "your-secret-key-here"
# mcp_stdio: false
```

| Field | Description |
|-------|-------------|
| `bind_host` | HTTP server bind address (default: `0.0.0.0`) |
| `bind_port` | HTTP server port (default: `3001`) |
| `pve_url` | Proxmox VE API URL |
| `pve_token_id` | API token ID (e.g. `root@pam!mcp`) |
| `pve_token` | API token secret |
| `mcp_api_key` | API key for authenticating MCP endpoint requests (Bearer token) |
| `mcp_stdio` | Enable stdio transport (default: `false`) |

### Running

```bash
go run main.go --config=$PWD/config/config.yaml
```

### Docker

```bash
./run-docker.sh

# Or with custom config/port
CONFIG_FILE=/path/to/config.yaml PORT=3001 ./run-docker.sh
```

## MCP Endpoint

The MCP server is available at `POST /mcp` using the streamable HTTP transport.

### Connecting from Claude Code

```json
{
  "mcpServers": {
    "proxmox": {
      "type": "url",
      "url": "http://your-host:3001/mcp",
      "headers": {
        "Authorization": "Bearer your-secret-key-here"
      }
    }
  }
}
```

When `mcp_api_key` is set, all requests to `/mcp` must include a `Authorization: Bearer <key>` header. If the key is not set, the endpoint is unauthenticated.

### Available Tools

| Category | Tools |
|----------|-------|
| Cluster | `get_cluster_status`, `list_nodes`, `get_node_status` |
| Guest | `list_vms`, `list_containers`, `list_cluster_resources`, `start_guest`, `stop_guest`, `get_next_id` |
| Create | `create_vm`, `create_container`, `clone_guest`, `delete_guest`, `convert_to_template` |
| Snapshot | `list_snapshots`, `create_snapshot`, `rollback_snapshot`, `delete_snapshot` |
| Backup | `backup_guest`, `list_backups`, `restore_backup` |
| Storage | `list_storage`, `list_templates`, `list_isos`, `download_template` |
| Task | `list_tasks`, `get_task_status`, `get_task_log` |

## API

```
GET  /api/v1/        # Default page
GET  /api/v1/health  # Health check
POST /mcp            # MCP streamable HTTP endpoint
```

## Project Structure

```
main.go              # Entry point
app/                 # Application lifecycle, logging, signal handling
server/              # HTTP server, routing, middleware
config/              # Configuration parsing
mcp/                 # MCP server, Proxmox client, tool registrations
```

## Development

```bash
# Build binary
go build -o ./tmp/main.exe .

# Run tests
go test ./...

# Lint
golangci-lint run --fix

# Update vendor after dependency changes
go mod tidy && go mod vendor
```
