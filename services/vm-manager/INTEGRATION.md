# VM Manager Integration Guide

How to integrate the VM Manager service with your application.

## Architecture Overview

```
┌──────────────────────────────────────────────┐
│    Your Next.js App (Mission Control Web)   │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │  Route: /api/workspace/start          │ │
│  │  - Call VM Manager POST /api/vms      │ │
│  │  - Get VM ID and terminal_url         │ │
│  │  - Return to frontend                 │ │
│  └────────────────────────────────────────┘ │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │  Component: TerminalWindow             │ │
│  │  - Connects WebSocket to terminal_url │ │
│  │  - Sends/receives terminal data       │ │
│  │  - Periodically posts /activity       │ │
│  └────────────────────────────────────────┘ │
└────────────────┬─────────────────────────────┘
                 │ HTTP + WebSocket
                 ▼
┌──────────────────────────────────────────────┐
│       VM Manager (this service)              │
│       :8080                                  │
├──────────────────────────────────────────────┤
│  POST /api/vms                               │
│  GET  /api/vms/{vm_id}                      │
│  WS   /api/terminal/connect                 │
│  POST /api/vms/{vm_id}/activity             │
│  DELETE /api/vms/{vm_id}                    │
└─────────────────┬────────────────────────────┘
                  │ HTTPS
                  ▼
        ┌─────────────────────────┐
        │  Fly.io Machines API    │
        │  api.machines.dev       │
        └─────────────────────────┘
```

## Integration Steps

### 1. Setup VM Manager Service

Ensure VM Manager is running:

```bash
# Local development
go run main.go

# Or Docker
docker run -p 8080:8080 -e FLY_API_TOKEN="..." mission-control-vm-manager
```

Available at: `http://localhost:8080` (or your deployment URL)

### 2. Backend Integration (Next.js API Route)

Create `app/api/workspace/start/route.ts`:

```typescript
import { NextRequest, NextResponse } from 'next/server';

const VM_MANAGER_URL = process.env.VM_MANAGER_URL || 'http://localhost:8080';

export async function POST(req: NextRequest) {
  const { repoUrl, repoRef = 'main', claudeApiKey, cpus = 2, memoryMb = 4096 } = await req.json();
  
  // Get user from Clerk
  const userId = req.headers.get('x-clerk-user-id');
  if (!userId) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    // Call VM Manager
    const response = await fetch(`${VM_MANAGER_URL}/api/vms`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        user_id: userId,
        org_id: 'org-123', // Get from Convex
        repo_url: repoUrl,
        repo_ref: repoRef,
        api_key: claudeApiKey,
        region: 'ord',
        cpus,
        memory_mb: memoryMb,
      }),
    });

    if (!response.ok) {
      const error = await response.text();
      return NextResponse.json({ error }, { status: response.status });
    }

    const vm = await response.json();

    // Save to Convex
    // await db.workspaces.create({
    //   userId,
    //   repoId: '...',
    //   vmId: vm.id,
    //   status: 'starting',
    //   startedAt: Date.now(),
    // });

    return NextResponse.json({
      vmId: vm.id,
      terminalUrl: vm.terminal_url,
      status: vm.status,
      createdAt: vm.created_at,
    });
  } catch (error) {
    console.error('Failed to create VM:', error);
    return NextResponse.json(
      { error: 'Failed to create workspace' },
      { status: 500 }
    );
  }
}
```

### 3. Frontend Integration (React Component)

Create `components/TerminalWindow.tsx`:

```typescript
'use client';

import { useEffect, useRef, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';

interface TerminalWindowProps {
  vmId: string;
  terminalUrl: string;
}

export default function TerminalWindow({ vmId, terminalUrl }: TerminalWindowProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstance = useRef<Terminal | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const activityTimerRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    if (!terminalRef.current) return;

    // Initialize xterm
    const term = new Terminal({
      cols: 120,
      rows: 40,
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
      },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);
    fitAddon.fit();

    terminalInstance.current = term;

    // Connect WebSocket
    const wsURL = new URL(terminalUrl);
    wsURL.searchParams.set('vm_id', vmId);
    wsURL.searchParams.set('client_id', `user-${Date.now()}`);

    const ws = new WebSocket(wsURL.toString());

    ws.onopen = () => {
      term.write('✅ Connected to workspace\r\n');
      term.focus();
      startActivityTracking();
    };

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      
      if (message.type === 'data' || message.type === 'data_ack') {
        term.write(message.data);
      } else if (message.type === 'pong') {
        // Heartbeat response
        console.log('Keep-alive');
      }
    };

    ws.onerror = (error) => {
      term.write('\r\n❌ Connection error\r\n');
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      term.write('\r\n⏹️  Connection closed\r\n');
      term.write('Click to reconnect\r\n');
    };

    // Handle terminal input
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(
          JSON.stringify({
            type: 'data',
            data,
          })
        );
      }
    });

    wsRef.current = ws;

    // Handle resize
    const handleResize = () => fitAddon.fit();
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      if (activityTimerRef.current) clearInterval(activityTimerRef.current);
      ws.close();
      term.dispose();
    };
  }, [vmId, terminalUrl]);

  const startActivityTracking = () => {
    // POST /api/vms/{vmId}/activity every 30 seconds
    activityTimerRef.current = setInterval(async () => {
      try {
        await fetch(`/api/workspace/${vmId}/activity`, { method: 'POST' });
      } catch (error) {
        console.error('Failed to record activity:', error);
      }
    }, 30000);
  };

  return (
    <div
      ref={terminalRef}
      style={{
        width: '100%',
        height: '100vh',
        backgroundColor: '#1e1e1e',
      }}
    />
  );
}
```

