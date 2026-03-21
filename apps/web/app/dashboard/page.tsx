"use client";

import { useQuery, useMutation } from "convex/react";
import { api } from "@/convex/_generated/api";
import { useAuth } from "@clerk/nextjs";
import { useEffect, useState } from "react";
import Link from "next/link";

export default function Dashboard() {
  const { isSignedIn } = useAuth();
  const [isLoading, setIsLoading] = useState(true);

  // Queries
  const user = useQuery(api.users.getCurrentUser);
  const repos = useQuery(api.repos.listRepos);
  const workspaces = useQuery(api.workspaces.listWorkspaces);
  const usage = useQuery(api.users.getCurrentUsage);
  const threads = useQuery(api.threads.listThreads);

  useEffect(() => {
    if (user !== undefined && repos !== undefined && workspaces !== undefined) {
      setIsLoading(false);
    }
  }, [user, repos, workspaces]);

  if (!isSignedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-b from-slate-950 to-slate-900">
        <div className="text-center">
          <p className="text-white text-xl mb-4">Please sign in to continue</p>
          <Link href="/" className="text-blue-400 hover:text-blue-300">
            Back to home
          </Link>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-b from-slate-950 to-slate-900">
        <div className="text-white text-xl">Loading...</div>
      </div>
    );
  }

  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-950 to-slate-900">
      <nav className="border-b border-slate-800 px-6 py-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-bold text-white">Dashboard</h1>
          <div className="text-slate-400">
            {user && `Welcome, ${user.githubUsername}`}
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-6 py-8">
        {/* Stats */}
        <div className="grid grid-cols-4 gap-4 mb-8">
          <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
            <div className="text-slate-400 text-sm">Repositories</div>
            <div className="text-3xl font-bold text-white">
              {repos?.length || 0}
            </div>
          </div>
          <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
            <div className="text-slate-400 text-sm">Workspaces</div>
            <div className="text-3xl font-bold text-white">
              {workspaces?.length || 0}
            </div>
          </div>
          <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
            <div className="text-slate-400 text-sm">Plan</div>
            <div className="text-3xl font-bold text-white capitalize">
              {user?.plan || "free"}
            </div>
          </div>
          <div className="bg-slate-800 p-6 rounded-lg border border-slate-700">
            <div className="text-slate-400 text-sm">Usage</div>
            <div className="text-3xl font-bold text-white">
              {usage?.totalMinutes || 0}m
            </div>
          </div>
        </div>

        {/* Repositories */}
        <section className="mb-8">
          <h2 className="text-2xl font-bold text-white mb-4">Repositories</h2>
          {repos && repos.length > 0 ? (
            <div className="grid grid-cols-2 gap-4">
              {repos.map((repo) => (
                <div key={repo._id} className="bg-slate-800 p-4 rounded-lg border border-slate-700">
                  <h3 className="font-bold text-white">{repo.name}</h3>
                  <p className="text-slate-400 text-sm">{repo.fullName}</p>
                  <div className="mt-3 flex gap-2">
                    <a
                      href={repo.htmlUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-400 hover:text-blue-300 text-sm"
                    >
                      View on GitHub
                    </a>
                    {repo.workspace?.status === "running" && (
                      <span className="text-green-400 text-sm">✓ Running</span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 text-center">
              <p className="text-slate-400 mb-4">No repositories connected yet</p>
              <button className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                Connect Repository
              </button>
            </div>
          )}
        </section>

        {/* Workspaces */}
        <section className="mb-8">
          <h2 className="text-2xl font-bold text-white mb-4">Workspaces</h2>
          {workspaces && workspaces.length > 0 ? (
            <div className="space-y-3">
              {workspaces.map((ws) => (
                <div key={ws._id} className="bg-slate-800 p-4 rounded-lg border border-slate-700 flex justify-between items-center">
                  <div>
                    <h3 className="font-bold text-white">{ws.repo?.name}</h3>
                    <p className="text-slate-400 text-sm">
                      Status: <span className="capitalize">{ws.status}</span>
                    </p>
                  </div>
                  {ws.status === "running" && (
                    <button className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 text-sm">
                      Stop
                    </button>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 text-center">
              <p className="text-slate-400 mb-4">No workspaces yet</p>
              <button className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
                Launch Workspace
              </button>
            </div>
          )}
        </section>

        {/* Threads */}
        <section>
          <h2 className="text-2xl font-bold text-white mb-4">Recent Threads</h2>
          {threads && threads.length > 0 ? (
            <div className="space-y-2">
              {threads.slice(0, 5).map((thread) => (
                <div key={thread._id} className="bg-slate-800 p-3 rounded border border-slate-700">
                  <h3 className="font-bold text-white">{thread.title}</h3>
                  <p className="text-slate-400 text-sm">{thread.messageCount || 0} messages</p>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 text-center text-slate-400">
              No threads yet
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
