package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages WebSocket connections grouped by channel name.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]struct{}
}

func newHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) register(channel string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[channel] == nil {
		h.clients[channel] = make(map[*websocket.Conn]struct{})
	}
	h.clients[channel][conn] = struct{}{}
}

func (h *Hub) unregister(channel string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[channel], conn)
	if len(h.clients[channel]) == 0 {
		delete(h.clients, channel)
	}
}

func (h *Hub) broadcast(channel string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients[channel] {
		_ = conn.WriteMessage(websocket.TextMessage, msg)
	}
}
