package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	sendBufferSize  = 256
	writeWait       = 10 * time.Second
	pingInterval    = 30 * time.Second
	pongWait        = 60 * time.Second
)

// Client wraps a WebSocket connection with a non-blocking send channel.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	channel string
	send    chan []byte
}

// writePump drains the send channel to the WebSocket connection.
// Runs in its own goroutine — never blocks the hub.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump keeps the connection alive and detects disconnects.
// Runs in its own goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

// Hub manages all active clients grouped by channel.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
}

func newHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.channel] == nil {
		h.clients[c.channel] = make(map[*Client]struct{})
	}
	h.clients[c.channel][c] = struct{}{}
	log.Printf("+ client  channel=%s total=%d", c.channel, len(h.clients[c.channel]))
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c.channel][c]; ok {
		delete(h.clients[c.channel], c)
		close(c.send)
		if len(h.clients[c.channel]) == 0 {
			delete(h.clients, c.channel)
		}
	}
	log.Printf("- client  channel=%s", c.channel)
}

// broadcast sends msg to all clients on channel without blocking.
// Slow clients are dropped when their send buffer is full.
func (h *Hub) broadcast(channel string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[channel] {
		select {
		case c.send <- msg:
		default:
			// Buffer full — client too slow, disconnect it
			log.Printf("! slow client dropped  channel=%s", channel)
			go h.unregister(c)
		}
	}
}

// ConnectionCount returns total active WebSocket connections.
func (h *Hub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}
