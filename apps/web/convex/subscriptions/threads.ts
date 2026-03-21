import { subscriptionQuery } from "../_generated/server";
import { v } from "convex/values";

/**
 * Subscribe to messages in a thread (real-time chat)
 */
export const subscribeToThreadMessages = subscriptionQuery({
  args: {
    threadId: v.id("threads"),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const thread = await ctx.db.get(args.threadId);
    if (!thread) throw new Error("Thread not found");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user || user._id !== thread.userId) {
      throw new Error("Unauthorized");
    }

    return await ctx.db
      .query("messages")
      .withIndex("by_thread", (q) => q.eq("threadId", args.threadId))
      .order("asc");
  },
});

/**
 * Subscribe to user's threads
 */
export const subscribeToUserThreads = subscriptionQuery({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    return await ctx.db
      .query("threads")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .order("desc");
  },
});

/**
 * Subscribe to workspace threads
 */
export const subscribeToWorkspaceThreads = subscriptionQuery({
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

    return await ctx.db
      .query("threads")
      .withIndex("by_workspace", (q) => q.eq("workspaceId", args.workspaceId))
      .order("desc");
  },
});
