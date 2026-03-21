# Phase 2: Mission Control Cloud — Technical Architecture

**Status:** APPROVED FOR IMPLEMENTATION  
**Started:** 2026-03-21 15:30 EST  
**Target Completion:** 2026-03-22 11:59 PM EST  

---

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    Mission Control Cloud                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  Next.js 16  │  │   Convex     │  │   Clerk      │           │
│  │  (Frontend)  │→ │  (Backend)   │→ │  (Auth)      │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│         │                  │                                      │
│         └──────────────────┼──────────────────┐                  │
│                            ▼                  ▼                   │
│                    ┌──────────────┐   ┌──────────────┐           │
│                    │  Stripe      │   │  GitHub      │           │
│                    │  (Billing)   │   │  (OAuth)     │           │
│                    └──────────────┘   └──────────────┘           │
│                            │                                      │
│                            ▼                                      │
│           ┌────────────────────────────────────┐                │
│           │     Fly.io Machines (VMs)          │                │
│           │  ┌──────────┐  ┌──────────┐        │                │
│           │  │ user-1   │  │ user-2   │  ...   │                │
│           │  │ Claude   │  │ Claude   │        │                │
│           │  │ Code +   │  │ Code +   │        │                │
│           │  │ Repos    │  │ Repos    │        │                │
│           │  └──────────┘  └──────────┘        │                │
│           └────────────────────────────────────┘                │
│                            ▲                                      │
│           ┌────────────────┼────────────────┐                   │
│           │                │                │                    │
│    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐            │
│    │ VM Manager   │ │ xterm.js     │ │ Cost         │            │
│    │ (Fly API)    │ │ (Browser     │ │ Tracker      │            │
│    │              │ │  Terminal)   │ │ (Usage API)  │            │
│    └──────────────┘ └──────────────┘ └──────────────┘            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Tech Stack (Final)

| Layer | Technology | Version | Notes |
|-------|-----------|---------|-------|
| **Frontend** | Next.js | 16.1 | React 19, SSR + RSC |
| **Framework** | React | 19 | Components |
| **Styling** | Tailwind CSS | v4 | UI + dark mode |
| **Components** | shadcn/ui | latest | Pre-built UI |
| **Auth** | Clerk | v6.3+ | GitHub OAuth primary |
| **Backend** | Convex | v1.15+ | Serverless, real-time |
| **Database** | Convex DB | native | NoSQL, built-in auth |
| **Billing** | Stripe | API v2024+ | Usage-based, metered |
| **VMs** | Fly.io Machines | stable | Isolated per user/repo |
| **Terminal** | xterm.js | v5+ | Browser terminal |
| **Relay** | Node.js | 20+ | WebSocket server |
| **Deployment** | Vercel | standard | Next.js hosting |
| **Storage** | Fly Volumes | stable | Persistent storage |

---

## Database Schema

**Convex Tables:**

### `users`
```typescript
{
  clerkId: string,          // Clerk user ID (primary)
  githubId: number,         // GitHub user ID
  githubUsername: string,
  email: string,
  avatar?: string,
  claudeApiKey?: string,    // Encrypted at rest
  stripeCustomerId?: string,
  freeMinutesUsed: number,  // Monthly tracking
  plan: "free" | "pro" | "team",
  billingPeriodStart?: timestamp,
  createdAt: timestamp,
  updatedAt: timestamp
}
```

**Indexes:**
- `by_clerk: clerkId` (auth lookup)
- `by_github: githubId` (sync source)

### `repos`
```typescript
{
  userId: id("users"),
  githubId: number,
  name: string,             // e.g., "mission-control"
  fullName: string,         // e.g., "michaelmonetized/mission-control"
  description?: string,
  private: boolean,
  defaultBranch: string,    // usually "main"
  cloneUrl: string,         // https://github.com/...
  htmlUrl: string,          // GitHub web URL
  lastSynced?: timestamp,
  syncedAt: timestamp
}
```

**Indexes:**
- `by_user: userId` (list user repos)
- `by_github_id: githubId` (dedup on sync)

### `workspaces`
```typescript
{
  userId: id("users"),
  repoId: id("repos"),
  vmId?: string,            // Fly Machine ID
  status: "stopped" | "starting" | "running" | "stopping" | "failed",
  startedAt?: timestamp,
  stoppedAt?: timestamp,
  createdAt: timestamp,
  updatedAt: timestamp
}
```

