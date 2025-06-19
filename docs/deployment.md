# Deployment Guide

This guide covers deploying MEL Agent in various environments using the modern CLI.

## Quick Deploy Options

### Docker Compose (Recommended for Development)

```bash
# Clone and start
git clone https://github.com/cedricziel/mel-agent.git
cd mel-agent
docker compose up --build

# Access your instance
open http://localhost:5173  # Web UI
curl http://localhost:8080/health  # API Health Check
```

### Single Binary (Production)

```bash
# Build
go build ./cmd/server
chmod +x server

# Configure
export DATABASE_URL="postgres://user:pass@localhost:5432/melagent"
export PORT=8080

# Run
./server server
```

## Production Deployment

### 1. System Service (systemd)

Create `/etc/systemd/system/mel-agent.service`:

```ini
[Unit]
Description=MEL Agent API Server
After=network.target postgresql.service

[Service]
Type=simple
User=melagent
Group=melagent
WorkingDirectory=/opt/mel-agent
ExecStart=/opt/mel-agent/server server
Restart=always
RestartSec=5

# Configuration
Environment=DATABASE_URL=postgres://user:pass@localhost:5432/melagent
Environment=PORT=8080

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/mel-agent

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable mel-agent
sudo systemctl start mel-agent
sudo systemctl status mel-agent
```

### 2. Docker Production

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  api:
    build: .
    command: ./server api-server  # Use api-server for horizontal scaling
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://postgres:${DB_PASSWORD}@db:5432/melagent
      - PORT=8080
    depends_on:
      - db
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      replicas: 2  # Scale API servers independently

  worker:
    build: .
    command: ./server worker --token ${WORKER_TOKEN}
    environment:
      - MEL_SERVER_URL=http://api:8080
      - MEL_WORKER_TOKEN=${WORKER_TOKEN}
    depends_on:
      - api
    restart: unless-stopped
    deploy:
      replicas: 3  # Scale workers independently

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=melagent
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
```

Deploy:
```bash
export DB_PASSWORD="secure-db-password"
export WORKER_TOKEN="secure-worker-token"
docker compose -f docker-compose.prod.yml up -d
```

### 3. Kubernetes

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mel-agent-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mel-agent-api
  template:
    metadata:
      labels:
        app: mel-agent-api
    spec:
      containers:
      - name: api
        image: mel-agent:latest
        command: ["./server", "api-server"]  # Use api-server for K8s scaling
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: mel-agent-secrets
              key: database-url
        - name: PORT
          value: "8080"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mel-agent-worker
spec:
  replicas: 5
  selector:
    matchLabels:
      app: mel-agent-worker
  template:
    metadata:
      labels:
        app: mel-agent-worker
    spec:
      containers:
      - name: worker
        image: mel-agent:latest
        command: ["./server", "worker"]
        args: ["--token", "$(WORKER_TOKEN)", "--concurrency", "10"]
        env:
        - name: MEL_SERVER_URL
          value: "http://mel-agent-api:8080"
        - name: WORKER_TOKEN
          valueFrom:
            secretKeyRef:
              name: mel-agent-secrets
              key: worker-token
---
apiVersion: v1
kind: Service
metadata:
  name: mel-agent-api
spec:
  selector:
    app: mel-agent-api
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

Deploy:
```bash
kubectl apply -f k8s/
```

## Configuration Management

### Environment-based Configuration

```bash
# Development
export MEL_ENV=development
export DATABASE_URL="postgres://localhost/melagent_dev"
export PORT=3000
./server server

# Staging  
export MEL_ENV=staging
export DATABASE_URL="postgres://staging-db/melagent"
export PORT=8080
./server server

# Production
export MEL_ENV=production
export DATABASE_URL="postgres://prod-db/melagent"
export PORT=8080
./server server
```

### Config File Management

```bash
# Development
cp config.yaml.example config.dev.yaml
# Edit config.dev.yaml for development settings

# Production
cat > /etc/mel-agent/config.yaml << EOF
server:
  port: "8080"
database:
  url: "postgres://prod-user:${DB_PASSWORD}@prod-db:5432/melagent"
worker:
  concurrency: 20
EOF
```

## Horizontal Scaling

### API Server Scaling

MEL Agent supports horizontal scaling by separating API servers from workers:

```bash
# Traditional: Single server with embedded workers
./server server  # API + embedded workers on port 8080

# Horizontal scaling: Separate API servers and workers
./server api-server --port 8080  # API server only (no workers)
./server api-server --port 8081  # Additional API server
./server worker --token $TOKEN   # Dedicated worker process
./server worker --token $TOKEN   # Additional worker process
```

#### Load Balancer Setup

```nginx
# nginx.conf
upstream mel_api {
    server api1.example.com:8080;
    server api2.example.com:8080;
    server api3.example.com:8080;
}

