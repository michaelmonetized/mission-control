# Mission Control — Location & Architecture

## ⚠️ CRITICAL: Mission Control is COMPLETELY SEPARATE from HurleyUS.com

---

## Canonical Repository

### **PRIMARY: /home/michael/.openclaw/workspace/mission-control**

**This is the authoritative location for all Mission Control work.**

```
~/.openclaw/workspace/mission-control/
├── apps/
│   ├── web/                    # Next.js 16 frontend
│   │   ├── app/               # Route handlers
│   │   ├── components/        # UI components
│   │   └── hooks/            # Convex integrations
│   └── daemon/               # Message relay service
├── convex/                    # Convex backend (queries, mutations)
├── package.json              # Root monorepo config
├── .env.local               # Convex deployment URL + Clerk keys
└── .git/                    # GitHub: michaelmonetized/mission-control
```

**GitHub Repo:** https://github.com/michaelmonetized/mission-control

---

## Secondary Locations (Clones/Mirrors)

### /home/michael/Projects/mission-control
- Mirror of ~/.openclaw/workspace/mission-control
- Keep in sync via git pull/push

### /home/michael/Projects/hurley-mission-control
⚠️ **DEPRECATED** — Do not use. Only ~one instance should exist. Consolidate to mission-control.

---

## ✅ What Mission Control Contains

- ✅ Thread list page (`/app/threads`)
- ✅ Thread detail page (`/app/threads/[id]`)
- ✅ Message input & display
- ✅ Real-time polling (2s interval)
- ✅ Convex backend (schema, mutations, queries)
- ✅ Clerk authentication
- ✅ Message relay daemon (connects to OpenClaw)

---

## ❌ What Mission Control Does NOT Touch

- ❌ HurleyUS.com (public marketing site)
- ❌ HurleyUS.com domain (https://www.hurleyus.com)
- ❌ Public-facing pages or routes
- ❌ iLeague, iTour, iCon product code
- ❌ Golf course partner integrations
- ❌ Brand/sponsor marketing features

---

## Deployment

### Mission Control (Private)
- **Vercel Project:** `hurley-mission-control` (separate from hurleyus)
- **URL:** Private/internal only (behind Clerk auth)
- **Database:** Convex (dev:accurate-goldfinch-601)
- **Auth:** Clerk JWT

### HurleyUS.com (Public)
- **Vercel Project:** `hurleyus` (clean of all mission control code as of 2026-03-20)
- **URL:** https://www.hurleyus.com (public)
- **Database:** ❌ No internal team features
- **Auth:** Optional (public access)

---

## For All Agents & Team Members

### When working on Mission Control:
```bash
cd ~/.openclaw/workspace/mission-control
# NOT ~/Projects/hurleyus.com
# NOT ~/Projects/hurley-mission-control
# NOT ~/Projects/mission-control (unless syncing)
```

### When committing:
```bash
git push origin main  # Always pushes to michaelmonetized/mission-control
```

### When deploying:
- Vercel project: **hurley-mission-control** (private)
- ❌ DO NOT deploy to hurleyus project
- ❌ DO NOT add routes to HurleyUS.com

---

## Cleanup History

**Date:** 2026-03-20 13:36 EDT

Mission Control code was accidentally merged into hurleyus.com repo. Rollback completed:

```bash
cd ~/Projects/hurleyus.com
git reset --hard 02a0f4d  # Removed all /app/threads/* code
git push origin main --force
```

**Removed commits:**
- 20f5fa3 [codex-dev] Add thread detail page with message list and input
- 7a40614 [codex-dev] Add real Convex integration + message polling with 2s interval
- 07ce062 [codex-dev] Phase 2-3: Add message polling + Convex integration
- 074de35 [codex-dev] Fix convex integration: add mock hooks + convex provider
- 403e6ee [codex-dev] Phase 1 complete: Add threads and messages system

---

## Summary

```
✅ HurleyUS.com = PUBLIC, marketing, golf courses & brands
✅ Mission Control = PRIVATE, internal, team communication

🔒 They are completely isolated.
🔒 They share nothing.
🔒 They never will.
```

---

**This document is canonical. Reference it whenever there's ambiguity about what belongs where.**
