import { mutation } from "../_generated/server";
import { v } from "convex/values";

/**
 * Launch a new workspace VM for a repo
 */
export const launchVM = mutation({
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

    // Check if workspace already exists for this repo
    const existing = await ctx.db
      .query("workspaces")
      .withIndex("by_user_repo", (q) => q.eq("userId", user._id).eq("repoId", args.repoId))
      .first();

    if (existing) {
      // If already running, return it
      if (existing.status === "running") {
        return existing._id;
      }
      // Otherwise, delete and create new
      await ctx.db.delete(existing._id);
    }

    // Create workspace
    const workspaceId = await ctx.db.insert("workspaces", {
      userId: user._id,
      repoId: args.repoId,
      vmId: undefined,
      status: "starting",
      startedAt: Date.now(),
      stoppedAt: undefined,
      createdAt: Date.now(),
      updatedAt: Date.now(),
    });

    // TODO: Call Fly.io or cloud provider API to actually spin up VM
    // For now, we'll use a scheduled action to update status

    return workspaceId;
  },
});

/**
 * Update workspace status
 */
export const updateWorkspaceStatus = mutation({
  args: {
    workspaceId: v.id("workspaces"),
    status: v.union(
      v.literal("stopped"),
      v.literal("starting"),
      v.literal("running"),
      v.literal("stopping"),
      v.literal("failed")
    ),
    vmId: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const workspace = await ctx.db.get(args.workspaceId);
    if (!workspace) throw new Error("Workspace not found");

    const update: any = {
      status: args.status,
      updatedAt: Date.now(),
    };

    if (args.vmId) {
      update.vmId = args.vmId;
    }

    if (args.status === "running" && !workspace.startedAt) {
      update.startedAt = Date.now();
    }

    if (args.status === "stopped") {
      update.stoppedAt = Date.now();
    }

    await ctx.db.patch(args.workspaceId, update);

    return args.workspaceId;
  },
});

/**
 * Stop a workspace VM
 */
export const stopVM = mutation({
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

    // TODO: Call cloud provider API to destroy VM

    await ctx.db.patch(args.workspaceId, {
      status: "stopping",
      updatedAt: Date.now(),
    });

    // Schedule deletion after a delay (or handle via scheduled action)
    // For now, immediately mark as stopped
    await ctx.db.patch(args.workspaceId, {
      status: "stopped",
      stoppedAt: Date.now(),
      updatedAt: Date.now(),
    });

    return args.workspaceId;
  },
});

/**
 * Delete workspace
 */
export const deleteWorkspace = mutation({
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

    // Delete usage records
    const usageRecords = await ctx.db
      .query("usageRecords")
      .withIndex("by_workspace", (q) => q.eq("workspaceId", args.workspaceId))
      .collect();

    for (const record of usageRecords) {
      await ctx.db.delete(record._id);
    }

    // Delete workspace
    await ctx.db.delete(args.workspaceId);

    return args.workspaceId;
  },
});
