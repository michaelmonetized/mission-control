# Mission Control — Phase 2: Cloud Platform

## Overview

Mission Control Cloud is a Next.js SaaS that brings the local TUI experience to the web. Users connect their GitHub repos, spin up isolated VPS VMs running Claude Code, and pay only for compute.

**Business Model:** BYO Claude Code subscription + pay-as-you-go compute

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Mission Control Cloud                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐            │
│  │   Next.js    │────▶│   Convex     │────▶│   GitHub     │            │
│  │   Frontend   │     │   Backend    │     │   OAuth      │            │
│  └──────────────┘     └──────────────┘     └──────────────┘            │
│         │                    │                                          │
│         │                    │                                          │
│         ▼                    ▼                                          │
│  ┌──────────────┐     ┌──────────────┐                                 │
│  │   Stripe     │     │   VM Pool    │                                 │
│  │   Billing    │     │   Manager    │                                 │
│  └──────────────┘     └──────────────┘                                 │
│                              │                                          │
│                              ▼                                          │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                    Isolated VPS VMs (per user)                    │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐                  │  │
│  │  │ Claude Code│  │ Claude Code│  │ Claude Code│  ...             │  │
│  │  │  + Repos   │  │  + Repos   │  │  + Repos   │                  │  │
│  │  └────────────┘  └────────────┘  └────────────┘                  │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Next.js 16+, React 19, Tailwind v4 |
| Backend | Convex (real-time, serverless) |
| Auth | Clerk (GitHub OAuth primary) |
| Payments | Stripe (usage-based billing) |
| VMs | Fly.io Machines or Hetzner Cloud API |
| Terminal | xterm.js + WebSocket relay |
| DNS | Cloudflare (*.missioncontrol.dev) |

---

## User Flow

### 1. Sign Up (GitHub Only)
```
User clicks "Sign in with GitHub"
  → Clerk OAuth flow
  → Request scopes: read:user, repo (private optional)
  → Create user in Convex
  → Redirect to dashboard
```

### 2. Connect Repos
```
Dashboard shows all user repos (public + private if granted)
  → User selects repos to connect
  → Store repo metadata in Convex
  → Clone to VM on first access
```

### 3. Connect Claude Code
```
User provides Claude API key (BYO)
  → Encrypted at rest in Convex
  → Injected into VM environment
  → User retains full control
```

### 4. Launch Workspace
```
User clicks "Open in Claude Code"
  → Spin up isolated Fly Machine
  → Mount repo (git clone --depth 1)
  → Start Claude Code CLI
  → Stream terminal via WebSocket
  → User interacts in browser
```

### 5. Billing
```
Compute time tracked per minute
  → Stripe usage-based billing
  → Monthly invoice
  → Free tier: 100 minutes/month
  → Pro: $0.02/min after free tier
```

---

## Data Model (Convex)

```typescript
// schema.ts
defineSchema({
  users: defineTable({
    clerkId: v.string(),
    githubId: v.string(),
    githubUsername: v.string(),
    claudeApiKey: v.optional(v.string()), // encrypted
    stripeCustomerId: v.optional(v.string()),
    freeMinutesUsed: v.number(),
    plan: v.union(v.literal("free"), v.literal("pro")),
  }).index("by_clerk", ["clerkId"]),

  repos: defineTable({
    userId: v.id("users"),
    githubId: v.number(),
    name: v.string(),
    fullName: v.string(),
    private: v.boolean(),
    defaultBranch: v.string(),
    cloneUrl: v.string(),
    lastSynced: v.optional(v.number()),
  }).index("by_user", ["userId"]),

  workspaces: defineTable({
    userId: v.id("users"),
    repoId: v.id("repos"),
    vmId: v.optional(v.string()), // Fly Machine ID
    status: v.union(
      v.literal("stopped"),
      v.literal("starting"),
      v.literal("running"),
      v.literal("stopping")
    ),
    startedAt: v.optional(v.number()),
    stoppedAt: v.optional(v.number()),
  }).index("by_user", ["userId"]),

  usageRecords: defineTable({
    userId: v.id("users"),
    workspaceId: v.id("workspaces"),
    minutes: v.number(),
    cost: v.number(),
    billingPeriod: v.string(), // "2026-02"
  }).index("by_user_period", ["userId", "billingPeriod"]),
});
```

---

## Security

### Isolation
- Each user gets dedicated Fly Machine
- VMs destroyed after 30 min idle
- No shared resources between users
- Network isolation via private networks

