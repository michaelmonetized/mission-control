# VM Manager — Deployment Guide

This document covers deploying the VM Manager service to production.

## Prerequisites

- Fly.io account with API token
- Docker (for containerized deployment)
- Go 1.22+ (for local development)
- 100+ concurrent connection capacity

## Local Development

### 1. Clone and Setup

```bash
cd mission-control/services/vm-manager
cp .env.example .env

# Edit .env with your Fly.io API token
vim .env
```

### 2. Run Locally

```bash
go run main.go
```

Server starts on `http://localhost:8080`

### 3. Test Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics

# System stats
curl http://localhost:8080/api/system/stats
```

## Docker Deployment

### 1. Build Image

```bash
docker build -t mission-control-vm-manager:latest .
```

### 2. Run Container

```bash
docker run -d \
  -p 8080:8080 \
  -e FLY_API_TOKEN="your-token" \
  -e FLY_APP="mission-control-vms" \
  --name vm-manager \
  mission-control-vm-manager:latest
```

### 3. Check Logs

```bash
docker logs -f vm-manager
```

## Fly.io Deployment

### 1. Create App (first time only)

```bash
fly launch
# Choose name: mission-control-vm-manager
# Choose region: ord (Chicago)
```

### 2. Set Secrets

```bash
fly secrets set FLY_API_TOKEN="your-fly-api-token"
```

### 3. Deploy

```bash
fly deploy
```

### 4. Monitor

```bash
# Check status
fly status

# View logs
fly logs

# Monitor health
fly monitor
```

## Kubernetes Deployment

For running on your own Kubernetes cluster:

### 1. Create ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: vm-manager-config
spec:
  data:
    FLY_APP: "mission-control-vms"
    VM_MANAGER_PORT: "8080"
    MAX_VMS_PER_ORG: "100"
    LOG_LEVEL: "info"
```

### 2. Create Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: vm-manager-secret
type: Opaque
stringData:
  FLY_API_TOKEN: "your-fly-api-token"
```

### 3. Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vm-manager
spec:
  replicas: 2
  selector:
    matchLabels:
      app: vm-manager
  template:
    metadata:
      labels:
        app: vm-manager
    spec:
      containers:
      - name: vm-manager
        image: mission-control-vm-manager:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: vm-manager-config
        - secretRef:
            name: vm-manager-secret
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            cpu: "500m"
            memory: "256Mi"
          limits:
            cpu: "1000m"
            memory: "512Mi"
      affinity:
        podAntiAffinity:
          preferred:
          - weight: 100
            preference:
              matchExpressions:
              - key: app
                operator: In
                values:
                - vm-manager
```

### 4. Create Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: vm-manager
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app: vm-manager
```

### 5. Deploy

```bash
kubectl apply -f config-map.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `FLY_API_TOKEN` | ✅ | - | Fly.io API token |
| `FLY_APP` | ❌ | `mission-control-vms` | Fly.io app name |
| `VM_MANAGER_PORT` | ❌ | `8080` | Server port |
| `MAX_VMS_PER_ORG` | ❌ | `100` | Max VMs per organization |
| `ALLOW_ORIGIN` | ❌ | `*` | CORS allowed origin |
| `METRICS_PATH` | ❌ | `/tmp/vm-manager-metrics.json` | Metrics file path |
| `LOG_LEVEL` | ❌ | `info` | Log level |

## Monitoring

### Health Checks

The service provides a health endpoint at `GET /health`:

```json
{
  "status": "ok",
  "running_vms": 45,
  "total_billed": 2850.50,
  "timestamp": 1711000000
}
```

### Metrics

View metrics at `GET /metrics`:

```json
{
  "total_vms_created": 1200,
  "total_vms_destroyed": 1155,
  "total_billed": 22500.00,
  "usage_by_org": {...},
  "last_updated": 1711000000
}
```

### Logs

Monitor logs for errors and activity:

```bash
# Docker
docker logs -f vm-manager

# Fly.io
fly logs --follow

# Kubernetes
kubectl logs -f deployment/vm-manager
```

## Performance Tuning

### Max Connections Per Instance

Adjust via `MaxVMs` configuration:

```go
vmManager := vm.NewManager(flyClient, metricsTracker, 100)
```

For larger deployments, increase to 200-500 depending on machine capacity.

### Terminal Relay Buffer

In `relay/terminal.go`:

```go
terminalRelay := relay.NewTerminalRelay(1000) // max 1000 connections
```

### WebSocket Buffer Size

In `internal/relay/terminal.go`:

```go
upgrader := websocket.Upgrader{
    ReadBufferSize:  1024,   // Increase for large messages
    WriteBufferSize: 1024,
}
```

## Scaling

### Horizontal Scaling

For high load, run multiple instances behind a load balancer:

```
Load Balancer
├── VM Manager #1 (Port 8080)
├── VM Manager #2 (Port 8080)
├── VM Manager #3 (Port 8080)
└── VM Manager #N (Port 8080)
```

Each instance:
- Manages its own local VM state
- Reports to shared metrics store
- Independently tracks idle VMs

### Metrics Persistence

Enable centralized metrics by mounting shared storage:

```bash
docker run -d \
  -v /shared-storage/metrics:/metrics \
  -e METRICS_PATH=/metrics/vm-manager.json \
  ...
```

## Disaster Recovery

### Backup Metrics

```bash
# Fly.io
fly ssh console
cp /tmp/vm-manager-metrics.json /backup/

# Docker
docker cp vm-manager:/tmp/vm-manager-metrics.json ./backup/
```

### Restore Metrics

```bash
docker cp ./backup/vm-manager-metrics.json vm-manager:/tmp/
docker restart vm-manager
```

## Rollback

### Rollback to Previous Version

```bash
# Fly.io
fly releases
fly rollback <release-id>

# Docker
docker run --name vm-manager-old mission-control-vm-manager:v1.0.0
```

## Troubleshooting

### Service won't start

```bash
# Check logs
fly logs

# Verify environment
fly secrets list

# Check regional capacity
fly regions list
```

### High VM creation latency

```bash
# Check Fly.io API status
curl https://api.machines.dev/v1/health

# Verify network connectivity
ping -c 1 api.machines.dev

# Check logs for rate limiting
fly logs --grep "rate"
```

### Terminal connections dropping

```bash
# Increase read/write buffer size in terminal.go

# Check WebSocket upgrade issues
fly logs --grep "websocket"

# Verify CORS settings
curl -H "Origin: your-origin" http://localhost:8080/health
```

### OOM (Out of Memory)

```bash
# Increase VM memory
fly machine update <machine-id> --memory 512

# Monitor memory usage
fly logs --grep "memory"

# Consider reducing MaxVMs per instance
```

## Success Criteria

After deployment, verify:

- ✅ Health endpoint returns 200 OK
- ✅ Can create VMs via Fly.io API
- ✅ WebSocket terminal connections work
- ✅ Metrics are being tracked and saved
- ✅ Graceful shutdown kills idle VMs
- ✅ No errors in logs for 1+ hour
- ✅ <2 second VM spinup latency
- ✅ 100+ concurrent WebSocket connections

## Support

Issues? Check:

1. `fly logs` for error messages
2. `FLY_API_TOKEN` is valid
3. `FLY_APP` exists in Fly.io
4. Firewall allows outbound HTTPS to api.machines.dev
5. Sufficient capacity in chosen region

Debug with verbose logging:

```bash
fly secrets set LOG_LEVEL=debug
fly deploy
fly logs --follow
```
