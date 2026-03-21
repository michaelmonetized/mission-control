import { mutation } from "../_generated/server";
import { v } from "convex/values";

/**
 * Create or update user after GitHub OAuth
 * Called by Clerk after successful authentication
 */
export const syncUser = mutation({
  args: {
    clerkId: v.string(),
    githubId: v.number(),
    githubUsername: v.string(),
    email: v.string(),
    avatar: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    // Check if user exists
    const existing = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", args.clerkId))
      .first();

    if (existing) {
      // Update user
      await ctx.db.patch(existing._id, {
        githubId: args.githubId,
        githubUsername: args.githubUsername,
        email: args.email,
        avatar: args.avatar,
        updatedAt: Date.now(),
      });
      return existing._id;
    }

    // Create new user
    const userId = await ctx.db.insert("users", {
      clerkId: args.clerkId,
      githubId: args.githubId,
      githubUsername: args.githubUsername,
      email: args.email,
      avatar: args.avatar,
      claudeApiKey: undefined,
      stripeCustomerId: undefined,
      freeMinutesUsed: 0,
      plan: "free",
      billingPeriodStart: Date.now(),
      createdAt: Date.now(),
      updatedAt: Date.now(),
    });

    return userId;
  },
});

/**
 * Update Claude API key (encrypted at rest)
 */
export const updateClaudeApiKey = mutation({
  args: {
    apiKey: v.string(),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    await ctx.db.patch(user._id, {
      claudeApiKey: args.apiKey,
      updatedAt: Date.now(),
    });

    return user._id;
  },
});

/**
 * Upgrade user to Pro plan
 */
export const upgradeToPro = mutation({
  args: {
    stripeCustomerId: v.string(),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    await ctx.db.patch(user._id, {
      plan: "pro",
      stripeCustomerId: args.stripeCustomerId,
      billingPeriodStart: Date.now(),
      updatedAt: Date.now(),
    });

    return user._id;
  },
});

/**
 * Record usage for billing
 */
export const recordUsage = mutation({
  args: {
    workspaceId: v.id("workspaces"),
    durationMinutes: v.number(),
    cost: v.number(),
  },
  handler: async (ctx, args) => {
    const workspace = await ctx.db.get(args.workspaceId);
    if (!workspace) throw new Error("Workspace not found");

    const user = await ctx.db.get(workspace.userId);
    if (!user) throw new Error("User not found");

    const billingPeriod = new Date(Date.now()).toISOString().slice(0, 7); // "2026-03"

    const recordId = await ctx.db.insert("usageRecords", {
      userId: workspace.userId,
      workspaceId: args.workspaceId,
      durationMinutes: args.durationMinutes,
      cost: args.cost,
      billingPeriod,
      recordedAt: Date.now(),
    });

    // Update user's free minutes (for free tier)
    if (user.plan === "free") {
      const newFreeMinutesUsed = user.freeMinutesUsed + args.durationMinutes;
      await ctx.db.patch(workspace.userId, {
        freeMinutesUsed: newFreeMinutesUsed,
        updatedAt: Date.now(),
      });
    }

    return recordId;
  },
});
