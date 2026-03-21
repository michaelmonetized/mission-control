import { query } from "../_generated/server";
import { v } from "convex/values";

/**
 * List all connected repos for current user
 */
export const listRepos = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const repos = await ctx.db
      .query("repos")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();

    // Enrich with workspace info
    const enriched = await Promise.all(
      repos.map(async (repo) => {
        const workspace = await ctx.db
          .query("workspaces")
          .withIndex("by_user_repo", (q) =>
            q.eq("userId", user._id).eq("repoId", repo._id)
          )
          .first();

        return {
          ...repo,
          workspace: workspace || null,
        };
      })
    );

    return enriched;
  },
});

/**
 * Get single repo by ID
 */
export const getRepo = query({
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

    // Get workspace info
    const workspace = await ctx.db
      .query("workspaces")
      .withIndex("by_user_repo", (q) =>
        q.eq("userId", user._id).eq("repoId", args.repoId)
      )
      .first();

    return {
      ...repo,
      workspace: workspace || null,
    };
  },
});

/**
 * Get repo by GitHub ID
 */
export const getRepoByGithubId = query({
  args: {
    githubId: v.number(),
  },
  handler: async (ctx, args) => {
    return await ctx.db
      .query("repos")
      .withIndex("by_github_id", (q) => q.eq("githubId", args.githubId))
      .first();
  },
});

/**
 * Search repos
 */
export const searchRepos = query({
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

    const repos = await ctx.db
      .query("repos")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();

    const lowerQuery = args.query.toLowerCase();
    const filtered = repos.filter(
      (repo) =>
        repo.name.toLowerCase().includes(lowerQuery) ||
        repo.fullName.toLowerCase().includes(lowerQuery) ||
        (repo.description &&
          repo.description.toLowerCase().includes(lowerQuery))
    );

    return filtered;
  },
});

/**
 * Get repo count for user
 */
export const getRepoCount = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_clerk", (q) => q.eq("clerkId", identity.subject))
      .first();

    if (!user) throw new Error("User not found");

    const repos = await ctx.db
      .query("repos")
      .withIndex("by_user", (q) => q.eq("userId", user._id))
      .collect();

    return repos.length;
  },
});
