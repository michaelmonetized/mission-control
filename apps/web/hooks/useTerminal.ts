import { useEffect, useRef, useState } from "react";

interface TerminalOptions {
  userId: string;
  workspaceId: string;
  vmId: string;
  onOutput?: (data: string) => void;
  onError?: (error: string) => void;
}

export function useTerminal({
  userId,
  workspaceId,
  vmId,
  onOutput,
  onError,
}: TerminalOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Connect to WebSocket relay
    const wsURL = new URL("ws://localhost:9001");
    wsURL.searchParams.set("userId", userId);
    wsURL.searchParams.set("workspaceId", workspaceId);
    wsURL.searchParams.set("vmId", vmId);
    wsURL.searchParams.set("session", `${userId}:${workspaceId}`);

    const ws = new WebSocket(wsURL.toString());

    ws.onopen = () => {
      console.log("[Terminal] Connected to relay");
      setConnected(true);
      setLoading(false);
    };

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);

        switch (message.type) {
          case "ready":
            console.log("[Terminal] Ready:", message.message);
            onOutput?.("\r\n" + message.message + "\r\n");
            break;

          case "output":
            // VM output data
            onOutput?.(message.content);
            break;

          case "pong":
            // Keep-alive response
            break;

          default:
            console.warn("[Terminal] Unknown message type:", message.type);
        }
      } catch (err) {
        console.error("[Terminal] Failed to parse message:", err);
      }
    };

    ws.onerror = (event) => {
      console.error("[Terminal] WebSocket error:", event);
      onError?.("Connection error");
      setConnected(false);
    };

    ws.onclose = () => {
      console.log("[Terminal] Disconnected from relay");
      setConnected(false);
    };

    wsRef.current = ws;

    // Keep-alive ping every 30 seconds
    const pingInterval = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "ping" }));
      }
    }, 30000);

    return () => {
      clearInterval(pingInterval);
      ws.close();
    };
  }, [userId, workspaceId, vmId, onOutput, onError]);

  const send = (input: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          type: "input",
          content: input,
        })
      );
    }
  };

  const resize = (cols: number, rows: number) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({
          type: "resize",
          cols,
          rows,
        })
      );
    }
  };

  return {
    connected,
    loading,
    send,
    resize,
  };
}
