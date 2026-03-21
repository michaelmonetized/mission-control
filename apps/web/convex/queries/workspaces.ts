import { query } from "../_generated/server";
import { v } from "convex/values";

/**
 * List all workspaces for current user
 */
export const listWorkspaces = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const workspaces = await ctx.db
      .query("workspaces")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();

    // Enrich with repo info
    const enriched = await Promise.all(
      workspaces.map(async (ws) => {
        const repo = await ctx.db.get(ws.repoId);
        return {
          ...ws,
          repo: repo || null,
        };
      })
    );

    return enriched;
  },
});

/**
 * Get single workspace
 */
export const getWorkspace = query({
  args: {
    workspaceId: v.id("workspaces"),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const workspace = await ctx.db.get(args.workspaceId);
    if (!workspace) throw new Error("Workspace not found");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user || user._id !== workspace.userId) {
      throw new Error("Unauthorized");
    }

    const repo = await ctx.db.get(workspace.repoId);

    return {
      ...workspace,
      repo: repo || null,
    };
  },
});

/**
 * Get workspace for repo
 */
export const getWorkspaceForRepo = query({
  args: {
    repoId: v.id("repos"),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const repo = await ctx.db.get(args.repoId);
    if (!repo) throw new Error("Repo not found");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user || user._id !== repo.userId) {
      throw new Error("Unauthorized");
    }

    return await ctx.db
      .query("workspaces")
      .withIndex("by_user_repo", (q) =>
        q.eq("userId", user._id).eq("repoId", args.repoId)
      )
      .first();
  },
});

/**
 * List active workspaces
 */
export const listActiveWorkspaces = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const workspaces = await ctx.db
      .query("workspaces")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();

    const active = workspaces.filter((ws) => ws.status === "running");

    // Enrich with repo info
    const enriched = await Promise.all(
      active.map(async (ws) => {
        const repo = await ctx.db.get(ws.repoId);
        return {
          ...ws,
          repo: repo || null,
        };
      })
    );

    return enriched;
  },
});

/**
 * Get workspace stats (for dashboard)
 */
export const getWorkspaceStats = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const workspaces = await ctx.db
      .query("workspaces")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();

    const stats = {
      total: workspaces.length,
      running: workspaces.filter((ws) => ws.status === "running").length,
      stopped: workspaces.filter((ws) => ws.status === "stopped").length,
      starting: workspaces.filter((ws) => ws.status === "starting").length,
      failed: workspaces.filter((ws) => ws.status === "failed").length,
    };

    return stats;
  },
});

/**
 * Get workspace uptime
 */
export const getWorkspaceUptime = query({
  args: {
    workspaceId: v.id("workspaces"),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const workspace = await ctx.db.get(args.workspaceId);
    if (!workspace) throw new Error("Workspace not found");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user || user._id !== workspace.userId) {
      throw new Error("Unauthorized");
    }

    if (!workspace.startedAt) return 0;

    const endTime = workspace.stoppedAt || Date.now();
    return endTime - workspace.startedAt;
  },
});
