package admin

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gorilla/websocket"
)

const wsWriteTimeout = 5 * time.Second

// Hub keeps track of the connected websocket clients. It is safe for concurrent use.
type Hub struct {
	mu      sync.Mutex
	clients map[*wsClient]struct{}
}

type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex // gorilla/websocket supports only one concurrent writer
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*wsClient]struct{})}
}

func (h *Hub) add(c *wsClient) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) remove(c *wsClient) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	c.conn.Close()
}

// Count returns the number of connected clients
func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

// Broadcast sends a JSON message to all connected clients. Clients that fail are disconnected.
func (h *Hub) Broadcast(msg interface{}) {
	h.mu.Lock()
	clients := make([]*wsClient, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.Unlock()

	for _, c := range clients {
		c.mu.Lock()
		c.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
		err := c.conn.WriteJSON(msg)
		c.mu.Unlock()
		if err != nil {
			h.remove(c)
		}
	}
}

// CheckOrigin returns true if the Origin header of the request is the same host as the
// request itself, or one of the explicitly allowed origins
func CheckOrigin(r *http.Request, allowed []string) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Not a browser request (browsers always send Origin for websockets)
		return true
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if strings.EqualFold(u.Host, r.Host) {
		return true
	}

	for _, a := range allowed {
		if strings.EqualFold(strings.TrimRight(a, "/"), origin) {
			return true
		}
	}

	return false
}

// WShandler upgrades the request to a websocket connection. The client must provide a
// valid JWT in the "token" query parameter (browsers can not set headers on websockets).
// Messages sent by clients are ignored; the websocket is used for server push only.
func WShandler(w http.ResponseWriter, r *http.Request, hub *Hub, mw *jwt.GinJWTMiddleware, cfg *Config) {
	token, err := mw.ParseTokenString(r.URL.Query().Get("token"))
	if err != nil || !token.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return CheckOrigin(r, cfg.AllowedOrigins) },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &wsClient{conn: conn}
	hub.add(client)
	defer hub.remove(client)

	conn.SetReadLimit(4096)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
