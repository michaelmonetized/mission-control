import { httpRouter } from "convex/server";
import { recordWebhookEvent } from "./mutations/threads";
import { getThreadCount, listThreads } from "./queries/threads";
import { getCurrentUser, getCurrentUsage } from "./queries/users";
import { v } from "convex/values";

const http = httpRouter();

/**
 * Health check endpoint
 */
http.route({
  path: "/health",
  method: "GET",
  handler: async () => {
    return new Response(
      JSON.stringify({
        status: "ok",
        timestamp: new Date().toISOString(),
      }),
      {
        headers: { "Content-Type": "application/json" },
      }
    );
  },
});

/**
 * Stripe webhook endpoint
 * Handles billing events
 */
http.route({
  path: "/webhooks/stripe",
  method: "POST",
  handler: async (request, ctx) => {
    const body = await request.text();
    const signature = request.headers.get("stripe-signature");

    // TODO: Verify Stripe signature using webhook secret
    // For now, just accept and log

    try {
      const event = JSON.parse(body);

      // Record webhook event
      await ctx.runMutation(recordWebhookEvent, {
        event: "stripe",
        action: event.type,
        payload: event,
      });

      // Handle different event types
      switch (event.type) {
        case "invoice.payment_succeeded":
          // Payment succeeded
          break;
        case "invoice.payment_failed":
          // Payment failed
          break;
        case "customer.subscription.updated":
          // Subscription updated
          break;
      }

      return new Response(JSON.stringify({ received: true }), {
        headers: { "Content-Type": "application/json" },
      });
    } catch (error) {
      console.error("Stripe webhook error:", error);
      return new Response(JSON.stringify({ error: "Invalid payload" }), {
        status: 400,
        headers: { "Content-Type": "application/json" },
      });
    }
  },
});

/**
 * GitHub webhook endpoint
 * Handles PR/issue events
 */
http.route({
  path: "/webhooks/github",
  method: "POST",
  handler: async (request, ctx) => {
    const signature = request.headers.get("x-hub-signature-256");
    const event = request.headers.get("x-github-event");
    const body = await request.text();

    // TODO: Verify GitHub signature using webhook secret
    // For now, just accept and log

    try {
      const payload = JSON.parse(body);

      // Record webhook event
      await ctx.runMutation(recordWebhookEvent, {
        event: `github.${event}`,
        action: payload.action,
        payload,
      });

      return new Response(JSON.stringify({ received: true }), {
        headers: { "Content-Type": "application/json" },
      });
    } catch (error) {
      console.error("GitHub webhook error:", error);
      return new Response(JSON.stringify({ error: "Invalid payload" }), {
        status: 400,
        headers: { "Content-Type": "application/json" },
      });
    }
  },
});

/**
 * Statistics endpoint
 * Returns user stats (public)
 */
http.route({
  path: "/stats",
  method: "GET",
  handler: async () => {
    return new Response(
      JSON.stringify({
        timestamp: new Date().toISOString(),
        // We don't have access to query context in HTTP routes
        // This is mainly for public stats or monitoring
      }),
      {
        headers: { "Content-Type": "application/json" },
      }
    );
  },
});

export default http;
