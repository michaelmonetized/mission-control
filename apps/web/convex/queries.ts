import { query, QueryCtx } from "./_generated/server";
import { Id } from "./_generated/dataModel";

// User Queries
export const getUser = query({
  args: {},
  async handler(ctx) {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) return null;

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    return user || null;
  },
});

export const listUserRepos = query({
  args: {},
  async handler(ctx) {
    const userId = await requireAuthQuery(ctx);

    const repos = await ctx.db
      .query("repos")
      .withIndex("by_user", (q) => q.eq("userId", userId))
      .collect();

    return repos.map((repo) => ({
      repoId: repo._id,
      name: repo.name,
      fullName: repo.fullName,
      description: repo.description,
      private: repo.private,
      htmlUrl: repo.htmlUrl,
      status: "connected" as const,
    }));
  },
});

export const listUserWorkspaces = query({
  args: {},
  async handler(ctx) {
    const userId = await requireAuthQuery(ctx);

    const workspaces = await ctx.db
      .query("workspaces")
      .withIndex("by_user", (q) => q.eq("userId", userId))
      .collect();

    const result = [];
    for (const ws of workspaces) {
      const repo = await ctx.db.get(ws.repoId);
      if (repo) {
        result.push({
          workspaceId: ws._id,
          repoName: repo.name,
          status: ws.status,
          startedAt: ws.startedAt || null,
          stoppedAt: ws.stoppedAt || null,
          vmId: ws.vmId || null,
        });
      }
    }

    return result;
  },
});

export const getWorkspaceStatus = query({
  args: {},
  async handler(ctx) {
    const userId = await requireAuthQuery(ctx);
    const workspaces = await ctx.db
      .query("workspaces")
      .withIndex("by_user", (q) => q.eq("userId", userId))
      .collect();

    if (!workspaces.length) return null;

    const ws = workspaces[0]; // Most recent
    return {
      status: ws.status,
      vmId: ws.vmId || null,
      uptime: ws.startedAt ? Math.round((Date.now() - ws.startedAt) / 60000) : 0,
      cpuUsage: 0, // TODO: Get from Fly API
      memoryUsage: 0, // TODO: Get from Fly API
      lastUpdate: ws.updatedAt,
    };
  },
});

export const getCurrentUsage = query({
  args: {},
  async handler(ctx) {
    const userId = await requireAuthQuery(ctx);
    const billingPeriod = new Date().toISOString().substring(0, 7); // "2026-03"

    const records = await ctx.db
      .query("usageRecords")
      .withIndex("by_user_period", (q) =>
        q.eq("userId", userId).eq("billingPeriod", billingPeriod)
      )
      .collect();

    const minutes = records.reduce((sum, r) => sum + r.durationMinutes, 0);
    const cost = records.reduce((sum, r) => sum + r.cost, 0);
    const limit = 100; // Free tier: 100 minutes/month

    return {
      minutes: Math.round(minutes),
      cost: Math.round(cost * 100) / 100,
      remaining: Math.max(0, limit - minutes),
      period: billingPeriod,
    };
  },
});

// Real-Time Subscriptions
export const subscribeToWorkspaceStatus = query({
  args: {},
  async handler(ctx) {
    const userId = await requireAuthQuery(ctx);
    const workspaces = await ctx.db
      .query("workspaces")
      .withIndex("by_user", (q) => q.eq("userId", userId))
      .collect();

    if (!workspaces.length) return null;

    const ws = workspaces[0];
    return {
      status: ws.status,
      vmId: ws.vmId || null,
      uptime: ws.startedAt ? Math.round((Date.now() - ws.startedAt) / 60000) : 0,
      cpuUsage: 0,
      memoryUsage: 0,
      lastUpdate: ws.updatedAt,
    };
  },
});

export const subscribeToUsage = query({
  args: {},
  async handler(ctx) {
    const userId = await requireAuthQuery(ctx);
    const billingPeriod = new Date().toISOString().substring(0, 7);

    const records = await ctx.db
      .query("usageRecords")
      .withIndex("by_user_period", (q) =>
        q.eq("userId", userId).eq("billingPeriod", billingPeriod)
      )
      .collect();

    const minutes = records.reduce((sum, r) => sum + r.durationMinutes, 0);
    const cost = records.reduce((sum, r) => sum + r.cost, 0);

    return {
      minutes: Math.round(minutes),
      cost: Math.round(cost * 100) / 100,
      remaining: Math.max(0, 100 - minutes),
    };
  },
});

// Helper
async function requireAuthQuery(ctx: QueryCtx) {
  const identity = await ctx.auth.getUserIdentity();
  if (!identity) throw new Error("Not authenticated");
  return identity.subject as Id<"users">;
}
