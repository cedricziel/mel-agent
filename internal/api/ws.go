package api

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// Hub maintains active websocket clients for an agent and broadcasts messages to them.
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

var (
	hubs     = make(map[string]*Hub)
	hubsMu   sync.Mutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// GetHub returns the Hub for a given agentID, creating it if necessary.
func GetHub(agentID string) *Hub {
	hubsMu.Lock()
	defer hubsMu.Unlock()
	h, ok := hubs[agentID]
	if !ok {
		h = &Hub{clients: make(map[*websocket.Conn]bool)}
		hubs[agentID] = h
	}
	return h
}

// wsHandler upgrades HTTP connection to WS and registers the client.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	hub := GetHub(agentID)
	hub.addClient(conn)
	// Start reading messages from this client
	go hub.readPump(conn)
}

// addClient registers a websocket connection to the Hub.
func (h *Hub) addClient(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
}

// removeClient unregisters and closes a websocket connection.
func (h *Hub) removeClient(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	conn.Close()
}

// broadcast sends a message to all registered clients.
func (h *Hub) broadcast(message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}

// Broadcast sends a message to all registered clients (exported).
func (h *Hub) Broadcast(message []byte) {
	h.broadcast(message)
}

// readPump listens for incoming messages and broadcasts them to other clients.
func (h *Hub) readPump(conn *websocket.Conn) {
	defer h.removeClient(conn)
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if msgType != websocket.TextMessage {
			continue
		}
		h.broadcast(msg)
	}
}
