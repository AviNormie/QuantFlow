package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handlePricesWS(c *gin.Context) {
	symbol := strings.ToUpper(strings.TrimSpace(c.Query("symbol")))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	apiKey := os.Getenv("FINNHUB_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FINNHUB_API_KEY not set"})
		return
	}

	clientConn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}
	defer clientConn.Close()

	finnhubConn, _, err := websocket.DefaultDialer.Dial("wss://ws.finnhub.io?token="+apiKey, nil)
	if err != nil {
		log.Printf("finnhub connect: %v", err)
		return
	}
	defer finnhubConn.Close()

	subMsg, _ := json.Marshal(map[string]string{
		"type":   "subscribe",
		"symbol": symbol,
	})
	if err := finnhubConn.WriteMessage(websocket.TextMessage, subMsg); err != nil {
		log.Printf("finnhub subscribe: %v", err)
		return
	}

	go func() {
		for {
			_, msg, err := finnhubConn.ReadMessage()
			if err != nil {
				clientConn.Close()
				return
			}
			if err := clientConn.WriteMessage(websocket.TextMessage, msg); err != nil {
				finnhubConn.Close()
				return
			}
		}
	}()

	for {
		if _, _, err := clientConn.ReadMessage(); err != nil {
			return
		}
	}
}
