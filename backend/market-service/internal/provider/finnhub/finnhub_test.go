package finnhub_test

import (
	"encoding/json"
	"testing"
)

func TestTradePayloadParsing(t *testing.T) {
	raw := []byte(`{"type":"trade","data":[{"s":"AAPL","p":150.1,"v":5,"t":1700000000000}]}`)
	var payload struct {
		Type string `json:"type"`
		Data []struct {
			S string  `json:"s"`
			P float64 `json:"p"`
			V float64 `json:"v"`
			T int64   `json:"t"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Type != "trade" || len(payload.Data) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.Data[0].S != "AAPL" {
		t.Fatalf("unexpected symbol")
	}
}
