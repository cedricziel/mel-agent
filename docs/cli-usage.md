# CLI Usage Guide

MEL Agent provides a modern command-line interface built with Cobra and Viper for flexible configuration and ease of use.

## Basic Commands

### Getting Help

```bash
./server --help                    # Show all commands
./server server --help             # Server-specific help
./server worker --help             # Worker-specific help
./server completion --help         # Shell completion help
```

### Server Command

Start the API server with embedded local workers:

```bash
./server server                    # Start with defaults (port 8080)
./server server --port 9090        # Custom port
./server server -p 8080            # Short flag version
```

### Worker Command

Start a remote worker process:

```bash
./server worker --token your-token                    # Basic worker
./server worker -t your-token -s http://localhost:8080 # With server URL
./server worker --token your-token \                   # Full configuration
  --server https://api.example.com \
  --id worker-production-1 \
  --concurrency 20
```

## Configuration Sources

MEL Agent supports multiple configuration sources with the following precedence:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration files**
4. **Default values** (lowest priority)

### Environment Variables

#### Legacy Variables (Backward Compatible)
```bash
export PORT=8080                           # Server port
export DATABASE_URL="postgres://..."       # Database connection
export MEL_SERVER_URL="http://..."         # Worker server URL
export MEL_WORKER_TOKEN="abc123"           # Worker authentication token
export MEL_WORKER_ID="worker-1"            # Worker identifier
```

#### New Prefixed Variables
```bash
export MEL_SERVER_PORT=8080                # Server port
export MEL_WORKER_SERVER="http://..."      # Worker server URL
export MEL_WORKER_CONCURRENCY=10           # Worker concurrency
```

### Configuration Files

MEL Agent automatically searches for configuration files in these locations:

1. `./config.yaml` (current directory)
2. `~/.mel-agent/config.yaml` (user home directory)  
3. `/etc/mel-agent/config.yaml` (system-wide)

Supported formats: **YAML**, **JSON**, **TOML**

#### Example config.yaml

```yaml
# Server configuration
server:
  port: "8080"

# Worker configuration  
worker:
  server: "https://api.production.example.com"
  token: "your-production-token"
  id: "worker-prod-1"
  concurrency: 15

# Database configuration
database:
  url: "postgres://user:pass@localhost:5432/melagent?sslmode=disable"
```

#### Example config.json

```json
{
  "server": {
    "port": "8080"
  },
  "worker": {
    "server": "https://api.example.com",
    "token": "your-token",
    "concurrency": 10
  }
}
```

## Configuration Examples

### Development Setup

```bash
# Start server with development settings
DATABASE_URL="postgres://localhost/dev" ./server server --port 3000

# Start worker connecting to development server
./server worker --token dev-token --server http://localhost:3000
```

### Production Setup

Create `/etc/mel-agent/config.yaml`:
```yaml
server:
  port: "8080"
database:
  url: "postgres://prod-user:password@db.example.com/melagent"
worker:
  server: "https://api.production.com"
  concurrency: 20
```

Then run:
```bash
# Server reads config automatically
./server server

# Worker with production token
./server worker --token $PRODUCTION_TOKEN
```

### Docker Configuration

```bash
# Using environment variables
docker run -e PORT=8080 -e DATABASE_URL="..." mel-agent ./server server

# Using config file volume
docker run -v /path/to/config.yaml:/etc/mel-agent/config.yaml mel-agent ./server server
```

## Shell Completion

MEL Agent provides auto-completion for bash, zsh, fish, and PowerShell.

### Installation

#### Bash
```bash
./server completion bash > /etc/bash_completion.d/mel-agent
source /etc/bash_completion.d/mel-agent
```

#### Zsh
```bash
./server completion zsh > ~/.zsh/completions/_mel-agent
# Add to ~/.zshrc: fpath=(~/.zsh/completions $fpath)
```

#### Fish
```bash
./server completion fish > ~/.config/fish/completions/mel-agent.fish
```

#### PowerShell
```powershell
./server completion powershell > mel-agent.ps1
# Import-Module ./mel-agent.ps1
```

### Usage
After installation, you can use tab completion:

```bash
./server <TAB>          # Shows: server, worker, completion, help
./server server --<TAB>  # Shows: --port, --help
./server worker --<TAB>  # Shows: --token, --server, --id, --concurrency, --help
```

## Advanced Usage

### Multiple Configuration Sources

```bash
# Config file sets defaults
echo "server: { port: '9000' }" > config.yaml

# Environment overrides config
export PORT=8080

# Flag overrides environment (final value: 7000)
./server server --port 7000
```

### Worker Scaling

```bash
# Start multiple workers with different IDs
./server worker --token $TOKEN --id worker-1 --concurrency 5 &
./server worker --token $TOKEN --id worker-2 --concurrency 5 &
./server worker --token $TOKEN --id worker-3 --concurrency 5 &

# Or use config file for each environment
./server worker --token $PROD_TOKEN --id worker-prod-$(hostname)
```

### Configuration Validation

```bash
# Test configuration without starting services
./server server --help     # Shows resolved port value
./server worker --help      # Shows resolved server URL
```

## Troubleshooting

### Common Issues

1. **Config file not found**: Check search paths and file permissions
2. **Environment variables ignored**: Ensure correct names (PORT vs MEL_SERVER_PORT)
3. **Worker connection failed**: Verify server URL and token
4. **Permission denied**: Check file permissions and user access

### Debug Configuration

```bash
# Show help to see resolved values
./server server --help | grep "default"
./server worker --help | grep "default"

# Test with explicit config
./server server --port 8080  # Should work
```

### Configuration File Debugging

```bash
# Test YAML syntax
cat config.yaml | python -c "import yaml, sys; yaml.safe_load(sys.stdin)"

# Test JSON syntax  
cat config.json | python -m json.tool
```

## Migration from Old CLI

If you're upgrading from the old manual CLI, here are the changes:

### Old vs New Commands

| Old Command | New Command |
|-------------|-------------|
| `./server` | `./server server` |
| `./server worker -token X` | `./server worker --token X` |
| No equivalent | `./server --help` |
| No equivalent | `./server completion bash` |

### Environment Variables

All existing environment variables continue to work:
- `PORT` → Still works
- `MEL_WORKER_TOKEN` → Still works  
- `MEL_SERVER_URL` → Still works
- `MEL_WORKER_ID` → Still works
- `DATABASE_URL` → Still works

### Breaking Changes

**None!** The new CLI is fully backward compatible.