**Indexes:**
- `by_user: userId` (list user workspaces)
- `by_repo: repoId` (unique workspace per repo)
- `by_user_repo: [userId, repoId]` (lookup by both)

### `usageRecords`
```typescript
{
  userId: id("users"),
  workspaceId: id("workspaces"),
  durationMinutes: number,  // How long VM was running
  cost: number,             // Calculated at record time ($0.001 per min example)
  billingPeriod: string,    // "2026-03" for March 2026
  recordedAt: timestamp
}
```

**Indexes:**
- `by_user: userId` (user billing history)
- `by_user_period: [userId, billingPeriod]` (monthly usage)
- `by_workspace: workspaceId` (per-workspace tracking)

### `threads` & `messages`
(From HurleyUS integration — same as Phase 1)

---

## API Contract

### Mutations (Convex)

**User Setup:**
```typescript
createUserOnFirstLogin(clerkId: string, github: { id, username, email })
  → { userId: id, plan: "free" }

updateClaudeApiKey(userId: id, encryptedKey: string)
  → { success: true }

updateStripeCustomerId(userId: id, stripeCustomerId: string)
  → { success: true }
```

**Repository Management:**
```typescript
syncGitHubRepos(userId: id)
  → { synced: number, repos: [{ id, name, url }] }

connectRepo(userId: id, githubId: number)
  → { repoId: id, name: string, status: "ready" }

disconnectRepo(repoId: id)
  → { success: true }
```

**Workspace Lifecycle:**
```typescript
launchWorkspace(userId: id, repoId: id)
  → { workspaceId: id, vmId: string, status: "starting" }

stopWorkspace(workspaceId: id)
  → { success: true, status: "stopping" }

recordUsage(workspaceId: id, durationMinutes: number)
  → { cost: number, recorded: true }
```

**Billing:**
```typescript
getCurrentUsage(userId: id, billingPeriod: string)
  → { minutes: number, cost: number, limit: number }

requestInvoice(userId: id)
  → { invoiceId: string, url: string }
```

### Queries (Convex)

```typescript
getUser(clerkId: string)
  → { userId: id, plan, freeMinutesUsed, ... }

listUserRepos(userId: id)
  → [{ repoId: id, name, status: "connected" }]

listUserWorkspaces(userId: id)
  → [{ workspaceId: id, repoName, status, startedAt }]

getWorkspaceStatus(workspaceId: id)
  → { status, vmId, uptime, cpuUsage, memoryUsage }

getCurrentUsage(userId: id)
  → { minutes: number, cost: number, remaining: number }
```

### Real-Time Subscriptions

```typescript
subscribeToWorkspaceStatus(workspaceId: id)
  → Stream<{ status, uptime, cpu, memory, lastUpdate }>

subscribeToUsage(userId: id)
  → Stream<{ minutes, cost, period }>
```

---

## Security Model

### Authentication
- **Provider:** Clerk (GitHub OAuth)
- **Scope:** `read:user` (public profile) + optional `repo` (private)
- **JWT:** Convex validates Clerk JWT on every request
- **Session:** 24-hour expiry, refresh automatic

### Authorization
- **User Isolation:** Every query/mutation checks `ctx.userId`
- **Convex RLS:** Database-level row-level security
- **No Cross-User Access:** User A cannot see User B's repos/workspaces/usage

### Secrets Management
- **Claude API Key:** Encrypted at rest in Convex (no plaintext in DB)
- **Injection:** Decrypted only when launching VM environment
- **Rotation:** User can update key anytime
- **Audit Trail:** `webhookEvents` table logs all changes

### Rate Limiting
- **Per-User:** 100 workspace launches per day
- **Per-Workspace:** 1 launch per 5 minutes (prevent spam)
- **API:** Convex rate limits (default: 100 reqs/sec)

### Cost Control
- **Free Tier:** 100 minutes/month (auto-kill VMs at limit)
- **Metered Billing:** $0.001 per minute (configurable)
- **Monthly Cap:** Optional spend cap (default: none)
- **Alerts:** Notify user when usage >75% of budget

---

## VM Lifecycle (Fly.io)