### Secrets
- Claude API keys encrypted with Convex encryption
- GitHub tokens stored in Clerk, never logged
- VM environment variables injected at runtime

### Compliance
- SOC 2 Type II (via Fly.io)
- GDPR ready (user data deletion)
- No training on user code

---

## Pricing

| Tier | Price | Includes |
|------|-------|----------|
| Free | $0 | 100 compute minutes/month |
| Pro | $20/month | 1000 minutes, then $0.02/min |
| Team | $50/user/month | Unlimited, shared workspaces |

**Note:** Claude API costs are user's responsibility (BYO subscription).

---

## MVP Scope (v0.1)

### Must Have
- [ ] GitHub OAuth sign-in (Clerk)
- [ ] List public repos
- [ ] Connect Claude API key
- [ ] Launch single workspace
- [ ] xterm.js terminal in browser
- [ ] Basic usage tracking
- [ ] Stripe checkout for Pro

### Nice to Have
- [ ] Private repos
- [ ] Multiple workspaces
- [ ] Workspace persistence
- [ ] Real-time collaboration
- [ ] Team accounts

### Out of Scope (v0.1)
- GitLab/Bitbucket
- Self-hosted GitHub Enterprise
- Custom VM images
- IDE integrations

---

## API Endpoints (Convex HTTP Actions)

```typescript
// /api/github/repos — List user repos
// /api/workspace/start — Spin up VM
// /api/workspace/stop — Tear down VM
// /api/workspace/connect — WebSocket upgrade
// /api/billing/usage — Current period usage
// /api/billing/checkout — Stripe checkout session
```

---

## VM Management (Fly.io)

```typescript
// lib/fly.ts
export async function createMachine(userId: string, repoUrl: string) {
  const machine = await fly.machines.create({
    app: "mission-control-vms",
    config: {
      image: "ghcr.io/hustleus/mc-workspace:latest",
      env: {
        REPO_URL: repoUrl,
        ANTHROPIC_API_KEY: "<injected>",
      },
      guest: {
        cpus: 2,
        memory_mb: 4096,
      },
      auto_destroy: true,
      restart: { policy: "no" },
    },
  });
  return machine.id;
}

export async function destroyMachine(machineId: string) {
  await fly.machines.destroy("mission-control-vms", machineId);
}
```

---

## Terminal Relay (WebSocket)

```typescript
// app/api/workspace/[id]/terminal/route.ts
export async function GET(req: Request, { params }) {
  const { id } = params;
  const workspace = await getWorkspace(id);
  
  // Proxy WebSocket to VM
  const vmWs = new WebSocket(`wss://${workspace.vmId}.fly.dev/terminal`);
  
  return new Response(null, {
    status: 101,
    webSocket: createProxyWebSocket(vmWs),
  });
}
```

---

## Phase 2 Milestones

| Milestone | Target | Status |
|-----------|--------|--------|
| M1: Auth + Repos | Week 1 | 🔲 |
| M2: VM Infra | Week 2 | 🔲 |
| M3: Terminal UI | Week 3 | 🔲 |
| M4: Billing | Week 4 | 🔲 |
| M5: Beta Launch | Week 5 | 🔲 |

---

## Phase 1 Prerequisites

Before starting Phase 2, Phase 1 must be complete:

- [x] TUI implementation (Go + Bubble Tea)
- [x] Project discovery (mc-discover)
- [x] Git status integration (mc-git-status)
- [x] GitHub integration (mc-gh-status)
- [x] Vercel integration (mc-vl-status)
- [x] Swift integration (mc-swift-status)
- [x] Aggregate stats (mc-stats)
- [x] Cache management (mc-cache)
- [x] Dev server management (mc-dev)
- [x] Caddy proxy config (mc-caddy)
- [x] Comprehensive test suite (18 tests)
- [ ] OpenClaw chat integration (stub exists)
- [ ] README updated to reflect Go implementation
- [ ] First public release (v1.0.0)

---

## Domain

**Primary:** missioncontrol.dev  
**Alt:** mc.hustlelaunch.com

---

## Open Questions

1. **VM Provider:** Fly.io vs Hetzner vs Railway?
2. **Persistence:** Ephemeral vs persistent workspaces?
3. **Collaboration:** Real-time shared terminals?
4. **Mobile:** Responsive web or native app?

---

*Phase 2 planning document — Last updated: 2026-02-15*