### 4. Workspace Page

Create `app/workspace/[id]/page.tsx`:

```typescript
'use client';

import { useEffect, useState } from 'react';
import TerminalWindow from '@/components/TerminalWindow';

export default function WorkspacePage({ params }: { params: { id: string } }) {
  const [vm, setVm] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchVM = async () => {
      try {
        // Get VM details from your backend
        const response = await fetch(`/api/workspace/${params.id}`);
        if (!response.ok) throw new Error('Failed to fetch workspace');
        
        const data = await response.json();
        setVm(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Error loading workspace');
      } finally {
        setLoading(false);
      }
    };

    fetchVM();
  }, [params.id]);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  if (!vm) return <div>Workspace not found</div>;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      <header style={{ padding: '1rem', backgroundColor: '#2d2d2d', color: '#fff' }}>
        <h1>{vm.repoName}</h1>
        <p>VM: {vm.vmId}</p>
      </header>
      
      <TerminalWindow vmId={vm.vmId} terminalUrl={vm.terminalUrl} />
    </div>
  );
}
```

### 5. Cleanup on Unmount

Add cleanup when user leaves workspace:

```typescript
useEffect(() => {
  return () => {
    // Cleanup: stop VM
    fetch(`/api/workspace/${params.id}`, { method: 'DELETE' });
  };
}, [params.id]);
```

## Convex Integration (Database)

Store VM state in Convex:

```typescript
// schema.ts
defineSchema({
  workspaces: defineTable({
    userId: v.id("users"),
    vmId: v.string(),
    repoUrl: v.string(),
    status: v.union(
      v.literal("starting"),
      v.literal("running"),
      v.literal("stopping"),
      v.literal("stopped")
    ),
    terminalUrl: v.string(),
    startedAt: v.number(),
    stoppedAt: v.optional(v.number()),
  }).index("by_user", ["userId"]),
});

// mutations.ts
export const createWorkspace = mutation({
  args: {
    vmId: v.string(),
    repoUrl: v.string(),
    terminalUrl: v.string(),
  },
  handler: async (ctx, args) => {
    const userId = await requireAuth(ctx);
    return await ctx.db.insert("workspaces", {
      userId,
      vmId: args.vmId,
      repoUrl: args.repoUrl,
      status: "starting",
      terminalUrl: args.terminalUrl,
      startedAt: Date.now(),
    });
  },
});

export const destroyWorkspace = mutation({
  args: {
    workspaceId: v.id("workspaces"),
  },
  handler: async (ctx, args) => {
    await ctx.db.delete(args.workspaceId);
  },
});

// queries.ts
export const getWorkspaces = query({
  args: {},
  handler: async (ctx) => {
    const userId = await requireAuth(ctx);
    return await ctx.db
      .query("workspaces")
      .withIndex("by_user", (q) => q.eq("userId", userId))
      .collect();
  },
});
```

## Environment Variables

Set these in your `.env.local`:

```bash
# VM Manager location
VM_MANAGER_URL=http://localhost:8080
# Or in production:
# VM_MANAGER_URL=https://mission-control-vm-manager.fly.dev

# Clerk
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=...
CLERK_SECRET_KEY=...

# Convex
NEXT_PUBLIC_CONVEX_URL=...
CONVEX_DEPLOYMENT=...
```

## Cost Tracking

Monitor billing through the metrics endpoint:

```typescript
async function getUserBilling(userId: string) {
  const response = await fetch(
    `${VM_MANAGER_URL}/metrics`
  );
  
  const metrics = await response.json();
  
  // Calculate user's total cost from usage_by_org
  let totalCost = 0;
  for (const [org, usages] of Object.entries(metrics.usage_by_org)) {
    for (const usage of usages as any[]) {
      if (usage.user_id === userId) {
        totalCost += usage.cost;
      }
    }
  }
  
  return totalCost;
}
```

## Error Handling

Handle common scenarios:

```typescript
// VM creation failed
if (response.status === 503) {
  // Org at capacity
  showError('All workspaces in your org are in use. Please stop one first.');
}

if (response.status === 500) {
  // Fly.io API error
  showError('Failed to create workspace. Please try again.');
}

// WebSocket connection failed
ws.onerror = () => {
  showError('Terminal connection lost. Reconnecting...');
  // Implement reconnect logic
};

// Activity tracking failed
if (activityResponse.status === 404) {
  // VM no longer exists, redirect user
  window.location.href = '/dashboard';
}
```

## Testing

Test the integration locally:

```bash
# Terminal 1: Start VM Manager
cd services/vm-manager
go run main.go

# Terminal 2: Start Next.js dev server
cd apps/web
npm run dev

# Terminal 3: Create workspace manually
curl -X POST http://localhost:3000/api/workspace/start \
  -H "Content-Type: application/json" \
  -H "x-clerk-user-id: user-123" \
  -d '{
    "repoUrl": "https://github.com/test/repo.git",
    "claudeApiKey": "sk-..."
  }'

# Then navigate to the returned vmId in browser
```

## Production Considerations

1. **Authenticate VM Manager calls** — Use API key or JWT
2. **Rate limit** — Prevent abuse (e.g., max 10 VMs per user)
3. **Cost limits** — Set monthly spending caps per user/org
4. **Monitoring** — Track metrics, alert on errors
5. **Backup** — Regularly backup Convex database
6. **Scaling** — Use load balancer if running multiple instances
7. **DNS** — Point custom domain to VM Manager service

## Support

See:
- `README.md` — Overview
- `API.md` — Complete API reference
- `DEPLOYMENT.md` — Deployment instructions
