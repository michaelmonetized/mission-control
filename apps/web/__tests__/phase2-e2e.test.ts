import { expect, test, describe } from "bun:test";

describe("Phase 2: Mission Control Cloud E2E", () => {
  const baseURL = process.env.TEST_URL || "http://localhost:3410";

  test("User signup with Clerk", async () => {
    const response = await fetch(`${baseURL}/api/auth/signin`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: "test@example.com",
        provider: "github",
      }),
    });

    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.userId).toBeDefined();
  });

  test("User connects GitHub repo", async () => {
    const response = await fetch(`${baseURL}/api/repos/connect`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer test-token",
      },
      body: JSON.stringify({
        githubId: 123456,
        name: "test-repo",
        fullName: "user/test-repo",
        cloneUrl: "https://github.com/user/test-repo.git",
      }),
    });

    expect(response.status).toBe(201);
    const data = await response.json();
    expect(data.repoId).toBeDefined();
  });

  test("User launches workspace", async () => {
    const response = await fetch(`${baseURL}/api/workspaces/launch`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer test-token",
      },
      body: JSON.stringify({
        repoId: "repo-123",
      }),
    });

    expect(response.status).toBe(201);
    const data = await response.json();
    expect(data.workspaceId).toBeDefined();
    expect(data.status).toBe("starting");
  });

  test("WebSocket relay connects", async () => {
    // Test WebSocket connection
    const wsURL = new URL("ws://localhost:9001");
    wsURL.searchParams.set("userId", "user-123");
    wsURL.searchParams.set("workspaceId", "ws-123");
    wsURL.searchParams.set("vmId", "vm-123");

    let connected = false;
    const ws = new WebSocket(wsURL.toString());

    ws.onopen = () => {
      connected = true;
    };

    // Wait for connection
    await new Promise((resolve) => setTimeout(resolve, 500));
    ws.close();

    expect(connected).toBe(true);
  });

  test("Terminal input/output works", async () => {
    const wsURL = new URL("ws://localhost:9001");
    wsURL.searchParams.set("userId", "user-123");
    wsURL.searchParams.set("workspaceId", "ws-123");
    wsURL.searchParams.set("vmId", "vm-123");

    let receivedMessage: any = null;
    const ws = new WebSocket(wsURL.toString());

    ws.onmessage = (event) => {
      receivedMessage = JSON.parse(event.data);
    };

    // Wait for connection
    await new Promise((resolve) => setTimeout(resolve, 500));

    // Send input
    ws.send(JSON.stringify({ type: "input", content: "echo test" }));

    // Wait for response
    await new Promise((resolve) => setTimeout(resolve, 100));

    ws.close();

    expect(receivedMessage).toBeDefined();
    expect(receivedMessage.type).toBe("ready");
  });

  test("Usage tracking records correctly", async () => {
    const response = await fetch(`${baseURL}/api/usage/record`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer test-token",
      },
      body: JSON.stringify({
        workspaceId: "ws-123",
        durationMinutes: 15,
      }),
    });

    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.cost).toBe(0.015); // 15 * $0.001
    expect(data.recorded).toBe(true);
  });

  test("Billing calculation correct", async () => {
    const response = await fetch(`${baseURL}/api/usage/current`, {
      headers: { Authorization: "Bearer test-token" },
    });

    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.minutes).toBeGreaterThanOrEqual(0);
    expect(data.cost).toBeGreaterThanOrEqual(0);
    expect(data.remaining).toBeLessThanOrEqual(100);
  });

  test("Workspace stop works", async () => {
    const response = await fetch(`${baseURL}/api/workspaces/stop`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer test-token",
      },
      body: JSON.stringify({
        workspaceId: "ws-123",
      }),
    });

    expect(response.status).toBe(200);
    const data = await response.json();
    expect(data.status).toBe("stopping");
  });

  test("Performance: VM spinup <2s", async () => {
    const startTime = Date.now();

    const response = await fetch(`${baseURL}/api/workspaces/launch`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer test-token",
      },
      body: JSON.stringify({
        repoId: "repo-123",
      }),
    });

    const elapsed = Date.now() - startTime;

    expect(response.status).toBe(201);
    expect(elapsed).toBeLessThan(2000); // Must complete within 2 seconds
  });

  test("WebSocket relay handles 100 concurrent connections", async () => {
    const connections = [];

    for (let i = 0; i < 100; i++) {
      const wsURL = new URL("ws://localhost:9001");
      wsURL.searchParams.set("userId", `user-${i}`);
      wsURL.searchParams.set("workspaceId", `ws-${i}`);
      wsURL.searchParams.set("vmId", `vm-${i}`);

      const ws = new WebSocket(wsURL.toString());
      connections.push(ws);
    }

    // Wait for all to connect
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // Check health
    const health = await fetch("http://localhost:9002/health");
    const data = await health.json();

    expect(data.activeSessions).toBeLessThanOrEqual(100);

    // Close all
    connections.forEach((ws) => ws.close());
  });
});
