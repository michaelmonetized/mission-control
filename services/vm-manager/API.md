# VM Manager API Reference

Complete API documentation for the VM Manager microservice.

## Base URL

```
http://localhost:8080
https://mission-control-vm-manager.fly.dev
```

## Authentication

Currently, no authentication is required. In production, add JWT or API key verification.

## Response Format

All responses are JSON. Errors return appropriate HTTP status codes.

```json
{
  "error": "Description of error"
}
```

---

## Health & Monitoring

### GET /health

System health check.

**Response:** `200 OK`

```json
{
  "status": "ok",
  "running_vms": 42,
  "total_billed": 2850.50,
  "timestamp": 1711000000
}
```

---

### GET /metrics

Detailed metrics and usage tracking.

**Response:** `200 OK`

```json
{
  "total_vms_created": 1200,
  "total_vms_destroyed": 1155,
  "total_billed": 22500.00,
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

---

### GET /api/system/stats

Current system statistics.

**Response:** `200 OK`

```json
{
  "running_vms": 42,
  "connected_clients": 8,
  "timestamp": 1711000000
}
```

---

### POST /api/system/cleanup

Clean up stale WebSocket connections (idle >1 hour).

**Response:** `200 OK`

```json
{
  "status": "cleanup_complete"
}
```

---

## VM Management

### POST /api/vms

Create a new VM.

**Request Body:**

```json
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
```

**Parameters:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `user_id` | string | ✅ | - | Unique user identifier |
| `org_id` | string | ✅ | - | Organization ID (for scaling limits) |
| `repo_url` | string | ✅ | - | Git repository URL |
| `repo_ref` | string | ❌ | `main` | Branch or commit to clone |
| `api_key` | string | ✅ | - | Claude API key (injected as env var) |
| `region` | string | ❌ | `ord` | Fly.io region (ord, sjc, iad, etc.) |
| `cpus` | integer | ❌ | `2` | vCPU count |
| `memory_mb` | integer | ❌ | `4096` | RAM in MB |

**Response:** `201 Created`

```json
{
  "id": "org-456-abc123def456",
  "user_id": "user-123",
  "org_id": "org-456",
  "repo_url": "https://github.com/user/repo.git",
  "created_at": "2026-03-21T15:34:18Z",
  "last_activity": "2026-03-21T15:34:18Z",
  "status": "starting",
  "machine_id": "abc123def456",
  "terminal_url": "wss://abc123def456.fly.dev/terminal",
  "cpu_count": 2,
  "memory_mb": 4096,
  "billed_minutes": 0
}
```

**Errors:**

- `400 Bad Request` — Missing required fields
- `500 Internal Server Error` — Fly.io API error
- `503 Service Unavailable` — Org has reached max VMs

---

### GET /api/vms/{vm_id}

Get details of a specific VM.

**Response:** `200 OK`

```json
{
  "id": "org-456-abc123def456",
  "status": "running",
  "last_activity": "2026-03-21T15:35:00Z",
  ...
}
```

**Errors:**

- `404 Not Found` — VM does not exist

---

### GET /api/vms/user/{user_id}

List all VMs for a user.

**Response:** `200 OK`

```json
{
  "user_id": "user-123",
  "vms": [
    {
      "id": "org-456-abc123",
      "status": "running",
      ...
    },
    {
      "id": "org-456-def456",
      "status": "stopped",
      ...
    }
  ],
  "count": 2
}
```

---

### POST /api/vms/{vm_id}/activity

Record user activity on a VM (prevents idle timeout).

**Response:** `200 OK`

```json
{
  "activity": "updated"
}
```

**Note:** Call this endpoint whenever the user interacts with the terminal. This resets the idle timer.

---

### POST /api/vms/{vm_id}/stop

Stop (but don't destroy) a VM.

**Response:** `200 OK`

```json
{
  "status": "stopped"
}
```

**Note:** Stopped VMs can be restarted later. Billing continues.

---

### DELETE /api/vms/{vm_id}

Destroy a VM and end billing.

**Response:** `200 OK`

```json
{
  "status": "destroyed"
}
```

**Behavior:**

1. Stops the Fly machine
2. Calculates billed minutes (creation time to destruction time)
3. Records usage and cost to metrics
4. Removes from local tracking

**Errors:**

- `404 Not Found` — VM does not exist
- `500 Internal Server Error` — Fly.io destruction failed

---

## Terminal Relay

### WS /api/terminal/connect

Establish WebSocket connection to a VM's terminal.

**Query Parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `vm_id` | ✅ | VM ID to connect to |
| `client_id` | ✅ | Unique client identifier (for tracking) |

**Example:**

```javascript
const ws = new WebSocket(
  'wss://vm-manager.example.com/api/terminal/connect?vm_id=org-456-abc123&client_id=user-123-session-xyz'
);

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  if (message.type === 'data') {
    // Terminal output
    console.log(message.data);
  } else if (message.type === 'pong') {
    // Heartbeat response
    console.log('Connection alive');
  }
};

