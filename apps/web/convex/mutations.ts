import { v } from "convex/values";
import { mutation, MutationCtx } from "./_generated/server";
import { Id } from "./_generated/dataModel";

// User Setup
export const createUserOnFirstLogin = mutation({
  args: {
    clerkId: v.string(),
    githubId: v.number(),
    githubUsername: v.string(),
    email: v.string(),
    avatar: v.optional(v.string()),
  },
  async handler(ctx, args) {
    const existing = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", args.clerkId))
      .first();

    if (existing) return existing;

    const userId = await ctx.db.insert("users", {
      clerkId: args.clerkId,
      githubId: args.githubId,
      githubUsername: args.githubUsername,
      email: args.email,
      avatar: args.avatar,
      plan: "free",
      freeMinutesUsed: 0,
      createdAt: Date.now(),
      updatedAt: Date.now(),
    });

    return { userId, plan: "free" };
  },
});

export const updateClaudeApiKey = mutation({
  args: {
    encryptedKey: v.string(),
  },
  async handler(ctx, args) {
    const userId = await requireAuth(ctx);

    await ctx.db.patch(userId, {
      claudeApiKey: args.encryptedKey,
      updatedAt: Date.now(),
    });

    return { success: true };
  },
});

export const updateStripeCustomerId = mutation({
  args: {
    stripeCustomerId: v.string(),
  },
  async handler(ctx, args) {
    const userId = await requireAuth(ctx);

    await ctx.db.patch(userId, {
      stripeCustomerId: args.stripeCustomerId,
      updatedAt: Date.now(),
    });

    return { success: true };
  },
});

// Repository Management
export const syncGitHubRepos = mutation({
  args: {},
  async handler(ctx) {
    const userId = await requireAuth(ctx);
    const user = await ctx.db.get(userId);

    if (!user) throw new Error("User not found");

    // Note: In real implementation, call GitHub API to fetch user repos
    // For now, return empty (assumes client will populate)

    const repos = await ctx.db
      .query("repos")
      .withIndex("by_user", (q) => q.eq("userId", userId))
      .collect();

    return {
      synced: repos.length,
      repos: repos.map((r) => ({ id: r._id, name: r.name, url: r.cloneUrl })),
    };
  },
});

export const connectRepo = mutation({
  args: {
    githubId: v.number(),
    name: v.string(),
    fullName: v.string(),
    description: v.optional(v.string()),
    private: v.boolean(),
    defaultBranch: v.string(),
    cloneUrl: v.string(),
    htmlUrl: v.string(),
  },
  async handler(ctx, args) {
    const userId = await requireAuth(ctx);

    // Check if already connected
    const existing = await ctx.db
      .query("repos")
      .withIndex("by_github_id", (q) => q.eq("githubId", args.githubId))
      .first();

    if (existing) return { repoId: existing._id, name: existing.name, status: "connected" };

    const repoId = await ctx.db.insert("repos", {
      userId,
      githubId: args.githubId,
      name: args.name,
      fullName: args.fullName,
      description: args.description,
      private: args.private,
      defaultBranch: args.defaultBranch,
      cloneUrl: args.cloneUrl,
      htmlUrl: args.htmlUrl,
      syncedAt: Date.now(),
    });

    return { repoId, name: args.name, status: "ready" };
  },
});

export const disconnectRepo = mutation({
  args: {
    repoId: v.id("repos"),
  },
  async handler(ctx, args) {
    const userId = await requireAuth(ctx);
    const repo = await ctx.db.get(args.repoId);

    if (!repo || repo.userId !== userId) throw new Error("Unauthorized");

    // Delete associated workspaces first
    const workspaces = await ctx.db
      .query("workspaces")
      .withIndex("by_repo", (q) => q.eq("repoId", args.repoId))
      .collect();

    for (const ws of workspaces) {
      // TODO: Stop any running VMs
      await ctx.db.delete(ws._id);
    }

    await ctx.db.delete(args.repoId);
    return { success: true };
  },
});

// Workspace Lifecycle
export const launchWorkspace = mutation({
  args: {
    repoId: v.id("repos"),
  },
  async handler(ctx, args) {
    const userId = await requireAuth(ctx);
    const repo = await ctx.db.get(args.repoId);

    if (!repo || repo.userId !== userId) throw new Error("Unauthorized");

    // Check if workspace already exists
    const existing = await ctx.db
      .query("workspaces")
      .withIndex("by_user_repo", (q) => q.eq("userId", userId).eq("repoId", args.repoId))
      .first();

    if (existing && existing.status !== "stopped") {
      return { workspaceId: existing._id, vmId: existing.vmId, status: existing.status };
    }

    // Create new workspace
    const workspaceId = await ctx.db.insert("workspaces", {
      userId,
      repoId: args.repoId,
      status: "starting",
      createdAt: Date.now(),
      updatedAt: Date.now(),
    });

    // TODO: Call Fly.io API to launch VM
    // await launchVm(userId, repo, workspaceId);

    return { workspaceId, vmId: "pending", status: "starting" };
  },
});

export const stopWorkspace = mutation({
  args: {
    workspaceId: v.id("workspaces"),
  },
  async handler(ctx, args) {
    const userId = await requireAuth(ctx);
    const workspace = await ctx.db.get(args.workspaceId);

    if (!workspace || workspace.userId !== userId) throw new Error("Unauthorized");

    // TODO: Call Fly.io API to stop VM
    // if (workspace.vmId) await stopVm(workspace.vmId);

    await ctx.db.patch(args.workspaceId, {
      status: "stopping",
      updatedAt: Date.now(),
    });

    return { success: true, status: "stopping" };
  },
});

export const recordUsage = mutation({
  args: {
    workspaceId: v.id("workspaces"),
    durationMinutes: v.number(),
  },
  async handler(ctx, args) {
    const userId = await requireAuth(ctx);
    const workspace = await ctx.db.get(args.workspaceId);

    if (!workspace || workspace.userId !== userId) throw new Error("Unauthorized");

    const costPerMinute = 0.001; // $0.001 per minute
    const cost = args.durationMinutes * costPerMinute;
    const billingPeriod = new Date().toISOString().substring(0, 7); // "2026-03"

    await ctx.db.insert("usageRecords", {
      userId,
      workspaceId: args.workspaceId,
      durationMinutes: args.durationMinutes,
      cost,
      billingPeriod,
      recordedAt: Date.now(),
    });

    return { cost, recorded: true };
  },
});

// Billing
export const requestInvoice = mutation({
  args: {},
  async handler(ctx) {
    const userId = await requireAuth(ctx);

    // TODO: Call Stripe API to generate invoice
    // const invoice = await stripe.invoices.create(...);

    return { invoiceId: "inv_pending", url: "#" };
  },
});

// Helper
async function requireAuth(ctx: MutationCtx) {
  const identity = await ctx.auth.getUserIdentity();
  if (!identity) throw new Error("Not authenticated");
  return identity.subject as Id<"users">;
}
