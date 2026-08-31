package hub

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 4096
	sendBuffer = 256
)

// ConfigureConn sets read limits and pong handling for keepalive.
func ConfigureConn(conn *websocket.Conn) {
	conn.SetReadLimit(maxMsgSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
}

// TouchReadDeadline extends the read deadline after a client message.
func TouchReadDeadline(conn *websocket.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
}