### Launch
```
User clicks "Open in Claude Code"
  → Convex: create workspace (status=starting)
  → VM Manager: POST /launch { userId, repoId, branch }
    → Fly API: create machine (region: SFO, 1 CPU, 1GB RAM)
    → Git clone repo (--depth 1 for speed)
    → Inject Claude API key into environment
    → Start Claude Code CLI
  → WebSocket: stream terminal output to browser
  → Convex: update workspace (status=running, vmId=...)
```

### Running
```
Terminal input (browser) → WebSocket relay → VM stdin
  → User types commands (e.g., "cargo build")
  → VM executes
  → stdout/stderr → WebSocket relay → browser terminal
  → Convex: record usage (1 minute elapsed)
```

### Stop
```
User clicks "Stop"
  OR VM idle >2h (auto-shutdown)
  → Convex: update workspace (status=stopping)
  → VM Manager: POST /stop { vmId }
    → Fly API: destroy machine
    → Cleanup volumes
  → Convex: record final usage
  → Stripe: charge for time used
  → Convex: update workspace (status=stopped)
```

### Error Handling
```
VM spawn fails:
  → Convex: status=failed, lastError="..."
  → UI: Show error + retry button
  → User can retry immediately

Terminal disconnects:
  → WebSocket auto-reconnect (3x with exponential backoff)
  → Resume session if VM still running
  → Kill session if VM is gone
```

---

## Deployment

### Development
```bash
# Frontend
cd apps/web
bun install
bun run convex:dev    # Start local Convex backend
bun run dev           # Start Next.js dev server
# → http://localhost:3410

# Fly.io VM Manager (separate Go service)
cd services/vm-manager
go run main.go        # Starts on localhost:9000
```

### Staging
```bash
# Deploy Convex to staging
bun run convex:deploy --env staging

# Deploy Next.js to Vercel preview
vercel deploy --prod
```

### Production
```bash
# 1. Deploy Convex (production database)
convex deploy --prod

# 2. Deploy Next.js to Vercel (production)
vercel deploy --prod

# 3. Verify Clerk + Stripe webhooks are live

# 4. Run smoke tests
bun test:e2e

# 5. Monitor logs
vercel logs -n 100
```

---

## Success Criteria (Phase 2)

✅ **Technical**
- All mutations tested + working
- All queries returning correct data
- Real-time subscriptions streaming
- WebSocket relay <200ms latency
- VM spinup <2 seconds
- Auth working (GitHub OAuth → Clerk)
- Billing tracking (usage records flowing)
- Cost calculation correct

✅ **Operational**
- Staged to production
- Clerk webhooks configured
- Stripe usage API integrated
- Fly.io machines launching consistently
- Terminal relay stable (100+ concurrent)
- Error logging (Sentry) active
- Health checks passing

✅ **User-Facing**
- Dashboard shows repos
- "Launch Workspace" button works
- Terminal renders in browser
- Chat integration working (threads)
- Usage displayed (minutes/cost)
- Billing page shows charges

---

## Timeline (48 Hours)

| Time | Task | Owner |
|------|------|-------|
| Sat 15:30 | Architecture approved | DHH |
| Sat 16:30 | Convex mutations live | sr-engineer-backend |
| Sat 17:00 | Next.js frontend skeleton | sr-engineer-frontend |
| Sat 18:00 | Fly.io VM manager complete | sr-engineer-infra |
| Sat 19:00 | Integration testing | QA |
| Sat 20:00 | Staging deployment | DevOps |
| Sat 22:00 | E2E smoke tests passing | QA |
| Sun 00:00 | Production deployment | DevOps |
| Sun 06:00 | Monitoring + alerts live | Operations |
| Sun 11:59 | PHASE 2 COMPLETE | All |

---

## Blockers & Risks

**None.** All dependencies are external services (Fly.io, Stripe, GitHub). No internal blockers.

If Fly.io API is down, use Hetzner Cloud API as fallback.
If Stripe is down, queue usage records locally and sync on recovery.

---

## Post-Launch (Phase 3+)

- [ ] Team collaboration (multi-user workspaces)
- [ ] Deploy preview links
- [ ] SSH into VM from CLI
- [ ] Persistent volumes (save work between sessions)
- [ ] Cost insights (usage graphs, trends)
- [ ] Custom VM specs (2+ CPU, 4GB RAM options)
- [ ] Auto-scaling (priority queue for launches)
- [ ] Marketplace (sell compute to others?)

---

**Approved. Agents: proceed with implementation.**

Launch: Sunday 11:59 PM EST. 🚀
