package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"shared/jwt"
	"websocket-service/internal/hub"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub            *hub.Hub
	pubSubChannel    string
	requireAuth      bool
}

func NewWSHandler(h *hub.Hub, pubSubChannel string, requireAuth bool) *WSHandler {
	return &WSHandler{hub: h, pubSubChannel: pubSubChannel, requireAuth: requireAuth}
}

func (h *WSHandler) Handle(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}
	}

	if h.requireAuth {
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			return
		}
		if _, err := jwt.VerifyAccessToken(token); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
	} else if token != "" {
		if _, err := jwt.VerifyAccessToken(token); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}

	client := hub.NewClient(conn, h.hub)
	h.hub.Register(client)

	go client.WritePump()
	defer func() {
		h.hub.Unregister(client)
		client.Close()
	}()

	// Legacy query-param subscription support.
	if symbol := strings.ToUpper(strings.TrimSpace(c.Query("symbol"))); symbol != "" {
		client.Subscribe(symbol)
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var payload struct {
			Action  string   `json:"action"`
			Symbols []string `json:"symbols"`
			Symbol  string   `json:"symbol"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}

		symbols := payload.Symbols
		if payload.Symbol != "" {
			symbols = append(symbols, payload.Symbol)
		}

		switch strings.ToLower(payload.Action) {
		case "subscribe":
			for _, symbol := range symbols {
				client.Subscribe(strings.ToUpper(strings.TrimSpace(symbol)))
			}
		case "unsubscribe":
			for _, symbol := range symbols {
				client.Unsubscribe(strings.ToUpper(strings.TrimSpace(symbol)))
			}
		}
	}
}
