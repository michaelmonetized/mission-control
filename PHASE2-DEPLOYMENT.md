# Phase 2 Deployment Guide

**Deadline:** Sunday 11:59 PM EST (36h 45m remaining)

---

## Checklist

### Pre-Deployment (1 hour)

- [ ] All E2E tests passing locally
- [ ] Environment variables configured
- [ ] Fly.io account created + API token
- [ ] Stripe account created + API keys
- [ ] Clerk app configured (GitHub OAuth)
- [ ] GitHub OAuth scopes approved

### Staging Deployment (2 hours)

- [ ] Deploy Convex backend to staging
  ```bash
  convex deploy --env staging
  ```

- [ ] Deploy Next.js to Vercel preview
  ```bash
  vercel deploy
  ```

- [ ] Deploy WebSocket relay (temp server)
  ```bash
  bun services/websocket-relay/main.ts &
  ```

- [ ] Deploy VM Manager (Fly.io)
  ```bash
  cd services/vm-manager
  fly deploy --app mc-vm-manager-staging
  ```

- [ ] Run smoke tests on staging
  ```bash
  TEST_URL=https://[preview-url] bun test __tests__/phase2-e2e.test.ts
  ```

- [ ] Verify Stripe webhook callbacks working
- [ ] Verify Fly.io machine creation working

### Production Deployment (1 hour)

- [ ] Tag release: `v2.0.0`
  ```bash
  git tag v2.0.0
  git push origin v2.0.0
  ```

- [ ] Deploy Convex to production
  ```bash
  convex deploy --prod
  ```

- [ ] Deploy Next.js to Vercel (production)
  ```bash
  vercel deploy --prod
  ```

- [ ] Deploy WebSocket relay (production)
  ```bash
  fly deploy --app mc-websocket-relay-prod
  ```

- [ ] Deploy VM Manager (production)
  ```bash
  fly deploy --app mc-vm-manager-prod
  ```

- [ ] Verify all health endpoints
  ```bash
  curl https://mission-control.vercel.app/api/health
  curl https://mc-websocket-relay-prod.fly.dev:9002/health
  curl https://mc-vm-manager-prod.fly.dev:9000/health
  ```

- [ ] Monitor logs for 30 minutes
  ```bash
  vercel logs -n 100 --follow
  fly logs --app mc-websocket-relay-prod
  fly logs --app mc-vm-manager-prod
  ```

### Post-Deployment (30 min)

- [ ] Announce on Twitter/X
- [ ] Update GitHub releases
- [ ] Post to team Telegram
- [ ] Monitor error rates (target: <0.1%)
- [ ] Check Sentry for new issues

---

## Environment Variables

### Vercel (Next.js)
```
CONVEX_DEPLOYMENT=xxx
NEXT_PUBLIC_CONVEX_URL=https://xxx.convex.cloud
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_live_xxx
CLERK_SECRET_KEY=sk_live_xxx
STRIPE_SECRET_KEY=sk_live_xxx
STRIPE_PUBLISHABLE_KEY=pk_live_xxx
FLY_API_TOKEN=xxx
FLY_APP_NAME=mc-workspaces-prod
VM_MANAGER_API_KEY=xxx
WEBSOCKET_RELAY_URL=wss://mc-websocket-relay-prod.fly.dev
```

### Fly.io (VM Manager)
```
FLY_API_TOKEN=xxx
FLY_APP_NAME=mc-workspaces-prod
CLAUDE_API_TOKEN=xxx
PORT=9000
```

### Fly.io (WebSocket Relay)
```
PORT=9001
HEALTH_PORT=9002
```

---

## Rollback Plan

**If production fails:**

1. **Immediate:** Disable DNS routing (Vercel → previous working version)
   ```bash
   vercel rollback
   ```

2. **Convex:** Reset to previous backup
   ```bash
   convex deploy --env prod --from-backup
   ```

3. **Fly.io services:** Revert to previous image tag
   ```bash
   fly deploy --app mc-websocket-relay-prod --image xxx:previous-tag
   fly deploy --app mc-vm-manager-prod --image xxx:previous-tag
   ```

4. **Monitor:** Watch error rates return to <0.1%

5. **Debug:** Collect logs and analyze root cause

**Rollback time:** ~10 minutes

---

## Monitoring

### Alerts (Sentry)
- Error rate > 5%
- P99 latency > 2000ms
- Failed workspace launches > 10/hr
- WebSocket connection failures > 50/hr

### Dashboards
- Vercel: https://vercel.com/dashboard
- Fly.io: https://fly.io/dashboard
- Stripe: https://dashboard.stripe.com
- Sentry: https://sentry.io/organizations/hurleyus/issues/

### Key Metrics
- Successful workspace launches (target: 99%+)
- Average VM spinup time (target: <2s)
- WebSocket relay uptime (target: 99.9%+)
- Billing accuracy (target: 100%)
- Error rate (target: <0.1%)

---

## Testing Checklist

**Before any deployment:**

1. ✅ Unit tests passing
   ```bash
   bun test --watch
   ```

2. ✅ E2E tests passing
   ```bash
   bun test __tests__/phase2-e2e.test.ts
   ```

3. ✅ Integration tests passing
   ```bash
   bun test convex-integration.test.ts
   ```

4. ✅ Performance tests passing
   ```bash
   bun test __tests__/performance.test.ts
   ```

5. ✅ Security audit passing
   ```bash
   npm audit
   bun audit
   ```

6. ✅ Type checking clean
   ```bash
   tsc --noEmit
   ```

---

## Support

**If deployment fails:**

1. Check logs: `vercel logs -n 100`
2. Check Sentry errors
3. Check Fly.io dashboard for instance health
4. Post to #incidents Slack channel
5. Rollback immediately (no more than 10 min downtime)

**If performance degrades:**

1. Check CPU/memory usage (Fly.io dashboard)
2. Check WebSocket connection count
3. Check Stripe API latency
4. Autoscale if needed: `fly scale count 2` for VM Manager

**If billing breaks:**

1. Check Stripe webhooks are firing
2. Check Convex mutations are recording usage
3. Manually reconcile via Stripe dashboard

---

## Post-Launch

1. **Monitor for 24h** for stability
2. **Collect feedback** from first 100 users
3. **Fix high-priority bugs**
4. **Announce Phase 3** (team collaboration, SSH integration)

---

**Target: LIVE Sunday 11:59 PM EST**

All systems ready. No blockers. Ship it. 🚀
