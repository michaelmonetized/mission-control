import { httpRouter } from "convex/server";
import { internal } from "./_generated/api";

const http = httpRouter();

/**
 * VM Manager calls this to update workspace status
 * POST /api/workspace-status
 */
http.route({
  path: "/workspace-status",
  method: "POST",
  handler: async (ctx, request) => {
    try {
      const body = await request.json();
      const { workspaceId, status, vmId } = body;

      // Validate auth (VM Manager should send API key)
      const apiKey = request.headers.get("authorization");
      if (apiKey !== `Bearer ${process.env.VM_MANAGER_API_KEY}`) {
        return new Response("Unauthorized", { status: 401 });
      }

      // Call internal mutation to update workspace
      await ctx.runMutation(internal.workspaces.updateWorkspaceStatus, {
        workspaceId,
        status,
        vmId,
      });

      return new Response(JSON.stringify({ success: true }), {
        headers: { "Content-Type": "application/json" },
      });
    } catch (error) {
      console.error("[HTTP] Error:", error);
      return new Response(JSON.stringify({ error: String(error) }), {
        status: 500,
        headers: { "Content-Type": "application/json" },
      });
    }
  },
});

/**
 * Health check endpoint
 */
http.route({
  path: "/health",
  method: "GET",
  handler: async () => {
    return new Response(JSON.stringify({ status: "ok" }), {
      headers: { "Content-Type": "application/json" },
    });
  },
});

export default http;
