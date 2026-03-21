import { query } from "../_generated/server";
import { v } from "convex/values";

/**
 * List all threads for current user
 */
export const listThreads = query({
  args: {
    workspaceId: v.optional(v.id("workspaces")),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    let threads;

    if (args.workspaceId) {
      // Verify workspace ownership
      const workspace = await ctx.db.get(args.workspaceId);
      if (!workspace || workspace.userId !== user._id) {
        throw new Error("Unauthorized");
      }

      threads = await ctx.db
        .query("threads")
        .withIndex("by_workspace", (q) => q.eq("workspaceId", args.workspaceId))
        .order("desc")
        .collect();
    } else {
      threads = await ctx.db
        .query("threads")
        .withIndex("by_user", (q) => q.eq("userId", user._id))
        .order("desc")
        .collect();
    }

    // Enrich with message count
    const enriched = await Promise.all(
      threads.map(async (thread) => {
        const messages = await ctx.db
          .query("messages")
          .withIndex("by_thread", (q) => q.eq("threadId", thread._id))
          .collect();

        return {
          ...thread,
          messageCount: messages.length,
        };
      })
    );

    return enriched;
  },
});

/**
 * Get single thread
 */
export const getThread = query({
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

    const messages = await ctx.db
      .query("messages")
      .withIndex("by_thread", (q) => q.eq("threadId", args.threadId))
      .order("asc")
      .collect();

    return {
      ...thread,
      messages,
      messageCount: messages.length,
    };
  },
});

/**
 * Get thread messages
 */
export const getThreadMessages = query({
  args: {
    threadId: v.id("threads"),
    limit: v.optional(v.number()),
    offset: v.optional(v.number()),
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

    let query = ctx.db
      .query("messages")
      .withIndex("by_thread", (q) => q.eq("threadId", args.threadId));

    const total = (
      await ctx.db
        .query("messages")
        .withIndex("by_thread", (q) => q.eq("threadId", args.threadId))
        .collect()
    ).length;

    const messages = await query
      .order("asc")
      .skip(args.offset || 0)
      .take(args.limit || 50);

    return {
      messages,
      total,
      hasMore: total > ((args.offset || 0) + (args.limit || 50)),
    };
  },
});

/**
 * Get recent threads
 */
export const getRecentThreads = query({
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

    const threads = await ctx.db
      .query("threads")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .order("desc")
      .take(args.limit || 10);

    return threads;
  },
});

/**
 * Search threads
 */
export const searchThreads = query({
  args: {
    query: v.string(),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const threads = await ctx.db
      .query("threads")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();

    const lowerQuery = args.query.toLowerCase();
    const filtered = threads.filter(
      (thread) =>
        thread.title.toLowerCase().includes(lowerQuery) ||
        (thread.description &&
          thread.description.toLowerCase().includes(lowerQuery))
    );

    return filtered;
  },
});

/**
 * Get thread count for user
 */
export const getThreadCount = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const threads = await ctx.db
      .query("threads")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();

    return threads.length;
  },
});
