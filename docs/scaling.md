# Horizontal Scaling Guide

This document outlines MEL Agent's horizontal scaling capabilities, limitations, and best practices.

## ✅ What Works in Multi-Instance Deployments

### API Endpoints
- All REST API endpoints work correctly across multiple instances
- Database operations are properly isolated with connection pooling
- Stateless request processing ensures consistent responses
- Health checks (`/health`, `/ready`) work reliably for load balancers

### Worker Distribution
- Remote workers can connect to any API server instance
- Work distribution happens through the shared database queue
- Worker registration and heartbeat tracking work across instances
- Automatic failover when API server instances go down

### Database Operations
- Connection pooling optimized for multiple instances (25 max connections per instance)
- Proper transaction isolation prevents data corruption
- Migration system works correctly with concurrent instances
- All CRUD operations are stateless and thread-safe

## ⚠️ Current Limitations

### WebSocket Collaboration (Critical Issue)
**Problem**: WebSocket connections for real-time collaboration are stored in memory per instance.

**Impact**:
- Users connected to different API server instances can't collaborate in real-time
- WebSocket messages only reach clients on the same server instance
- Load balancer may route users to different servers, breaking collaboration

**Current Workaround**: Use sticky sessions in your load balancer:
```nginx
upstream api_servers {
    ip_hash;  # Route same client IP to same server
    server api-1:8080;
    server api-2:8080;
    server api-3:8080;
}
```

**Long-term Solution**: Implement Redis pub/sub for cross-server WebSocket message broadcasting.

### In-Memory State
**Problem**: Some components use in-memory state that isn't shared between instances.

**Affected Components**:
- WebSocket connection hubs (`internal/api/ws.go`)
- Global plugin registry (partially mitigated by consistent registration)
- Pending workflow calls in Mel instances

**Mitigation**: Most in-memory state is rebuilt on startup, but WebSocket state is lost.

## 🚀 Deployment Strategies

### 1. Simple Horizontal Scaling (Recommended)

Use Docker Compose built-in scaling:

```bash
# Start with 3 API servers and 2 workers
docker-compose -f docker-compose.scale.yml up -d --scale api=3 --scale worker=2

# Scale up during high traffic
docker-compose -f docker-compose.scale.yml up -d --scale api=5 --scale worker=4

# Scale down during low traffic  
docker-compose -f docker-compose.scale.yml up -d --scale api=2 --scale worker=1
```

### 2. Separate API and Worker Scaling

For different scaling needs:

```bash
# High API load, normal worker load
docker-compose -f docker-compose.scale.yml up -d --scale api=5 --scale worker=2

# Normal API load, high processing load
docker-compose -f docker-compose.scale.yml up -d --scale api=2 --scale worker=6
```

### 3. Geographic Distribution

Deploy API servers in multiple regions with centralized workers:

```yaml
# docker-compose.region-a.yml
services:
  api:
    command: ./server api-server --port 8080
    environment:
      - DATABASE_URL=postgres://user:pass@global-db:5432/melagent

# Workers connect to any region
  worker:
    command: ./server worker --token $TOKEN --server https://api.example.com
```

### 4. Production Kubernetes Deployment

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
    spec:
      containers:
      - name: api
        image: mel-agent:latest
        command: ["./server", "api-server"]
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: mel-agent-secrets
              key: database-url
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
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
    spec:
      containers:
      - name: worker
        image: mel-agent:latest
        command: ["./server", "worker"]
        env:
        - name: MEL_WORKER_TOKEN
          valueFrom:
            secretKeyRef:
              name: mel-agent-secrets
              key: worker-token
        - name: MEL_SERVER_URL
          value: "http://mel-agent-api-service"
```

## 📊 Load Balancer Configuration

### Nginx Configuration

For maximum compatibility with current limitations:

```nginx
upstream api_servers {
    # Use ip_hash for WebSocket compatibility
    ip_hash;
    
    server api-1:8080 max_fails=3 fail_timeout=30s;
    server api-2:8080 max_fails=3 fail_timeout=30s;
    server api-3:8080 max_fails=3 fail_timeout=30s;
}

server {
    location / {
        proxy_pass http://api_servers;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Standard headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### Alternative: Round Robin (REST-only)

If you don't need real-time collaboration:

```nginx
upstream api_servers {
    # Better load distribution for REST APIs
    least_conn;
    
    server api-1:8080;
    server api-2:8080;
    server api-3:8080;
}
```

## 🔧 Configuration Best Practices

### Database Connection Limits

With multiple API server instances, monitor your PostgreSQL connection limits:

```sql
-- Check current connections
SELECT count(*) FROM pg_stat_activity;

-- Check max connections
SHOW max_connections;

-- Increase if needed (requires restart)
ALTER SYSTEM SET max_connections = 200;
```

### Environment Variables

Set per-instance limits:

```yaml
services:
  api:
    environment:
      # Reduce per-instance connections for more instances
      - DB_MAX_OPEN_CONNS=15   # Default: 25
      - DB_MAX_IDLE_CONNS=5    # Default: 10
```

### Monitoring

Monitor key metrics across all instances:

```bash
# Health checks across all instances
curl http://api-1:8080/health
curl http://api-2:8080/health
curl http://api-3:8080/health

# Database connection pools
curl http://api-1:8080/ready | jq '.checks.database'

# Worker distribution
docker-compose logs worker | grep "claimed work"
```

## 🛣️ Roadmap for Full WebSocket Scaling

To fully support WebSocket scaling without sticky sessions:

1. **Phase 1**: Implement Redis pub/sub for WebSocket message broadcasting
2. **Phase 2**: Add WebSocket connection registry in Redis
3. **Phase 3**: Create dedicated WebSocket service with full clustering support
4. **Phase 4**: Implement cross-server user presence tracking

## 📋 Quick Start Testing

Test your scaled deployment:

```bash
# Start scaled services
./test-scaling.sh

# Test load balancing
for i in {1..10}; do
  curl -s http://localhost:8080/health | jq '.timestamp'
done

# Test API functionality
curl http://localhost:8080/api/node-types | jq 'length'

# Monitor logs
docker-compose -f docker-compose.scale.yml logs -f api worker
```

## 🔍 Troubleshooting

### Common Issues

1. **WebSocket connections dropping**: Use sticky sessions or upgrade to Redis pub/sub
2. **Database connection limits**: Reduce per-instance connection pool size
3. **Inconsistent behavior**: Check that all instances are running the same version
4. **Load balancer health checks failing**: Verify `/health` endpoint accessibility

### Debug Commands

```bash
# Check which instance handled the request
curl -H "X-Debug: true" http://localhost:8080/health

# Monitor connection distribution
docker-compose -f docker-compose.scale.yml exec nginx cat /var/log/nginx/access.log

# Check database connections per instance
docker-compose -f docker-compose.scale.yml logs api | grep "Database connected"
```