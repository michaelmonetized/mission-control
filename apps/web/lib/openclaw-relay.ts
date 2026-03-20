/**
 * OpenClaw Relay
 * 
 * Connects Mission Control threads to OpenClaw sessions
 * - Listens for messages in threads
 * - Relays to OpenClaw via WebSocket
 * - Receives responses and posts back to threads
 */

interface OpenClawMessage {
  type: 'message' | 'status' | 'error';
  sessionId: string;
  content: string;
  timestamp: Date;
  from: 'mission-control' | 'openclaw';
}

export class OpenClawRelay {
  private ws?: WebSocket;
  private sessionId: string;
  private gatewayUrl: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 3000;

  private handlers: Map<string, (message: OpenClawMessage) => void> = new Map();

  constructor(sessionId: string, gatewayUrl: string = 'ws://192.168.1.134:18789') {
    this.sessionId = sessionId;
    this.gatewayUrl = gatewayUrl;
  }

  /**
   * Connect to OpenClaw gateway
   */
  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(`${this.gatewayUrl}?session=${this.sessionId}`);

        this.ws.onopen = () => {
          console.log(`[OpenClawRelay] Connected to ${this.gatewayUrl}`);
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: OpenClawMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('[OpenClawRelay] Failed to parse message:', error);
          }
        };

        this.ws.onerror = (error) => {
          console.error('[OpenClawRelay] WebSocket error:', error);
          reject(error);
        };

        this.ws.onclose = () => {
          console.log('[OpenClawRelay] Disconnected');
          this.attemptReconnect();
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Send message to OpenClaw
   */
  async sendMessage(threadId: string, content: string): Promise<void> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('OpenClaw relay not connected');
    }

    const message: OpenClawMessage = {
      type: 'message',
      sessionId: this.sessionId,
      content,
      timestamp: new Date(),
      from: 'mission-control',
    };

    this.ws.send(JSON.stringify({
      ...message,
      threadId, // Include thread context
    }));
  }

  /**
   * Listen for messages from OpenClaw
   */
  onMessage(handler: (message: OpenClawMessage) => void): void {
    this.handlers.set('message', handler);
  }

  /**
   * Handle incoming message
   */
  private handleMessage(message: OpenClawMessage): void {
    const handler = this.handlers.get(message.type);
    if (handler) {
      handler(message);
    }
  }

  /**
   * Attempt reconnection
   */
  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[OpenClawRelay] Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    console.log(
      `[OpenClawRelay] Reconnecting (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`
    );

    setTimeout(() => {
      this.connect().catch((error) => {
        console.error('[OpenClawRelay] Reconnection failed:', error);
      });
    }, this.reconnectDelay);
  }

  /**
   * Disconnect
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
    }
  }
}

/**
 * Global relay instance
 */
let globalRelay: OpenClawRelay | null = null;

export function getOpenClawRelay(sessionId: string): OpenClawRelay {
  if (!globalRelay) {
    globalRelay = new OpenClawRelay(sessionId);
  }
  return globalRelay;
}

export function initOpenClawRelay(sessionId: string): Promise<void> {
  const relay = getOpenClawRelay(sessionId);
  return relay.connect();
}
