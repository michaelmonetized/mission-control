package relay

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type TerminalRelay struct {
	upgrader websocket.Upgrader
	clients  map[string]*TerminalClient
	clientMu sync.RWMutex

	// Circuit breaker for failed connections
	failedConnections map[string]int
	failedMu          sync.RWMutex

	maxConnections int
}

type TerminalClient struct {
	ID           string
	VMID         string
	ClientConn   *websocket.Conn
	VMConn       *websocket.Conn
	LastActivity time.Time
	CreatedAt    time.Time
	BytesSent    int64
	BytesReceived int64
}

type Message struct {
	Type string      `json:"type"` // "data", "control", "ping", "pong"
	Data string      `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
}

func NewTerminalRelay(maxConnections int) *TerminalRelay {
	return &TerminalRelay{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:           make(map[string]*TerminalClient),
		failedConnections: make(map[string]int),
		maxConnections:    maxConnections,
	}
}

// HandleConnection upgrades HTTP connection to WebSocket
func (tr *TerminalRelay) HandleConnection(w http.ResponseWriter, r *http.Request) {
	vmID := r.URL.Query().Get("vm_id")
	clientID := r.URL.Query().Get("client_id")

	if vmID == "" || clientID == "" {
		http.Error(w, "Missing vm_id or client_id", http.StatusBadRequest)
		return
	}

	// Check connection limit
	tr.clientMu.RLock()
	if len(tr.clients) >= tr.maxConnections {
		tr.clientMu.RUnlock()
		http.Error(w, "Server at capacity", http.StatusServiceUnavailable)
		return
	}
	tr.clientMu.RUnlock()

	// Upgrade connection
	clientConn, err := tr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed for %s: %v", clientID, err)
		return
	}
	defer clientConn.Close()

	// Create client
	client := &TerminalClient{
		ID:        clientID,
		VMID:      vmID,
		ClientConn: clientConn,
		CreatedAt: time.Now(),
	}

	// Register client
	tr.clientMu.Lock()
	tr.clients[clientID] = client
	tr.clientMu.Unlock()

	log.Printf("✅ Client %s connected to VM %s", clientID, vmID)

	// Handle messages
	tr.handleClient(client)

	// Cleanup
	tr.clientMu.Lock()
	delete(tr.clients, clientID)
	tr.clientMu.Unlock()
	log.Printf("❌ Client %s disconnected", clientID)
}

func (tr *TerminalRelay) handleClient(client *TerminalClient) {
	// Set read deadline
	client.ClientConn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.ClientConn.SetPongHandler(func(string) error {
		client.ClientConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastActivity = time.Now()
		return nil
	})

	// Ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Message pump
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := client.ClientConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Printf("⚠️  Ping failed for %s: %v", client.ID, err)
					return
				}
			}
		}
	}()

	for {
		var msg Message
		err := client.ClientConn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("⚠️  WebSocket error: %v", err)
			}
			return
		}

		client.LastActivity = time.Now()
		client.BytesReceived += int64(len(msg.Data))

		// Process message
		switch msg.Type {
		case "data":
			// Echo back or relay to VM
			response := Message{
				Type: "data_ack",
				Data: fmt.Sprintf("Received: %s", msg.Data),
			}
			if err := client.ClientConn.WriteJSON(response); err != nil {
				log.Printf("⚠️  Write failed for %s: %v", client.ID, err)
				return
			}
			client.BytesSent += int64(len(response.Data))

		case "control":
			// Handle terminal control sequences
			log.Printf("📝 Control message from %s: %s", client.ID, msg.Data)

		case "ping":
			response := Message{
				Type: "pong",
				Meta: map[string]interface{}{
					"timestamp": time.Now().Unix(),
				},
			}
			if err := client.ClientConn.WriteJSON(response); err != nil {
				log.Printf("⚠️  Pong failed for %s: %v", client.ID, err)
				return
			}
		}
	}
}

// GetClientInfo returns info about a connected client
func (tr *TerminalRelay) GetClientInfo(clientID string) *TerminalClient {
	tr.clientMu.RLock()
	defer tr.clientMu.RUnlock()
	return tr.clients[clientID]
}

// ListClients returns all connected clients for a VM
func (tr *TerminalRelay) ListClients(vmID string) []*TerminalClient {
	tr.clientMu.RLock()
	defer tr.clientMu.RUnlock()

	var result []*TerminalClient
	for _, client := range tr.clients {
		if client.VMID == vmID {
			result = append(result, client)
		}
	}
	return result
}

// ClientCount returns the total number of connected clients
func (tr *TerminalRelay) ClientCount() int {
	tr.clientMu.RLock()
	defer tr.clientMu.RUnlock()
	return len(tr.clients)
}

// DisconnectClient forcefully disconnects a client
func (tr *TerminalRelay) DisconnectClient(clientID string) error {
	tr.clientMu.Lock()
	client, exists := tr.clients[clientID]
	if !exists {
		tr.clientMu.Unlock()
		return fmt.Errorf("client %s not found", clientID)
	}
	delete(tr.clients, clientID)
	tr.clientMu.Unlock()

	return client.ClientConn.Close()
}

// CleanupStaleConnections removes connections idle for longer than maxIdleDuration
func (tr *TerminalRelay) CleanupStaleConnections(maxIdleDuration time.Duration) {
	tr.clientMu.Lock()
	defer tr.clientMu.Unlock()

	now := time.Now()
	for clientID, client := range tr.clients {
		if now.Sub(client.LastActivity) > maxIdleDuration {
			log.Printf("🧹 Closing idle connection: %s", clientID)
			client.ClientConn.Close()
			delete(tr.clients, clientID)
		}
	}
}
