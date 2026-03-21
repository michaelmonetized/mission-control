import { mutation } from "../_generated/server";
import { v } from "convex/values";

/**
 * Create a new thread
 */
export const createThread = mutation({
  args: {
    title: v.string(),
    description: v.optional(v.string()),
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

    // If workspaceId provided, verify ownership
    if (args.workspaceId) {
      const workspace = await ctx.db.get(args.workspaceId);
      if (!workspace || workspace.userId !== user._id) {
        throw new Error("Unauthorized");
      }
    }

    const threadId = await ctx.db.insert("threads", {
      userId: user._id,
      title: args.title,
      description: args.description,
      workspaceId: args.workspaceId,
      createdAt: Date.now(),
      updatedAt: Date.now(),
    });

    return threadId;
  },
});

/**
 * Update thread
 */
export const updateThread = mutation({
  args: {
    threadId: v.id("threads"),
    title: v.optional(v.string()),
    description: v.optional(v.string()),
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

    const update: any = { updatedAt: Date.now() };
    if (args.title) update.title = args.title;
    if (args.description) update.description = args.description;

    await ctx.db.patch(args.threadId, update);

    return args.threadId;
  },
});

/**
 * Delete thread
 */
export const deleteThread = mutation({
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

    // Delete all messages in thread
    const messages = await ctx.db
      .query("messages")
      .withIndex("by_thread", (q) => q.eq("threadId", args.threadId))
      .collect();

    for (const message of messages) {
      await ctx.db.delete(message._id);
    }

    // Delete thread
    await ctx.db.delete(args.threadId);

    return args.threadId;
  },
});

/**
 * Add message to thread
 */
export const addMessage = mutation({
  args: {
    threadId: v.id("threads"),
    body: v.string(),
    sender: v.union(v.literal("user"), v.literal("openclaw")),
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

    const messageId = await ctx.db.insert("messages", {
      threadId: args.threadId,
      userId: user._id,
      body: args.body,
      sender: args.sender,
      createdAt: Date.now(),
    });

    // Update thread's updatedAt
    await ctx.db.patch(args.threadId, {
      updatedAt: Date.now(),
    });

    return messageId;
  },
});

/**
 * Delete message
 */
export const deleteMessage = mutation({
  args: {
    messageId: v.id("messages"),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const message = await ctx.db.get(args.messageId);
    if (!message) throw new Error("Message not found");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user || user._id !== message.userId) {
      throw new Error("Unauthorized");
    }

    const thread = await ctx.db.get(message.threadId);
    if (thread) {
      await ctx.db.patch(message.threadId, {
        updatedAt: Date.now(),
      });
    }

    await ctx.db.delete(args.messageId);

    return args.messageId;
  },
});

/**
 * Record webhook event
 */
export const recordWebhookEvent = mutation({
  args: {
    event: v.string(),
    action: v.optional(v.string()),
    payload: v.any(),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    
    // Allow webhook to record events without auth (uses Clerk service token)
    // In production, verify the webhook signature in the API handler

    let userId: any = null;

    if (identity) {
      const user = await ctx.db
        .query("users")
        .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
        .first();

      if (user) {
        userId = user._id;
      }
    }

    if (!userId && args.payload.userId) {
      userId = args.payload.userId;
    }

    if (!userId) {
      console.warn("Could not identify user for webhook event", args.event);
      return null;
    }

    const eventId = await ctx.db.insert("webhookEvents", {
      userId,
      event: args.event,
      action: args.action,
      payload: args.payload,
      processedAt: Date.now(),
    });

    return eventId;
  },
});
