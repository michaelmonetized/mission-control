# SHIP: Phase 2 Deployment — One-Click Launch

**Status:** READY TO DEPLOY  
**Commit:** 84c88e2 (all changes committed)  
**Buffer:** 20+ hours before deadline  

---

## 🚀 TO DEPLOY: Copy & Run

### Step 1: Verify (2 min)
```bash
cd ~/.openclaw/workspace/mission-control
git status                    # Should show clean tree
git log -1 --oneline         # Should show 84c88e2
```

### Step 2: Push to GitHub (1 min)
```bash
git push origin main
git push origin v2.0.0       # Tag if not already pushed
```

### Step 3: Run Smoke Test (30 min)
```bash
cd apps/web
bun test __tests__/phase2-e2e.test.ts
# Should show: 10/10 tests passing ✅
```

### Step 4: Deploy to Production (1 hour)

**Terminal 1 — Deploy Convex:**
```bash
cd apps/web
convex deploy --prod
# Wait for "Deployment complete"
```

**Terminal 2 — Deploy Next.js:**
```bash
vercel deploy --prod
# Wait for build, note the URL
```

**Terminal 3 — Deploy VM Manager:**
```bash
cd services/vm-manager
fly deploy --app mc-vm-manager-prod
# Or: fly deploy (if fly.toml configured)
```

**Terminal 4 — Deploy WebSocket Relay:**
```bash
cd services/websocket-relay
fly deploy --app mc-websocket-relay-prod
# Or: fly deploy (if fly.toml configured)
```

### Step 5: Verify All Services (10 min)

```bash
# Health checks
curl https://mission-control.vercel.app/api/health
# Should return: {"status":"ok"}

curl https://mc-websocket-relay-prod.fly.dev:9002/health
# Should return: {"status":"ok","service":"websocket-relay",...}

curl https://mc-vm-manager-prod.fly.dev:9000/health
# Should return: {"status":"ok","service":"vm-manager",...}
```

### Step 6: Smoke Test in Production (15 min)

1. Open https://mission-control.vercel.app
2. Sign in with GitHub
3. Click "Sync Repositories"
4. Select a repo → "Connect"
5. Click "Launch Workspace"
6. Wait for status → "running"
7. Type in terminal → should work
8. Check usage displays

**If all working:** ✅ Live!

### Step 7: Announce (10 min)

**Twitter/X:**
```
🚀 Mission Control Cloud is LIVE

Connect GitHub repos → launch isolated VPS workspaces → code with Claude in browser

Free tier: 100 min/month
$0.001/min after that

https://mission-control.vercel.app

#CodeWithoutLimits #AI #IDE
```

**GitHub Release:**
```bash
git tag v2.0.0 -m "Phase 2: Mission Control Cloud SaaS"
git push origin v2.0.0
# Go to GitHub → Releases → Create Release from tag
```

**Telegram:**
```
✅ Phase 2 LIVE: Mission Control Cloud
- Web IDE with Clerk + Convex + Fly.io
- Terminal in browser
- Real-time billing ($0.001/min)
- https://mission-control.vercel.app
```

### Step 8: Monitor (30 min)

```bash
# Watch logs
vercel logs -n 100 --follow

# Watch Fly.io
fly logs --app mc-websocket-relay-prod --follow
fly logs --app mc-vm-manager-prod --follow

# Check error rate (should be 0%)
# Check latency (should be <200ms)
```

---

## 📋 Checklist

- [ ] `git status` clean
- [ ] `git push origin main` done
- [ ] E2E tests passing
- [ ] Convex deployed
- [ ] Next.js deployed
- [ ] VM Manager deployed
- [ ] WebSocket relay deployed
- [ ] All 4 health checks passing
- [ ] Production smoke test working
- [ ] Twitter post sent
- [ ] GitHub release created
- [ ] Telegram announced
- [ ] 30+ min monitoring done
- [ ] No errors in logs

---

## 🚨 If Something Fails

**Convex won't deploy:**
→ Check API token: `echo $CONVEX_DEPLOY_KEY`

**Vercel build fails:**
→ Check build logs: `vercel logs -n 200`

**Fly.io deploy fails:**
→ Check `fly.toml` exists in service dirs
→ Run `fly auth login` if not authenticated

**Health check fails:**
→ Wait 60 seconds (services still booting)
→ Check Vercel/Fly logs for startup errors

**Smoke test fails:**
→ Check Clerk app is configured
→ Check GitHub OAuth settings
→ Check .env vars are correct

**Rollback (if needed):**
```bash
vercel rollback              # Revert to previous Vercel
convex deploy --prod --from-backup  # Restore Convex
```

---

## 📊 Success Metrics

After deployment, verify:
- [ ] Error rate: 0% (first 5 min), <0.1% (ongoing)
- [ ] Latency: P50 <100ms, P99 <500ms
- [ ] WebSocket uptime: 99%+
- [ ] VM spinup: <2s
- [ ] Users signing in successfully
- [ ] Workspaces launching
- [ ] Terminal I/O working
- [ ] Billing tracking

---

## 🎉 Success = DONE

When all checks pass and smoke test works:

**Phase 2 is LIVE on production.** 🚀

Next: Phase 3 (team collaboration, SSH integration)

---

**Time to deploy:** ~3 hours  
**Estimated go-live:** Saturday evening OR Sunday morning (your choice)  
**Deadline buffer:** 17+ hours  

Ready when you are. 💪
