import Bun from "bun";

interface TerminalSession {
  userId: string;
  workspaceId: string;
  vmId: string;
  ws: WebSocket;
  createdAt: number;
  lastActivityAt: number;
}

const sessions = new Map<string, TerminalSession>();

// WebSocket server on port 9001 for browser terminal connections
const server = Bun.serve<TerminalSession>({
  port: 9001,
  fetch(req, server) {
    const url = new URL(req.url);
    const sessionKey = url.searchParams.get("session");
    const userId = url.searchParams.get("userId");
    const workspaceId = url.searchParams.get("workspaceId");
    const vmId = url.searchParams.get("vmId");

    if (!sessionKey || !userId || !workspaceId || !vmId) {
      return new Response("Missing required parameters", { status: 400 });
    }

    // Upgrade to WebSocket
    if (server.upgrade(req, { data: { userId, workspaceId, vmId, sessionKey } })) {
      return undefined; // Handled by upgrade
    }

    return new Response("WebSocket upgrade failed", { status: 400 });
  },

  websocket: {
    open(ws) {
      const { userId, workspaceId, vmId, sessionKey } = ws.data;
      const terminalKey = `${userId}:${workspaceId}`;

      // Store session
      const session: TerminalSession = {
        userId,
        workspaceId,
        vmId,
        ws: ws as any,
        createdAt: Date.now(),
        lastActivityAt: Date.now(),
      };

      sessions.set(terminalKey, session);

      console.log(`[OPEN] Terminal session: ${terminalKey}`);

      // Send initial message
      ws.send(
        JSON.stringify({
          type: "ready",
          message: "Connected to workspace terminal",
          vmId,
        })
      );
    },

    message(ws, message) {
      const { userId, workspaceId, vmId } = ws.data;
      const terminalKey = `${userId}:${workspaceId}`;

      if (typeof message !== "string") return;

      const session = sessions.get(terminalKey);
      if (!session) return;

      // Update activity
      session.lastActivityAt = Date.now();

      try {
        const data = JSON.parse(message);

        switch (data.type) {
          case "input":
            // User typed in terminal
            // TODO: Send to VM via SSH/socket
            console.log(`[INPUT] ${terminalKey}: ${data.content?.substring(0, 50)}`);
            break;

          case "resize":
            // Terminal resized (cols, rows)
            console.log(`[RESIZE] ${terminalKey}: ${data.cols}x${data.rows}`);
            break;

          case "ping":
            // Keep-alive
            ws.send(JSON.stringify({ type: "pong" }));
            break;

          default:
            console.warn(`[UNKNOWN] ${terminalKey}: ${data.type}`);
        }
      } catch (err) {
        console.error(`[ERROR] ${terminalKey}: ${err}`);
      }
    },

    close(ws) {
      const { userId, workspaceId } = ws.data;
      const terminalKey = `${userId}:${workspaceId}`;

      sessions.delete(terminalKey);
      console.log(`[CLOSE] Terminal session: ${terminalKey}`);
    },

    error(ws, error) {
      const { userId, workspaceId } = ws.data;
      const terminalKey = `${userId}:${workspaceId}`;

      console.error(`[ERROR] ${terminalKey}: ${error?.message}`);
      sessions.delete(terminalKey);
    },
  },
});

// Health endpoint
Bun.serve({
  port: 9002,
  fetch(req) {
    if (new URL(req.url).pathname === "/health") {
      return new Response(
        JSON.stringify({
          status: "ok",
          service: "websocket-relay",
          activeSessions: sessions.size,
          uptime: process.uptime(),
        }),
        { headers: { "Content-Type": "application/json" } }
      );
    }
    return new Response("Not found", { status: 404 });
  },
});

console.log("WebSocket relay running on ws://localhost:9001");
console.log("Health check on http://localhost:9002/health");

// Cleanup idle sessions (>1h inactive)
setInterval(() => {
  const now = Date.now();
  const idleThreshold = 1000 * 60 * 60; // 1 hour

  for (const [key, session] of sessions) {
    if (now - session.lastActivityAt > idleThreshold) {
      console.log(`[CLEANUP] Closing idle session: ${key}`);
      session.ws.close();
      sessions.delete(key);
    }
  }
}, 60000); // Check every minute