ws.send(JSON.stringify({
  type: 'data',
  data: 'ls -la\n'
}));
```

**Message Types:**

#### Client → Server

```json
{
  "type": "data",
  "data": "command\n"
}
```

Terminal input data (keyboard input, control sequences).

```json
{
  "type": "control",
  "data": "\u001b[A"
}
```

Terminal control sequences (arrow keys, etc.).

```json
{
  "type": "ping",
  "meta": {}
}
```

Heartbeat to keep connection alive.

#### Server → Client

```json
{
  "type": "data_ack",
  "data": "Received: command"
}
```

Acknowledgment of received data.

```json
{
  "type": "pong",
  "meta": {
    "timestamp": 1711000000
  }
}
```

Heartbeat response.

**Errors:**

- `400 Bad Request` — Missing vm_id or client_id
- `503 Service Unavailable` — Server at connection capacity
- WebSocket 101 Switching Protocols on success

---

### GET /api/terminal/clients

List connected terminal clients.

**Query Parameters:**

| Parameter | Optional | Description |
|-----------|----------|-------------|
| `vm_id` | ❌ | Filter clients by VM (if not provided, lists all) |

**Response:** `200 OK`

```json
{
  "clients": [
    {
      "id": "user-123-session-xyz",
      "vm_id": "org-456-abc123",
      "created_at": "2026-03-21T15:34:18Z",
      "last_activity": "2026-03-21T15:35:00Z",
      "bytes_sent": 1024,
      "bytes_received": 2048
    }
  ],
  "count": 1
}
```

---

### DELETE /api/terminal/clients/{client_id}

Force disconnect a terminal client.

**Response:** `200 OK`

```json
{
  "status": "disconnected"
}
```

**Errors:**

- `404 Not Found` — Client does not exist

---

## Error Responses

### 400 Bad Request

Invalid input parameters.

```json
{
  "error": "missing required field: user_id"
}
```

### 404 Not Found

Resource does not exist.

```json
{
  "error": "vm org-456-abc123 not found"
}
```

### 500 Internal Server Error

Server-side error (usually Fly.io API).

```json
{
  "error": "failed to create fly machine: invalid api token"
}
```

### 503 Service Unavailable

Server at capacity or temporarily unavailable.

```json
{
  "error": "org org-456 has reached max VMs (100)"
}
```

---

## Rate Limiting

No explicit rate limiting is enforced. Implement at load balancer level:

```
- 100 requests/sec per IP
- 1000 WebSocket connections per client
- 10,000 total concurrent connections
```

---

## Examples

### Create VM and Connect Terminal (cURL + WebSocket)

```bash
# 1. Create VM
curl -X POST http://localhost:8080/api/vms \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "org_id": "org-456",
    "repo_url": "https://github.com/user/repo.git",
    "api_key": "sk-...",
    "cpus": 2,
    "memory_mb": 4096
  }'

# Response:
# {
#   "id": "org-456-abc123",
#   "terminal_url": "wss://abc123.fly.dev/terminal",
#   "status": "starting",
#   ...
# }

# 2. Wait for VM to start (check status)
sleep 5

# 3. Get VM status
curl http://localhost:8080/api/vms/org-456-abc123

# 4. Connect terminal via WebSocket
wscat -c "ws://localhost:8080/api/terminal/connect?vm_id=org-456-abc123&client_id=user-123-session-1"

# 5. Send command
{"type":"data","data":"ls -la\n"}

# 6. Record activity (keep-alive)
curl -X POST http://localhost:8080/api/vms/org-456-abc123/activity

# 7. Destroy when done
curl -X DELETE http://localhost:8080/api/vms/org-456-abc123
```

### Node.js Client

```javascript
const http = require('http');

// Create VM
const options = {
  hostname: 'localhost',
  port: 8080,
  path: '/api/vms',
  method: 'POST',
  headers: { 'Content-Type': 'application/json' }
};

const req = http.request(options, (res) => {
  let data = '';
  res.on('data', chunk => data += chunk);
  res.on('end', () => {
    const vm = JSON.parse(data);
    console.log('VM created:', vm.id);
    
    // Connect terminal
    const WebSocket = require('ws');
    const ws = new WebSocket(
      `ws://localhost:8080/api/terminal/connect?vm_id=${vm.id}&client_id=session-1`
    );
    
    ws.on('open', () => {
      ws.send(JSON.stringify({
        type: 'data',
        data: 'echo "Hello from Node.js"\n'
      }));
    });
    
    ws.on('message', (data) => {
      console.log('Terminal:', JSON.parse(data).data);
    });
  });
});

req.write(JSON.stringify({
  user_id: 'user-123',
  org_id: 'org-456',
  repo_url: 'https://github.com/user/repo.git',
  api_key: 'sk-...'
}));
req.end();
```

---

## Best Practices

1. **Always record activity** when user interacts with terminal
2. **Gracefully handle disconnections** — reconnect with same vm_id
3. **Destroy VMs** when done to stop billing
4. **Monitor metrics** to track costs and usage
5. **Set appropriate region** based on user location
6. **Use reasonable CPU/memory** — start with 2CPU/4GB, scale as needed

---

## Changelog

### v1.0.0 (2026-03-21)

- Initial release
- VM lifecycle management
- Terminal WebSocket relay
- Cost tracking
- Graceful shutdown
