# Deployment Checklist — Phase 2 Mission Control Cloud

**Target:** Live by Sunday 11:59 PM EST

---

## PRE-DEPLOYMENT (Check these now)

- [ ] All tests passing locally
  ```bash
  bun test __tests__/phase2-e2e.test.ts
  ```

- [ ] Git tree clean
  ```bash
  git status
  ```

- [ ] Latest code committed
  ```bash
  git log --oneline -3
  ```

- [ ] Environment variables documented
  - [ ] CONVEX_DEPLOYMENT
  - [ ] NEXT_PUBLIC_CONVEX_URL
  - [ ] CLERK keys (pub + secret)
  - [ ] STRIPE keys (pub + secret)
  - [ ] FLY_API_TOKEN
  - [ ] VM_MANAGER_API_KEY

---

## STAGING DEPLOYMENT (1 hour)

### Deploy Convex (Staging)
- [ ] Run: `convex deploy --env staging`
- [ ] Verify: tables exist (`users`, `repos`, `workspaces`, etc)
- [ ] Test: mutations working (createUserOnFirstLogin, etc)

### Deploy Next.js (Vercel Preview)
- [ ] Run: `vercel deploy`
- [ ] Wait for build (~3 min)
- [ ] Verify URL works: https://[preview-url].vercel.app
- [ ] Test sign-in: "Sign In with GitHub" button
- [ ] Test dashboard loads

### Deploy WebSocket Relay (Temporary)
- [ ] Start locally: `bun services/websocket-relay/main.ts`
- [ ] Verify: `curl http://localhost:9002/health` returns 200

### Deploy VM Manager (Temporary)
- [ ] Start locally: `cd services/vm-manager && go run main.go`
- [ ] Verify: `curl http://localhost:9000/health` returns 200

### Smoke Test
- [ ] Go to staging dashboard
- [ ] Sign in with GitHub
- [ ] Click "Sync Repositories"
- [ ] Choose a repo
- [ ] Click "Connect Repo"
- [ ] Click "Launch Workspace"
- [ ] Wait for status → "running"
- [ ] Terminal tab activates
- [ ] Type in terminal (test input)
- [ ] Check usage displayed

### Verify Integrations
- [ ] Stripe webhook received (check Stripe dashboard)
- [ ] Fly.io machine created (check Fly dashboard)
- [ ] Clerk session created (check Clerk dashboard)
- [ ] Convex mutations executed (check logs)

### Sign-Off
- [ ] All smoke tests passed: ✅
- [ ] No errors in logs: ✅
- [ ] Ready for production: ✅

---

## PRODUCTION DEPLOYMENT (1 hour)

### Create Release Tag
- [ ] Run: `git tag v2.0.0`
- [ ] Run: `git push origin v2.0.0`

### Deploy Convex (Production)
- [ ] Run: `convex deploy --prod`
- [ ] Wait for deployment
- [ ] Verify: production database live

### Deploy Next.js (Production)
- [ ] Run: `vercel deploy --prod`
- [ ] Wait for build
- [ ] Verify: https://mission-control.vercel.app works
- [ ] Check Vercel analytics
- [ ] Test sign-in flow

### Deploy WebSocket Relay (Production)
- [ ] Run: `fly deploy --app mc-websocket-relay-prod`
- [ ] Verify: `curl https://mc-websocket-relay-prod.fly.dev:9002/health`

### Deploy VM Manager (Production)
- [ ] Run: `cd services/vm-manager && fly deploy --app mc-vm-manager-prod`
- [ ] Verify: `curl https://mc-vm-manager-prod.fly.dev:9000/health`

### Verify All Services
- [ ] Vercel: https://mission-control.vercel.app/api/health → 200
- [ ] WebSocket: `curl https://mc-websocket-relay-prod.fly.dev:9002/health` → 200
- [ ] VM Manager: `curl https://mc-vm-manager-prod.fly.dev:9000/health` → 200
- [ ] Convex: Console shows live data

### Production Smoke Test
- [ ] Open: https://mission-control.vercel.app
- [ ] Sign in with GitHub
- [ ] Complete full flow (connect repo → launch → terminal)
- [ ] Check billing displays correctly
- [ ] All features working

### Monitor & Verify
- [ ] Watch logs for errors: `vercel logs -n 100 --follow`
- [ ] Check error rate (<0.1%): Sentry dashboard
- [ ] Check performance: P50, P99 latencies
- [ ] Check WebSocket connections: Fly logs
- [ ] 30+ minutes of monitoring

### Production Sign-Off
- [ ] All health checks green: ✅
- [ ] No errors in logs: ✅
- [ ] Performance targets met: ✅
- [ ] Ready for public: ✅

---

## ANNOUNCEMENT (30 min)

- [ ] Write Twitter/X post
- [ ] Write GitHub releases (v2.0.0)
- [ ] Send Telegram to team
- [ ] Verify links are live
- [ ] Post all channels

### Social Media
- [ ] Twitter announcement
- [ ] GitHub releases
- [ ] Telegram team

### Internal
- [ ] Team Telegram group
- [ ] Slack #announcements (if exists)
- [ ] Dashboard/status page update

### Public (if applicable)
- [ ] Website update
- [ ] Blog post (optional)
- [ ] ProductHunt (optional)

---

## POST-LAUNCH (24 hours)

- [ ] Monitor error rate (<0.1%)
- [ ] Monitor latency (P99 <2s)
- [ ] Monitor WebSocket uptime (99%+)
- [ ] Check Stripe billing flowing
- [ ] Respond to user feedback
- [ ] Fix any critical bugs (immediately)
- [ ] Document any issues

---

## ROLLBACK (If needed)

If production fails:

1. [ ] Stop deployment: `vercel rollback`
2. [ ] Revert Convex: previous backup
3. [ ] Revert Fly.io: previous image tag
4. [ ] Verify systems online
5. [ ] Analyze root cause
6. [ ] Document incident

**Target rollback time:** 10 minutes max

---

## Sign-Offs

- [ ] Architecture approved: **DHH**
- [ ] Code reviewed: **DHH**
- [ ] Tests passing: **QA**
- [ ] Deployment approved: **DHH**
- [ ] Go-live approved: **DHH**

---

## Timeline

| Step | Estimated Time | Actual Time |
|------|---|---|
| Pre-deployment | 30 min | __ |
| Staging deploy | 30 min | __ |
| Staging test | 30 min | __ |
| Production deploy | 60 min | __ |
| Monitoring | 30 min | __ |
| Announcement | 30 min | __ |
| **TOTAL** | **4 hours** | __ |

---

**Status:** Ready to deploy  
**Approval:** Awaiting signal  
**Timeline:** Sunday before midnight ✅

Send "GO" and we launch. 🚀
