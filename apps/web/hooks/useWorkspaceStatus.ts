import { useQuery, useMutation } from "convex/react";
import { api } from "@/convex/_generated/api";
import { useEffect, useState } from "react";
import { Id } from "@/convex/_generated/dataModel";

interface WorkspaceStatusUpdate {
  status: "stopped" | "starting" | "running" | "stopping" | "failed";
  uptime: number;
  cpuUsage: number;
  memoryUsage: number;
  lastUpdate: number;
}

export function useWorkspaceStatus(workspaceId: Id<"workspaces">) {
  const status = useQuery(api.queries.subscribeToWorkspaceStatus);
  const recordUsage = useMutation(api.mutations.recordUsage);
  const stopWorkspace = useMutation(api.mutations.stopWorkspace);

  const [autoStop, setAutoStop] = useState(false);

  // Auto-stop after 2 hours of inactivity
  useEffect(() => {
    if (status?.status === "running" && status.uptime > 120) {
      console.warn("[Workspace] Uptime exceeded 2 hours, stopping...");
      setAutoStop(true);
      stopWorkspace({ workspaceId });
    }
  }, [status?.uptime, status?.status, workspaceId, stopWorkspace]);

  // Record usage when stopping
  useEffect(() => {
    if (status?.status === "stopped" && status.uptime > 0 && !autoStop) {
      recordUsage({
        workspaceId,
        durationMinutes: status.uptime,
      }).catch(console.error);
    }
  }, [status?.status, status?.uptime, workspaceId, recordUsage, autoStop]);

  return {
    status: status?.status || "unknown",
    uptime: status?.uptime || 0,
    cpuUsage: status?.cpuUsage || 0,
    memoryUsage: status?.memoryUsage || 0,
    isHealthy: status?.status === "running",
  };
}
