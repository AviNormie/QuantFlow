package hub

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Hub manages active websocket clients and symbol subscriptions.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	h.clients[client.id] = client
	h.mu.Unlock()
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	delete(h.clients, client.id)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(symbol string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.IsSubscribed(symbol) {
			client.TrySend(payload)
		}
	}
}

// Client represents a single websocket connection.
type Client struct {
	id      string
	conn    *websocket.Conn
	hub     *Hub
	symbols map[string]bool
	send    chan []byte
	mu      sync.RWMutex
	closed  bool
}

func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		id:      uuid.NewString(),
		conn:    conn,
		hub:     hub,
		symbols: make(map[string]bool),
		send:    make(chan []byte, 64),
	}
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Subscribe(symbol string) {
	c.mu.Lock()
	c.symbols[symbol] = true
	c.mu.Unlock()
}

func (c *Client) Unsubscribe(symbol string) {
	c.mu.Lock()
	delete(c.symbols, symbol)
	c.mu.Unlock()
}

func (c *Client) IsSubscribed(symbol string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.symbols[symbol]
}

func (c *Client) TrySend(payload []byte) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()

	select {
	case c.send <- payload:
	default:
	}
}

func (c *Client) WritePump() {
	for payload := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			break
		}
	}
}

func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

	close(c.send)
	c.conn.Close()
}
