package model

// NormalizedTick is the canonical market tick used inside StockFlow.
type NormalizedTick struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp"`
	Source    string  `json:"source"`
}

// SymbolInfo describes a tradable symbol for charting clients.
type SymbolInfo struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Exchange    string `json:"exchange"`
	Type        string `json:"type"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

// Candle is a single OHLC bar.
type Candle struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume,omitempty"`
}
