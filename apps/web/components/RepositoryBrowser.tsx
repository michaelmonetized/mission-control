"use client";

import { useMutation } from "convex/react";
import { api } from "@/convex/_generated/api";
import { useState } from "react";

interface Repo {
  repoId: string;
  name: string;
  fullName: string;
  description?: string;
  private: boolean;
  htmlUrl: string;
  status: string;
}

export default function RepositoryBrowser({ repos }: { repos: Repo[] }) {
  const [syncing, setSyncing] = useState(false);
  const syncRepos = useMutation(api.mutations.syncGitHubRepos);

  const handleSync = async () => {
    setSyncing(true);
    try {
      await syncRepos({});
    } catch (err) {
      console.error("Sync failed:", err);
    } finally {
      setSyncing(false);
    }
  };

  return (
    <div className="space-y-4">
      <button
        onClick={handleSync}
        disabled={syncing}
        className="rounded-lg bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:bg-zinc-700"
      >
        {syncing ? "Syncing..." : "Sync Repositories"}
      </button>

      {repos && repos.length > 0 ? (
        <div className="space-y-2">
          {repos.map((repo) => (
            <a
              key={repo.repoId}
              href={repo.htmlUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="block rounded-lg border border-zinc-700 bg-zinc-900 p-4 hover:border-zinc-600 hover:bg-zinc-800"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <h3 className="font-semibold">
                    {repo.private && "🔒 "}
                    {repo.name}
                  </h3>
                  <p className="text-sm text-zinc-400">{repo.fullName}</p>
                  {repo.description && (
                    <p className="mt-1 text-sm text-zinc-500">{repo.description}</p>
                  )}
                </div>
                <span className="rounded bg-green-900/50 px-2 py-1 text-xs text-green-200">
                  {repo.status}
                </span>
              </div>
            </a>
          ))}
        </div>
      ) : (
        <div className="rounded-lg border border-zinc-700 bg-zinc-900 p-8 text-center">
          <p className="text-zinc-400">No repositories connected yet</p>
          <p className="mt-2 text-sm text-zinc-500">
            Click "Sync Repositories" to connect your GitHub repos
          </p>
        </div>
      )}
    </div>
  );
}