server {
    listen 80;
    location / {
        proxy_pass http://mel_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

#### Worker Scaling

```bash
# Start multiple workers on same machine
for i in {1..5}; do
  ./server worker --token $TOKEN --id "worker-${HOSTNAME}-${i}" &
done

# Start workers on different machines
# Machine 1
./server worker --token $TOKEN --id "worker-machine1" --server https://api.example.com

# Machine 2  
./server worker --token $TOKEN --id "worker-machine2" --server https://api.example.com
```

## Advanced Horizontal Scaling

### Docker Compose Scaling

MEL Agent includes a pre-configured scaling setup with load balancer:

```bash
# Start with default setup (1 API server, 1 worker)
docker-compose -f docker-compose.scale.yml up -d

# Scale to multiple instances
docker-compose -f docker-compose.scale.yml up -d --scale api=3 --scale worker=2

# Scale during runtime
docker-compose -f docker-compose.scale.yml up -d --scale api=5 --scale worker=4

# Scale down
docker-compose -f docker-compose.scale.yml up -d --scale api=2 --scale worker=1
```

The scaling setup includes:
- **Nginx load balancer** on port 8080
- **Multiple API server instances** using `api-server` command
- **Multiple worker instances** connecting through the load balancer
- **Shared PostgreSQL database** with optimized connection pooling
- **Health checks** for all components

### Scaling Considerations

⚠️ **WebSocket Limitations**: The current implementation has limitations with WebSocket-based real-time collaboration across multiple API server instances. See [scaling.md](scaling.md) for details and workarounds.

✅ **What Works**:
- All REST API endpoints scale perfectly
- Worker distribution across instances
- Database operations with connection pooling
- Health checks and graceful shutdown

### Load Balancer Configuration

The included Nginx configuration supports:

```nginx
upstream api_servers {
    least_conn;  # Distribute load evenly
    server api:8080 max_fails=3 fail_timeout=30s;
}

server {
    location / {
        proxy_pass http://api_servers;
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Testing Scaling

Use the included test script:

```bash
# Test the scaled deployment
./test-scaling.sh

# Manual testing
curl http://localhost:8080/health
curl http://localhost:8080/ready  
curl http://localhost:8080/api/node-types
```

### Auto-scaling with Docker Swarm

```yaml
# docker-compose.swarm.yml
version: '3.8'

services:
  worker:
    image: mel-agent:latest
    command: ./server worker --token ${WORKER_TOKEN}
    environment:
      - MEL_SERVER_URL=http://api:8080
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
```

Deploy with:
```bash
docker stack deploy -c docker-compose.swarm.yml mel-agent
```

## Monitoring & Health Checks

### Health Endpoint

```bash
# Check API health
curl http://localhost:8080/health

# Health check in scripts
if curl -f http://localhost:8080/health; then
  echo "API is healthy"
else
  echo "API is down"
  exit 1
fi
```

### Process Monitoring

```bash
# Check if server is running
pgrep -f "server server" || echo "Server not running"

# Check worker processes
pgrep -f "server worker" | wc -l  # Count running workers
```

### Log Management

```bash
# Systemd logs
journalctl -u mel-agent -f

# Docker logs
docker logs mel-agent-api-1 -f

# File-based logging (add to service)
./server server 2>&1 | tee /var/log/mel-agent/server.log
./server worker --token $TOKEN 2>&1 | tee /var/log/mel-agent/worker.log
```

## Security Considerations

### Token Management

```bash
# Generate secure tokens
export WORKER_TOKEN=$(openssl rand -hex 32)

# Store in secrets management
# AWS Secrets Manager, HashiCorp Vault, Kubernetes Secrets, etc.
```

### Network Security

```bash
# Bind to specific interface
./server server --port 8080  # Binds to all interfaces

# Use reverse proxy (nginx, traefik, etc.)
# Let proxy handle TLS termination
```

### Database Security

```bash
# Use connection pooling and SSL
export DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=require&pool_max_conns=20"
```

## Backup & Recovery

### Database Backup

```bash
# Regular backup
pg_dump $DATABASE_URL > backup-$(date +%Y%m%d).sql

# Restore
psql $DATABASE_URL < backup-20240101.sql
```

### Configuration Backup

```bash
# Backup configuration
cp /etc/mel-agent/config.yaml /backup/config-$(date +%Y%m%d).yaml

# Backup environment
env | grep MEL_ > /backup/env-$(date +%Y%m%d).txt
```

## Troubleshooting

### Common Issues

1. **Server won't start**
   ```bash
   ./server server --help  # Check configuration
   ./server server --port 8080  # Try explicit port
   ```

2. **Workers can't connect**
   ```bash
   ./server worker --help  # Check configuration
   curl http://api-server:8080/health  # Test connectivity
   ```

3. **Database connection issues**
   ```bash
   psql $DATABASE_URL  # Test database directly
   ```

### Debug Mode

```bash
# Enable verbose logging (if implemented)
DEBUG=true ./server server
VERBOSE=true ./server worker --token $TOKEN
```