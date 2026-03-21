import { subscriptionQuery } from "../_generated/server";
import { v } from "convex/values";

/**
 * Subscribe to workspace status updates
 */
export const subscribeToWorkspaceStatus = subscriptionQuery({
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

    return await ctx.db.get(args.workspaceId);
  },
});

/**
 * Subscribe to all user workspaces
 */
export const subscribeToUserWorkspaces = subscriptionQuery({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    return await ctx.db
      .query("workspaces")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();
  },
});

/**
 * Subscribe to usage updates in real-time
 */
export const subscribeToUsageUpdates = subscriptionQuery({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const billingPeriod = new Date(Date.now()).toISOString().slice(0, 7);

    return await ctx.db
      .query("usageRecords")
      .withIndex("by_user_period", (q) =>
        q.eq("userId", user._id).eq("billingPeriod", billingPeriod)
      )
      .collect();
  },
});
