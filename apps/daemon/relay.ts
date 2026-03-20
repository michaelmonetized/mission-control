#!/usr/bin/env bun

/**
 * Mission Control Daemon Relay
 * 
 * Runs on Rusty's m1pro (192.168.1.23) and Theo's m1pro-13
 * 
 * Purpose:
 * - Listens on local port 9999
 * - Relays messages from Mission Control → OpenClaw gateway
 * - Relays responses from OpenClaw → back to Mission Control
 * - Handles connection pooling and message queuing
 * - Logs all messages for auditing
 * 
 * Usage:
 *   bun relay.ts [port] [openclaw-url]
 * 
 * Example:
 *   bun relay.ts 9999 ws://192.168.1.134:18789
 */

import { WebSocketServer, WebSocket } from 'ws';
import * as fs from 'fs';
import * as path from 'path';

const PORT = parseInt(process.argv[2] || '9999');
const OPENCLAW_URL = process.argv[3] || 'ws://192.168.1.134:18789';
const HOSTNAME = process.env.HOSTNAME || 'unknown-host';

console.log(`🚀 Mission Control Daemon Relay`);
console.log(`   Listening on: localhost:${PORT}`);
console.log(`   OpenClaw Gateway: ${OPENCLAW_URL}`);
console.log(`   Host: ${HOSTNAME}\n`);

// Message queue for reliability
const messageQueue: any[] = [];
let openclawConnection: WebSocket | null = null;

// Create WebSocket server
const wss = new WebSocketServer({ port: PORT });

interface RelayMessage {
  id: string;
  type: 'message' | 'status' | 'error';
  from: 'mission-control' | 'openclaw';
  sessionId: string;
  threadId?: string;
  content: string;
  timestamp: Date;
  retries?: number;
}

/**
 * Connect to OpenClaw gateway
 */
function connectToOpenClaw() {
  console.log(`📡 Connecting to OpenClaw at ${OPENCLAW_URL}...`);

  openclawConnection = new WebSocket(OPENCLAW_URL);

  openclawConnection.on('open', () => {
    console.log(`✅ Connected to OpenClaw`);

    // Drain message queue
    while (messageQueue.length > 0) {
      const msg = messageQueue.shift();
      if (msg) {
        openclawConnection?.send(JSON.stringify(msg));
      }
    }
  });

  openclawConnection.on('message', (data) => {
    try {
      const message: RelayMessage = JSON.parse(data.toString());
      logMessage(message, 'OPENCLAW_RECV');

      // Broadcast to all connected mission-control clients
      wss.clients.forEach((client) => {
        if (client.readyState === WebSocket.OPEN) {
          client.send(JSON.stringify(message));
        }
      });
    } catch (error) {
      console.error('Error parsing OpenClaw message:', error);
    }
  });

  openclawConnection.on('close', () => {
    console.warn(`⚠️  Disconnected from OpenClaw. Reconnecting in 5s...`);
    openclawConnection = null;
    setTimeout(connectToOpenClaw, 5000);
  });

  openclawConnection.on('error', (error) => {
    console.error('OpenClaw connection error:', error);
  });
}

/**
 * Log message to file
 */
function logMessage(message: RelayMessage, direction: string): void {
  const timestamp = new Date().toISOString();
  const logDir = path.join(process.env.HOME!, '.hurleyus', 'daemon-logs');

  fs.mkdirSync(logDir, { recursive: true });

  const logFile = path.join(logDir, `relay-${new Date().toISOString().split('T')[0]}.log`);
  const logEntry = `[${timestamp}] ${direction} | ${message.type} | Session: ${message.sessionId} | ${message.content.slice(0, 100)}...\n`;

  fs.appendFileSync(logFile, logEntry);
}

/**
 * Relay message to OpenClaw
 */
function relayToOpenClaw(message: RelayMessage): void {
  logMessage(message, 'MC_SEND');

  if (openclawConnection && openclawConnection.readyState === WebSocket.OPEN) {
    openclawConnection.send(JSON.stringify(message));
  } else {
    console.warn(`⚠️  OpenClaw not connected. Queuing message...`);
    messageQueue.push(message);

    // If queue gets too large, drop oldest messages
    if (messageQueue.length > 100) {
      messageQueue.shift();
    }
  }
}

/**
 * Handle client connection
 */
wss.on('connection', (ws, req) => {
  const clientId = req.headers['x-client-id'] || `client-${Math.random().toString(36).slice(2)}`;
  console.log(`👤 Client connected: ${clientId}`);

  ws.on('message', (data) => {
    try {
      const message: RelayMessage = JSON.parse(data.toString());
      message.timestamp = new Date();

      // Relay to OpenClaw
      relayToOpenClaw(message);
    } catch (error) {
      console.error('Error processing client message:', error);
      ws.send(JSON.stringify({
        type: 'error',
        content: 'Failed to process message',
      }));
    }
  });

  ws.on('close', () => {
    console.log(`👤 Client disconnected: ${clientId}`);
  });

  ws.on('error', (error) => {
    console.error(`Client error (${clientId}):`, error);
  });

  // Send welcome message
  ws.send(JSON.stringify({
    type: 'status',
    content: `Connected to relay on ${HOSTNAME}. Ready to relay messages.`,
  }));
});

/**
 * Health check endpoint (via HTTP)
 */
if (Bun.env.HTTP_HEALTH_CHECK === 'true') {
  Bun.serve({
    port: PORT + 1000,
    async fetch(req) {
      if (req.url.endsWith('/health')) {
        return new Response(
          JSON.stringify({
            status: 'ok',
            relay: 'healthy',
            hostname: HOSTNAME,
            connectedClients: wss.clients.size,
            openclawConnected: openclawConnection?.readyState === WebSocket.OPEN,
            queueSize: messageQueue.length,
          }),
          { headers: { 'Content-Type': 'application/json' } }
        );
      }
      return new Response('Not found', { status: 404 });
    },
  });

  console.log(`🏥 Health check endpoint on localhost:${PORT + 1000}/health`);
}

// Connect to OpenClaw immediately
connectToOpenClaw();

// Graceful shutdown
process.on('SIGINT', () => {
  console.log('\n🛑 Shutting down relay gracefully...');
  openclawConnection?.close();
  wss.close(() => {
    console.log('✅ Relay shutdown complete');
    process.exit(0);
  });
});

console.log(`\n✅ Relay ready. Waiting for connections...`);
