import { mutation } from "../_generated/server";
import { v } from "convex/values";

/**
 * Add GitHub repo to user's connected repos
 */
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
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    // Check if repo already connected
    const existing = await ctx.db
      .query("repos")
      .withIndex("by_github_id", (q) => q.eq("githubId", args.githubId))
      .first();

    if (existing && existing.userId === user._id) {
      // Already connected, just update sync time
      await ctx.db.patch(existing._id, {
        syncedAt: Date.now(),
        lastSynced: Date.now(),
      });
      return existing._id;
    }

    // Connect new repo
    const repoId = await ctx.db.insert("repos", {
      userId: user._id,
      githubId: args.githubId,
      name: args.name,
      fullName: args.fullName,
      description: args.description,
      private: args.private,
      defaultBranch: args.defaultBranch,
      cloneUrl: args.cloneUrl,
      htmlUrl: args.htmlUrl,
      lastSynced: undefined,
      syncedAt: Date.now(),
    });

    return repoId;
  },
});

/**
 * Disconnect a repo
 */
export const disconnectRepo = mutation({
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

    // Delete associated workspaces
    const workspaces = await ctx.db
      .query("workspaces")
      .withIndex("by_repo", (q) => q.eq("repoId", args.repoId))
      .collect();

    for (const workspace of workspaces) {
      await ctx.db.delete(workspace._id);
    }

    // Delete repo
    await ctx.db.delete(args.repoId);

    return args.repoId;
  },
});

/**
 * Sync repos from GitHub
 * Called periodically or on user request
 */
export const syncRepos = mutation({
  args: {
    repos: v.array(
      v.object({
        githubId: v.number(),
        name: v.string(),
        fullName: v.string(),
        description: v.optional(v.string()),
        private: v.boolean(),
        defaultBranch: v.string(),
        cloneUrl: v.string(),
        htmlUrl: v.string(),
      })
    ),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const results = [];

    for (const repo of args.repos) {
      const existing = await ctx.db
        .query("repos")
        .withIndex("by_github_id", (q) => q.eq("githubId", repo.githubId))
        .first();

      if (existing && existing.userId === user._id) {
        // Update existing
        await ctx.db.patch(existing._id, {
          name: repo.name,
          fullName: repo.fullName,
          description: repo.description,
          private: repo.private,
          defaultBranch: repo.defaultBranch,
          cloneUrl: repo.cloneUrl,
          htmlUrl: repo.htmlUrl,
          syncedAt: Date.now(),
        });
        results.push(existing._id);
      } else if (!existing) {
        // Create new
        const repoId = await ctx.db.insert("repos", {
          userId: user._id,
          githubId: repo.githubId,
          name: repo.name,
          fullName: repo.fullName,
          description: repo.description,
          private: repo.private,
          defaultBranch: repo.defaultBranch,
          cloneUrl: repo.cloneUrl,
          htmlUrl: repo.htmlUrl,
          lastSynced: undefined,
          syncedAt: Date.now(),
        });
        results.push(repoId);
      }
    }

    return results;
  },
});
