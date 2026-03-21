"use client";

import { useMutation } from "convex/react";
import { api } from "@/convex/_generated/api";

interface Workspace {
  workspaceId: string;
  repoName: string;
  status: "stopped" | "starting" | "running" | "stopping" | "failed";
  startedAt: number | null;
  stoppedAt: number | null;
  vmId: string | null;
}

export default function WorkspaceManager({ workspaces }: { workspaces: Workspace[] }) {
  const stopWorkspace = useMutation(api.mutations.stopWorkspace);

  const handleStop = async (workspaceId: string) => {
    try {
      await stopWorkspace({ workspaceId: workspaceId as any });
    } catch (err) {
      console.error("Stop failed:", err);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "running":
        return "bg-green-900/50 text-green-200";
      case "starting":
        return "bg-yellow-900/50 text-yellow-200";
      case "stopping":
        return "bg-orange-900/50 text-orange-200";
      case "stopped":
        return "bg-zinc-900/50 text-zinc-200";
      case "failed":
        return "bg-red-900/50 text-red-200";
      default:
        return "bg-zinc-900/50 text-zinc-200";
    }
  };

  const formatUptime = (startedAt: number | null) => {
    if (!startedAt) return "-";
    const minutes = Math.round((Date.now() - startedAt) / 60000);
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.round(minutes / 60);
    return `${hours}h`;
  };

  return (
    <div className="space-y-4">
      {workspaces && workspaces.length > 0 ? (
        <div className="space-y-2">
          {workspaces.map((ws) => (
            <div
              key={ws.workspaceId}
              className="flex items-center justify-between rounded-lg border border-zinc-700 bg-zinc-900 p-4"
            >
              <div className="flex-1">
                <h3 className="font-semibold">{ws.repoName}</h3>
                <div className="mt-1 flex items-center gap-4 text-sm text-zinc-400">
                  <span className={`rounded px-2 py-1 text-xs ${getStatusColor(ws.status)}`}>
                    {ws.status}
                  </span>
                  {ws.startedAt && (
                    <span>Uptime: {formatUptime(ws.startedAt)}</span>
                  )}
                </div>
              </div>

              <div className="flex gap-2">
                {ws.status === "running" && (
                  <button
                    onClick={() => handleStop(ws.workspaceId)}
                    className="rounded bg-red-900/30 px-3 py-1 text-sm text-red-200 hover:bg-red-900/50"
                  >
                    Stop
                  </button>
                )}
                {ws.vmId && (
                  <button
                    disabled
                    className="rounded bg-blue-900/30 px-3 py-1 text-sm text-blue-200"
                  >
                    Connect
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="rounded-lg border border-zinc-700 bg-zinc-900 p-8 text-center">
          <p className="text-zinc-400">No workspaces yet</p>
          <p className="mt-2 text-sm text-zinc-500">
            Click "Launch Workspace" on a repository to get started
          </p>
        </div>
      )}
    </div>
  );
}
