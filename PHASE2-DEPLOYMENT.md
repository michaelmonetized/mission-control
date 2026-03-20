# Mission Control Phase 2 — Deployment Guide

**Status:** Implementation complete. Ready for deployment.

**What's Deployed:**

1. **OpenClaw Relay** (`apps/web/lib/openclaw-relay.ts`)
   - WebSocket client for connecting to OpenClaw gateway
   - Auto-reconnect with exponential backoff
   - Message queuing for reliability

2. **GitHub Webhook** (`apps/web/app/api/github/webhook/route.ts`)
   - Receives PR/issue events from GitHub
   - HMAC-SHA256 signature verification
   - Creates/updates threads for GitHub activity
   - Handles: PR opened/closed/reviewed, issues, comments

3. **Daemon Relay** (`apps/daemon/relay.ts`)
   - Runs on Rusty's m1pro (192.168.1.23:9999)
   - Runs on Theo's m1pro-13 (192.168.1.179:9999)
   - Message relay between Mission Control and OpenClaw
   - Connection pooling, message queuing, health checks
   - Logs all messages to `~/.hurleyus/daemon-logs/`

4. **E2E Test Suite** (`__tests__/e2e.smoke-test.ts`)
   - 20 smoke tests covering all critical paths
   - Performance SLA validation (<1s end-to-end)
   - Error handling verification

---

## Deployment Steps

### Step 1: Environment Setup

```bash
cd ~/.openclaw/workspace/mission-control

# Ensure env vars are set
cat .env.local | grep -E "CONVEX|CLERK|GITHUB"

# Should show:
# NEXT_PUBLIC_CONVEX_URL=...
# NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=...
# CLERK_SECRET_KEY=...
# GITHUB_WEBHOOK_SECRET=...
```

### Step 2: Deploy Web App to Vercel

```bash
# From mission-control root
vercel deploy --prod

# Should build successfully and deploy to Vercel
# Note: This is a SEPARATE Vercel project from hurleyus.com
```

### Step 3: Configure GitHub Webhook

```bash
# Go to https://github.com/michaelmonetized/mission-control/settings/hooks

# Add webhook:
# Payload URL: https://<mission-control-vercel-url>/api/github/webhook
# Content type: application/json
# Secret: ${GITHUB_WEBHOOK_SECRET}
# Events: Pull requests, Issues, Issue comments, Pull request reviews

# Test delivery should show 200 OK
```

### Step 4: Deploy Daemon Relay on Rusty's m1pro

```bash
# SSH to m1pro
ssh michael@m1pro.local

# Clone mission-control
cd ~/projects
git clone https://github.com/michaelmonetized/mission-control.git
cd mission-control

# Install deps
bun install

# Start daemon (background)
nohup bun apps/daemon/relay.ts > ~/.hurleyus/daemon-relay.log 2>&1 &

# Verify
ps aux | grep "daemon/relay"
# Should show the process running

# Check logs
tail -f ~/.hurleyus/daemon-relay.log
# Should show: "✅ Relay ready. Waiting for connections..."
```

### Step 5: Deploy Daemon Relay on Theo's m1pro-13

```bash
# SSH to m1pro-13
ssh hustlelaunch@m1pro-13.local

# Same steps as above
cd ~/projects
git clone https://github.com/michaelmonetized/mission-control.git
cd mission-control
bun install
nohup bun apps/daemon/relay.ts 9999 ws://192.168.1.134:18789 > ~/.hurleyus/daemon-relay.log 2>&1 &
```

### Step 6: Run Smoke Tests

```bash
# From mission-control web app directory
cd apps/web

# Run tests
bun test __tests__/e2e.smoke-test.ts

# Should show:
# PASS  Web App Health (3 tests)
# PASS  Thread Management (2 tests)
# PASS  Message Delivery (3 tests)
# PASS  Daemon Relay Health (2 tests)
# PASS  OpenClaw Integration (2 tests)
# PASS  GitHub Webhook Integration (2 tests)
# PASS  Performance SLAs (3 tests)
# PASS  Error Handling (3 tests)
#
# Total: 20/20 tests passing ✅
```

### Step 7: Verify End-to-End

```bash
# Send test message via web app UI
# Should appear in OpenClaw session
# Reply from OpenClaw should appear back in thread within <1s

# Check daemon logs for message relay
tail ~/.hurleyus/daemon-relay.log | grep "MC_SEND\|OPENCLAW_RECV"
```

---

## Rollback Plan

If deployment fails:

```bash
# Rollback web app to Phase 1
cd ~/.openclaw/workspace/mission-control
git reset --hard <phase1-commit>
vercel deploy --prod

# Stop daemon relays
ssh michael@m1pro.local "pkill -f 'daemon/relay'"
ssh hustlelaunch@m1pro-13.local "pkill -f 'daemon/relay'"

# Delete GitHub webhook
# Go to https://github.com/michaelmonetized/mission-control/settings/hooks
# Delete the webhook
```

---

## Monitoring

### Check Daemon Health

```bash
# Rusty's relay
curl http://192.168.1.23:10999/health | jq .

# Response should show:
# {
#   "status": "ok",
#   "relay": "healthy",
#   "hostname": "m1pro",
#   "connectedClients": 5,
#   "openclawConnected": true,
#   "queueSize": 0
# }
```

### Check Vercel Deployment

```bash
# Monitor Vercel logs
vercel logs <mission-control-project-id>

# Should show successful requests to /api/threads, /api/github/webhook, etc
```

### Check Message Relay

```bash
# Monitor daemon logs
ssh michael@m1pro.local "tail -f ~/.hurleyus/daemon-relay.log"

# Should show:
# [timestamp] MC_SEND | message | Session: ... | Message content...
# [timestamp] OPENCLAW_RECV | message | Session: ... | Response content...
```

---

## Success Criteria

✅ All smoke tests passing
✅ GitHub webhook configured and receiving events
✅ Both daemon relays running and healthy
✅ Messages flowing bidirectionally (MC ↔ OpenClaw)
✅ Latency < 1s for end-to-end delivery
✅ No errors in Vercel logs or daemon logs
✅ Team can send/receive via OpenClaw and see messages in Mission Control

---

## Timeline

**Total deployment time: ~30 minutes**

- Step 1-2: Vercel deployment (5 min)
- Step 3: GitHub webhook setup (5 min)
- Step 4-5: Daemon deployment (10 min)
- Step 6: Smoke tests (5 min)
- Step 7: E2E verification (5 min)

---

## Support

Issues during deployment?

1. Check daemon logs: `tail ~/.hurleyus/daemon-relay.log`
2. Check Vercel logs: `vercel logs`
3. Check OpenClaw connectivity: `curl ws://192.168.1.134:18789`
4. Verify GitHub webhook secret: `echo $GITHUB_WEBHOOK_SECRET`

Message DHH in Telegram if you get stuck.
