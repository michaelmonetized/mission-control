# Mission Control VM Manager

A high-performance Go microservice for managing isolated VMs, terminal relay, and usage billing.

## Features

✅ **Fly.io Machines API Integration** — Create, list, stop, destroy VMs  
✅ **VM Lifecycle Management** — Spinup, mount repos, inject env, shutdown  
✅ **Terminal Relay** — WebSocket server for xterm.js client (1000+ concurrent connections)  
✅ **Auto-Scaling** — Enforce max 100 VMs per org  
✅ **Cost Tracking** — Per-minute compute billing ($0.02/min)  
✅ **Graceful Shutdown** — Kill VMs idle >2 hours  
✅ **Health Checks** — Periodic VM status monitoring  
✅ **Metrics Persistence** — Track usage and save to disk  

## Architecture

```
┌─────────────────────────────────┐
│   Mission Control Web App        │
│  (Next.js + Convex)             │
└────────────┬────────────────────┘
             │ HTTP + WebSocket
             ▼
┌─────────────────────────────────┐
│   VM Manager (this service)      │
├─────────────────────────────────┤
│ • Fly.io Client (API)            │
│ • VM Manager (lifecycle)         │
│ • Terminal Relay (WebSocket)     │
│ • Metrics Tracker (billing)      │
│ • Config (environment)           │
└────────────┬────────────────────┘
             │ HTTPS + WebSocket
             ▼
┌─────────────────────────────────┐
│   Fly.io Machines               │
│  (Isolated user VMs)            │
└─────────────────────────────────┘
```

## Quick Start

### 1. Install Dependencies

```bash
cd services/vm-manager
go mod tidy
```

### 2. Set Environment Variables

```bash
export FLY_API_TOKEN="your-fly-api-token"
export FLY_APP="mission-control-vms"
export VM_MANAGER_PORT=8080
export MAX_VMS_PER_ORG=100
export ALLOW_ORIGIN="*"
export LOG_LEVEL="info"
```

### 3. Run

```bash
go run main.go
```

Server starts on `http://localhost:8080`

## API Endpoints

### VMs

```bash
# Create a VM
POST /api/vms
{
  "user_id": "user-123",
  "org_id": "org-456",
  "repo_url": "https://github.com/user/repo.git",
  "repo_ref": "main",
  "api_key": "sk-...",
  "region": "ord",
  "cpus": 2,
  "memory_mb": 4096
}

# Get VM details
GET /api/vms/{vm_id}

# List user VMs
GET /api/vms/user/{user_id}

# Stop a VM
POST /api/vms/{vm_id}/stop

# Destroy a VM
DELETE /api/vms/{vm_id}

# Record user activity
POST /api/vms/{vm_id}/activity
```

### Terminal

```bash
# Connect terminal (WebSocket)
WS /api/terminal/connect?vm_id=xxx&client_id=yyy

# List connected clients
GET /api/terminal/clients?vm_id=xxx

# Disconnect client
DELETE /api/terminal/clients/{client_id}
```

### System

```bash
# Health check
GET /health

# Metrics
GET /metrics

# System stats
GET /api/system/stats

# Cleanup stale connections
POST /api/system/cleanup
```

## Performance Targets

| Metric | Target | Implementation |
|--------|--------|-----------------|
| VM Spinup | <2s | Fly.io machine creation is async |
| Terminal Latency | <100ms | WebSocket direct relay |
| Concurrent Connections | 100+ | Goroutine-per-client, WebSocket pooling |
| Metrics Persistence | 5m interval | Background goroutine |
| Health Checks | 30s interval | Concurrent polling |

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `FLY_API_TOKEN` | (required) | Fly.io API authentication |
| `FLY_APP` | `mission-control-vms` | Fly.io app name |
| `VM_MANAGER_PORT` | `8080` | Server port |
| `MAX_VMS_PER_ORG` | `100` | Scaling limit |
| `ALLOW_ORIGIN` | `*` | CORS origin |
| `METRICS_PATH` | `/tmp/vm-manager-metrics.json` | Metrics file location |
| `LOG_LEVEL` | `info` | Logging level |

## VM Lifecycle

```
1. CreateVM
   └─> CreateMachineInput to Fly.io
   └─> Get machine ID
   └─> Store locally
   └─> Start health check

2. UpdateActivity (user interacts)
   └─> Record last activity timestamp
   └─> Used for graceful shutdown

3. GracefulShutdown (idle >2h)
   └─> Check all VMs every 5 minutes
   └─> Destroy if idle duration > 2 hours
   └─> Record usage and cost
   └─> Update metrics

4. Health Check (every 30s)
   └─> Fetch machine state from Fly.io
   └─> Update local status
   └─> Log failures
```

## Cost Tracking

**Billing Model:** $0.02 per minute

**Flow:**
1. VM created at time T0
2. User interacts → LastActivity updated
3. VM idle >2 hours → Graceful shutdown triggered
4. Billed minutes = (T_destroy - T_create) in minutes
5. Cost = minutes × $0.02
6. Record to usage metrics
7. Write metrics.json every 5 minutes

## Testing

```bash
# Build
go build -o vm-manager main.go

# Run tests (if added)
go test ./...

# Manual test: health check
curl http://localhost:8080/health

# Manual test: create VM (requires FLY_API_TOKEN)
curl -X POST http://localhost:8080/api/vms \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user",
    "org_id": "test-org",
    "repo_url": "https://github.com/test/repo.git",
    "api_key": "sk-test",
    "region": "ord"
  }'
```

## Deployment

### Docker

```dockerfile
FROM golang:1.22-alpine
WORKDIR /app
COPY . .
RUN go build -o vm-manager main.go
EXPOSE 8080
CMD ["./vm-manager"]
```

```bash
docker build -t mission-control-vm-manager .
docker run -p 8080:8080 \
  -e FLY_API_TOKEN="..." \
  mission-control-vm-manager
```

### Fly.io

```bash
fly launch --name mission-control-vm-manager
fly secrets set FLY_API_TOKEN="..."
fly deploy
```

## Monitoring

### Logs

```bash
tail -f /var/log/vm-manager.log
```

### Metrics

```bash
curl http://localhost:8080/metrics | jq .
```

```json
{
  "total_vms_created": 145,
  "total_vms_destroyed": 142,
  "total_billed": 2850.50,
  "usage_by_org": {
    "org-456": [
      {
        "user_id": "user-123",
        "minutes": 125.5,
        "cost": 2.51,
        "date": "2026-03-21"
      }
    ]
  },
  "last_updated": 1711000000
}
```

## Architecture Decisions

### Why Go?
- Fast startup and execution
- Low memory footprint (important for 100+ VMs)
- Excellent concurrency with goroutines
- Easy deployment (single binary)

### Why WebSocket?
- Full-duplex terminal communication
- Lower latency than polling
- Persistent connection reduces overhead
- Native browser support

### Why Fly.io?
- Global edge deployment
- Per-second billing (cost-effective)
- Docker-native (easy container orchestration)
- Private networks for isolation

## Future Improvements

- [ ] Prometheus metrics export
- [ ] gRPC interface for high-performance client
- [ ] Terminal multiplexing (multi-pane)
- [ ] VM snapshots and persistence
- [ ] Real-time collaboration (shared terminal)
- [ ] Circuit breaker for Fly.io API

## Contributing

1. Fork the repo
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a PR

## License

MIT
