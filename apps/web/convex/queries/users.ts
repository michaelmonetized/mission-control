import { query } from "../_generated/server";
import { v } from "convex/values";

/**
 * Get current user
 */
export const getCurrentUser = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) return null;

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    return user || null;
  },
});

/**
 * Get user by ID
 */
export const getUserById = query({
  args: {
    userId: v.id("users"),
  },
  handler: async (ctx, args) => {
    return await ctx.db.get(args.userId);
  },
});

/**
 * Get current user's usage for billing period
 */
export const getCurrentUsage = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const billingPeriod = new Date(Date.now()).toISOString().slice(0, 7); // "2026-03"

    const records = await ctx.db
      .query("usageRecords")
      .withIndex("by_user_period", (q) =>
        q.eq("userId", user._id).eq("billingPeriod", billingPeriod)
      )
      .collect();

    const totalMinutes = records.reduce((sum, r) => sum + r.durationMinutes, 0);
    const totalCost = records.reduce((sum, r) => sum + r.cost, 0);

    const freeMinutesRemaining =
      user.plan === "free" ? Math.max(0, 100 - user.freeMinutesUsed) : 1000;

    return {
      userId: user._id,
      plan: user.plan,
      billingPeriod,
      totalMinutes,
      totalCost,
      freeMinutesUsed: user.freeMinutesUsed,
      freeMinutesRemaining,
      records,
    };
  },
});

/**
 * Get usage history
 */
export const getUsageHistory = query({
  args: {
    limit: v.optional(v.number()),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const records = await ctx.db
      .query("usageRecords")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .order("desc")
      .take(args.limit || 100);

    return records;
  },
});

/**
 * Check if user can launch more VMs (free tier limit)
 */
export const canLaunchVM = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    if (user.plan !== "free") return true;

    const freeMinutesRemaining = Math.max(0, 100 - user.freeMinutesUsed);
    return freeMinutesRemaining > 0;
  },
});
