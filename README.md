# Service CLI

A command-line interface tool for managing and deploying containerized services with built-in Tailscale VPN integration.

## Features

- Create Docker-based services with automatic Tailscale networking
- Manage service lifecycle (create, delete, list)
- Generate Docker Compose configurations automatically
- Provision Tailscale authentication keys via API
- Support for bare services without Tailscale integration

## Prerequisites

- Go 1.25.5 or later
- Docker and Docker Compose
- Tailscale account and API token

## Configuration

Create a `.env` file in the project root:

```env
TAILSCALE_API_TOKEN=tskey-api-YOUR_TOKEN_HERE
TAILSCALE_TAILNET=https://api.tailscale.com/api/v2/tailnet/your-tailnet/keys
SERVICE_DIR=/path/to/service/directory
```

## Usage

### Create a service

```bash
service create --name myapp --img nginx --port 8080
```

**Flags:**
- `-n, --name` (required) - Name of the service
- `-i, --img` (required) - Docker image to use
- `-p, --port` - Port to expose (default: 8080)
- `-b, --bare` - Create without Tailscale integration

### Delete a service

```bash
service delete --name myapp
```

**Flags:**
- `-n, --name` (required) - Name of the service
- `-r, --remove-key` - Also revoke the Tailscale authentication key

### List services

```bash
service list
```

### Check status

```bash
service status
```

## Project Structure

```
cli/
├── main.go              # Entry point
├── deploy.sh            # Build and deployment script
├── cmd/                 # Command definitions
│   ├── root.go          # Root command setup
│   ├── create.go        # Service creation
│   ├── delete.go        # Service deletion
│   ├── list.go          # Service listing
│   └── status.go        # Status check
└── internal/
    ├── config/
    │   └── config.go    # Configuration management
    └── util/
        └── helpers.go   # Utility functions
```

## Generated Service Structure

When you create a service, the following structure is generated:

```
service-name/
├── docker-compose.yml   # Container configuration
├── .env                 # Tailscale credentials
└── config/
    └── serve.json       # Tailscale serve config
```

The Docker Compose setup includes:
- A Tailscale sidecar container for VPN networking
- Your application container using the specified image

## Dependencies

| Package | Purpose |
|---------|---------|
| github.com/spf13/cobra | CLI framework |
| github.com/joho/godotenv | Environment variable loading |

## License

See [LICENSE](LICENSE) file.
