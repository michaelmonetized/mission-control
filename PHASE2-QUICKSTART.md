# Mission Control Cloud — Phase 2 Quick Start

**Get the full stack running locally in 10 minutes.**

---

## Prerequisites

- Bun 1.0+ (`bun --version`)
- Go 1.21+ (`go version`)
- Node 20+ (`node --version`)
- Clerk account (GitHub OAuth)
- Stripe account (test mode)
- Fly.io account + CLI (`fly auth login`)

---

## Setup

### 1. Environment Variables

Copy `.env.local` template:

```bash
cp apps/web/.env.local.example apps/web/.env.local
```

Fill in:
```
CONVEX_DEPLOYMENT=your-convex-deployment-id
NEXT_PUBLIC_CONVEX_URL=https://your-deployment.convex.cloud
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_xxx
CLERK_SECRET_KEY=sk_test_xxx
STRIPE_SECRET_KEY=sk_test_xxx
FLY_API_TOKEN=your-fly-token
FLY_APP_NAME=mc-workspaces-dev
VM_MANAGER_API_KEY=dev-secret-key-12345
```

### 2. Install Dependencies

```bash
cd apps/web
bun install
cd ../../
```

### 3. Start Services (4 terminals)

**Terminal 1 — Convex (backend)**
```bash
cd apps/web
bun run convex:dev
# Output: Convex dev server running on http://localhost:8200
```

**Terminal 2 — Next.js (frontend)**
```bash
cd apps/web
bun run dev
# Output: http://localhost:3410
```

**Terminal 3 — WebSocket Relay (terminal streaming)**
```bash
bun services/websocket-relay/main.ts
# Output: WebSocket relay running on ws://localhost:9001
```

**Terminal 4 — VM Manager (Go service)**
```bash
cd services/vm-manager
go run main.go
# Output: VM Manager listening on :9000
```

---

## Test the Full Stack

### 1. Sign In

1. Open http://localhost:3410
2. Click "Sign In with GitHub"
3. Authorize Clerk
4. You're logged in!

### 2. Connect a Repository

1. Go to Dashboard
2. Click "Sync Repositories"
3. Click on a repo to connect
4. Status should show "connected"

### 3. Launch a Workspace

1. Click "Launch Workspace" on any repo
2. Status should change to "starting"
3. After ~30 seconds, it's "running"

### 4. Open Terminal

1. Once workspace is "running", the Terminal tab activates
2. Terminal connects via WebSocket
3. Type commands (they go to VM)
4. Output streams back in real-time

### 5. Check Billing

Dashboard shows:
- Minutes used (cumulative)
- Current cost ($)
- Free minutes remaining

---

## Run Tests

### Unit Tests
```bash
cd apps/web
bun test
```

### E2E Tests
```bash
TEST_URL=http://localhost:3410 bun test __tests__/phase2-e2e.test.ts
```

### Performance Tests
```bash
bun test __tests__/performance.test.ts
```

---

## Troubleshooting

**"Convex backend not responding"**
→ Check Terminal 1, restart: `bun run convex:dev`

**"Sign In button doesn't work"**
→ Verify `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` in `.env.local`

**"WebSocket connection refused"**
→ Check Terminal 3 is running: `bun services/websocket-relay/main.ts`

**"VM launch times out"**
→ Fly.io API token invalid, check `FLY_API_TOKEN`

**"Terminal shows no output"**
→ VM Manager not running (Terminal 4), start it

---

## File Structure

```
mission-control/
├── apps/
│   └── web/               # Next.js frontend
│       ├── app/           # Pages, layouts
│       ├── components/    # React components (Terminal, etc)
│       ├── convex/        # Backend schema, mutations, queries
│       ├── hooks/         # useTerminal, useWorkspaceStatus
│       └── lib/           # Utilities (Stripe, etc)
├── services/
│   ├── vm-manager/        # Go HTTP server (VM launch/stop)
│   └── websocket-relay/   # Bun WebSocket server (terminal relay)
└── PHASE2-*.md            # Documentation
```

---

## Next Steps

1. **Develop locally** — make changes, see them live
2. **Run tests** — `bun test` before committing
3. **Deploy to staging** — `vercel deploy`
4. **Deploy to production** — `vercel deploy --prod`

See `PHASE2-DEPLOYMENT.md` for full deployment guide.

---

## Architecture

```
Browser (http://localhost:3410)
  ↓
Next.js Frontend (port 3410)
  ↓
Convex Backend (port 8200)
  ↓
VM Manager Go Service (port 9000) ← Fly.io API
  ↓
WebSocket Relay (port 9001) ← Terminal (xterm.js)
  ↓
Stripe API (billing)
```

---

**All set?** You're ready to develop Phase 2. 🚀

Questions? Check logs in each terminal, or post to #incidents.
