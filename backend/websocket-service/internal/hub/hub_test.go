package hub_test

import (
	"testing"

	"websocket-service/internal/hub"

	"github.com/gorilla/websocket"
)

type fakeConn struct{}

func (f *fakeConn) WriteMessage(messageType int, data []byte) error { return nil }
func (f *fakeConn) Close() error                                    { return nil }

func TestHubBroadcastFiltersSubscriptions(t *testing.T) {
	h := hub.NewHub()

	connA := &websocket.Conn{}
	connB := &websocket.Conn{}

	clientA := hub.NewClient(connA, h)
	clientB := hub.NewClient(connB, h)
	clientA.Subscribe("AAPL")
	clientB.Subscribe("MSFT")

	h.Register(clientA)
	h.Register(clientB)

	payload := []byte(`{"type":"trade"}`)
	h.Broadcast("AAPL", payload)

	if !clientA.IsSubscribed("AAPL") {
		t.Fatal("client A should be subscribed to AAPL")
	}
}

func TestClientSubscribeUnsubscribe(t *testing.T) {
	h := hub.NewHub()
	client := hub.NewClient(&websocket.Conn{}, h)
	client.Subscribe("AAPL")
	if !client.IsSubscribed("AAPL") {
		t.Fatal("expected subscription")
	}
	client.Unsubscribe("AAPL")
	if client.IsSubscribed("AAPL") {
		t.Fatal("expected unsubscribe")
	}
}
