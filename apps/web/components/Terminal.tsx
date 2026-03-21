"use client";

import { useEffect, useRef } from "react";
import { useTerminal } from "@/hooks/useTerminal";
import { useWorkspaceStatus } from "@/hooks/useWorkspaceStatus";
import { Id } from "@/convex/_generated/dataModel";

interface TerminalProps {
  userId: string;
  workspaceId: Id<"workspaces">;
  vmId: string;
}

export default function Terminal({ userId, workspaceId, vmId }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const outputRef = useRef<string>("");
  const { connected, loading, send, resize } = useTerminal({
    userId,
    workspaceId: workspaceId as any,
    vmId,
    onOutput: (data) => {
      outputRef.current += data;
      if (terminalRef.current) {
        terminalRef.current.textContent = outputRef.current;
        terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
      }
    },
    onError: (error) => {
      console.error("[Terminal] Error:", error);
      if (terminalRef.current) {
        terminalRef.current.textContent += `\r\nError: ${error}\r\n`;
      }
    },
  });

  const status = useWorkspaceStatus(workspaceId);

  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (!connected) return;

      // Don't send meta keys
      if (e.ctrlKey && e.key === "c") {
        send("\u0003"); // Ctrl+C
        return;
      }

      if (e.ctrlKey && e.key === "d") {
        send("\u0004"); // Ctrl+D
        return;
      }

      // Regular character input
      if (e.key.length === 1) {
        send(e.key);
      } else if (e.key === "Enter") {
        send("\r\n");
      } else if (e.key === "Backspace") {
        send("\b");
      }
    };

    const terminal = terminalRef.current;
    if (terminal) {
      terminal.addEventListener("keydown", handleKeyPress);
      terminal.focus();
    }

    return () => {
      if (terminal) {
        terminal.removeEventListener("keydown", handleKeyPress);
      }
    };
  }, [connected, send]);

  if (loading) {
    return (
      <div className="flex h-96 items-center justify-center rounded-lg border border-zinc-700 bg-black">
        <div className="text-center">
          <div className="mb-4 h-8 w-8 animate-spin rounded-full border-4 border-zinc-700 border-t-zinc-50"></div>
          <p className="text-sm text-zinc-400">Connecting to workspace...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Status Bar */}
      <div className="flex items-center justify-between rounded-lg border border-zinc-700 bg-zinc-900 px-4 py-2">
        <div className="flex items-center gap-2">
          <div
            className={`h-2 w-2 rounded-full ${
              connected ? "bg-green-500" : "bg-red-500"
            }`}
          />
          <span className="text-sm text-zinc-300">
            {connected ? "Connected" : "Disconnected"}
          </span>
        </div>
        <div className="text-sm text-zinc-400">
          Uptime: {Math.floor(status.uptime / 60)}h {status.uptime % 60}m
        </div>
      </div>

      {/* Terminal */}
      <div
        ref={terminalRef}
        className="h-96 overflow-auto rounded-lg border border-zinc-700 bg-black p-4 font-mono text-sm text-green-400"
        style={{
          whiteSpace: "pre-wrap",
          wordWrap: "break-word",
          cursor: "text",
        }}
        tabIndex={0}
      />

      {/* Resource Usage */}
      <div className="grid grid-cols-2 gap-4">
        <div className="rounded-lg border border-zinc-700 bg-zinc-900 p-3">
          <div className="text-xs text-zinc-400">CPU</div>
          <div className="mt-1 text-lg font-semibold">{status.cpuUsage.toFixed(1)}%</div>
        </div>
        <div className="rounded-lg border border-zinc-700 bg-zinc-900 p-3">
          <div className="text-xs text-zinc-400">Memory</div>
          <div className="mt-1 text-lg font-semibold">{status.memoryUsage.toFixed(0)}MB</div>
        </div>
      </div>
    </div>
  );
}
