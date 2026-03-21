import { defineSchema, defineTable } from "convex/server";
import { v } from "convex/values";

export default defineSchema({
  // User accounts (GitHub OAuth via Clerk)
  users: defineTable({
    clerkId: v.string(),
    githubId: v.number(),
    githubUsername: v.string(),
    email: v.string(),
    avatar: v.optional(v.string()),
    claudeApiKey: v.optional(v.string()), // encrypted at rest by Convex
    stripeCustomerId: v.optional(v.string()),
    freeMinutesUsed: v.number(),
    plan: v.union(v.literal("free"), v.literal("pro"), v.literal("team")),
    billingPeriodStart: v.optional(v.number()), // timestamp
    createdAt: v.number(),
    updatedAt: v.number(),
  })
    .index("by_clerk", ["clerkId"])
    .index("by_github", ["githubId"]),

  // GitHub repositories (public & private if granted)
  repos: defineTable({
    userId: v.id("users"),
    githubId: v.number(),
    name: v.string(),
    fullName: v.string(),
    description: v.optional(v.string()),
    private: v.boolean(),
    defaultBranch: v.string(),
    cloneUrl: v.string(),
    htmlUrl: v.string(),
    lastSynced: v.optional(v.number()),
    syncedAt: v.number(),
  })
    .index("by_user", ["userId"])
    .index("by_github_id", ["githubId"]),

  // Isolated workspaces (one per repo)
  workspaces: defineTable({
    userId: v.id("users"),
    repoId: v.id("repos"),
    vmId: v.optional(v.string()), // Fly Machine ID or equivalent
    status: v.union(
      v.literal("stopped"),
      v.literal("starting"),
      v.literal("running"),
      v.literal("stopping"),
      v.literal("failed")
    ),
    startedAt: v.optional(v.number()),
    stoppedAt: v.optional(v.number()),
    createdAt: v.number(),
    updatedAt: v.number(),
  })
    .index("by_user", ["userId"])
    .index("by_repo", ["repoId"])
    .index("by_user_repo", ["userId", "repoId"]),

  // Usage tracking for billing
  usageRecords: defineTable({
    userId: v.id("users"),
    workspaceId: v.id("workspaces"),
    durationMinutes: v.number(),
    cost: v.number(),
    billingPeriod: v.string(), // "2026-03"
    recordedAt: v.number(),
  })
    .index("by_user", ["userId"])
    .index("by_user_period", ["userId", "billingPeriod"])
    .index("by_workspace", ["workspaceId"]),

  // Thread management for OpenClaw integration
  threads: defineTable({
    userId: v.id("users"),
    title: v.string(),
    description: v.optional(v.string()),
    workspaceId: v.optional(v.id("workspaces")),
    createdAt: v.number(),
    updatedAt: v.number(),
  })
    .index("by_user", ["userId"])
    .index("by_workspace", ["workspaceId"]),

  // Messages within threads
  messages: defineTable({
    threadId: v.id("threads"),
    userId: v.id("users"),
    body: v.string(),
    sender: v.union(v.literal("user"), v.literal("openclaw")),
    createdAt: v.number(),
  })
    .index("by_thread", ["threadId"])
    .index("by_user", ["userId"]),

  // GitHub webhook events (for audit trail)
  webhookEvents: defineTable({
    userId: v.id("users"),
    event: v.string(),
    action: v.optional(v.string()),
    payload: v.any(),
    processedAt: v.number(),
  })
    .index("by_user", ["userId"])
    .index("by_event", ["event"]),
});
