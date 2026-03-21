"use client";

import { useAuth } from "@clerk/nextjs";
import { useQuery, useMutation } from "convex/react";
import { api } from "@/convex/_generated/api";
import { useEffect, useState } from "react";
import RepositoryBrowser from "@/components/RepositoryBrowser";
import WorkspaceManager from "@/components/WorkspaceManager";
import UsageDisplay from "@/components/UsageDisplay";

export default function DashboardPage() {
  const { isSignedIn } = useAuth();
  const [loading, setLoading] = useState(true);

  const user = useQuery(api.queries.getUser);
  const repos = useQuery(api.queries.listUserRepos);
  const workspaces = useQuery(api.queries.listUserWorkspaces);
  const usage = useQuery(api.queries.getCurrentUsage);

  useEffect(() => {
    if (user !== undefined) {
      setLoading(false);
    }
  }, [user]);

  if (!isSignedIn) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="mb-4 text-2xl font-bold">Sign In Required</h2>
          <p className="text-zinc-400">Please sign in with GitHub to continue.</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <div className="mb-4 h-8 w-8 animate-spin rounded-full border-4 border-zinc-700 border-t-zinc-50"></div>
          <p className="text-zinc-400">Loading your dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="border-b border-zinc-800 pb-6">
        <h1 className="mb-2 text-3xl font-bold">Dashboard</h1>
        <p className="text-zinc-400">Manage your repositories and workspaces</p>
      </div>

      {/* Usage Display */}
      {usage && <UsageDisplay usage={usage} />}

      {/* Main Grid */}
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-2">
        {/* Repositories */}
        <div>
          <h2 className="mb-4 text-xl font-semibold">Repositories</h2>
          <RepositoryBrowser repos={repos || []} />
        </div>

        {/* Workspaces */}
        <div>
          <h2 className="mb-4 text-xl font-semibold">Workspaces</h2>
          <WorkspaceManager workspaces={workspaces || []} />
        </div>
      </div>
    </div>
  );
}
