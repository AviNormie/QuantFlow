package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"market-service/internal/model"
	"market-service/internal/provider"

	"github.com/gorilla/websocket"
)

const (
	wsURL    = "wss://ws.finnhub.io"
	restBase = "https://finnhub.io/api/v1"
)

// Provider streams Finnhub trades and exposes REST helpers.
type Provider struct {
	apiKey string
	client *http.Client

	conn   *websocket.Conn
	trades chan provider.RawTrade
	mu     sync.Mutex
}

func NewProvider(apiKey string) *Provider {
	return &Provider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 5 * time.Second},
		trades: make(chan provider.RawTrade, 256),
	}
}

func (p *Provider) Connect(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("FINNHUB_API_KEY is not set")
	}

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.DialContext(ctx, wsURL+"?token="+p.apiKey, nil)
	if err != nil {
		return fmt.Errorf("finnhub connect: %w", err)
	}

	p.mu.Lock()
	p.conn = conn
	p.mu.Unlock()

	go p.readLoop(conn)
	return nil
}

func (p *Provider) readLoop(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var payload struct {
			Type string `json:"type"`
			Data []struct {
				S string  `json:"s"`
				P float64 `json:"p"`
				V float64 `json:"v"`
				T int64   `json:"t"`
			} `json:"data"`
		}

		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}
		if payload.Type != "trade" {
			continue
		}

		for _, trade := range payload.Data {
			p.trades <- provider.RawTrade{
				Symbol:    strings.ToUpper(trade.S),
				Price:     trade.P,
				Volume:    trade.V,
				Timestamp: trade.T,
			}
		}
	}
}

func (p *Provider) Subscribe(symbols []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn == nil {
		return fmt.Errorf("not connected")
	}

	for _, symbol := range symbols {
		msg, _ := json.Marshal(map[string]string{
			"type":   "subscribe",
			"symbol": strings.ToUpper(symbol),
		})
		if err := p.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return err
		}
	}
	return nil
}

func (p *Provider) Trades() <-chan provider.RawTrade {
	return p.trades
}

func (p *Provider) Reconnect(ctx context.Context) error {
	p.Close()
	return p.Connect(ctx)
}

func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		err := p.conn.Close()
		p.conn = nil
		return err
	}
	return nil
}

func (p *Provider) SearchSymbols(ctx context.Context, query string) ([]model.SymbolInfo, error) {
	endpoint := fmt.Sprintf("%s/search?q=%s&token=%s", restBase, url.QueryEscape(query), p.apiKey)
	resp, err := p.doGET(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var body struct {
		Result []struct {
			Symbol       string `json:"symbol"`
			Description  string `json:"description"`
			DisplaySymbol string `json:"displaySymbol"`
			Type         string `json:"type"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &body); err != nil {
		return nil, err
	}

	out := make([]model.SymbolInfo, 0, len(body.Result))
	for _, item := range body.Result {
		symbolType := strings.TrimSpace(item.Type)
		if symbolType != "" && symbolType != "Common Stock" && symbolType != "ETP" && symbolType != "ETF" {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		out = append(out, model.SymbolInfo{
			Symbol:      symbol,
			Name:        item.DisplaySymbol,
			Description: item.Description,
			Type:        symbolType,
			Exchange:    "US",
			Currency:    "USD",
		})
	}
	return out, nil
}

func (p *Provider) ResolveSymbol(ctx context.Context, symbol string) (*model.SymbolInfo, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	results, err := p.SearchSymbols(ctx, symbol)
	if err != nil {
		return nil, err
	}
	for _, item := range results {
		if item.Symbol == symbol {
			return &item, nil
		}
	}
	if len(results) > 0 {
		return &results[0], nil
	}
	return &model.SymbolInfo{
		Symbol:      symbol,
		Name:        symbol,
		Description: symbol,
		Type:        "stock",
		Exchange:    "US",
		Currency:    "USD",
	}, nil
}

func (p *Provider) GetQuote(ctx context.Context, symbol string) (map[string]float64, error) {
	endpoint := fmt.Sprintf("%s/quote?symbol=%s&token=%s", restBase, url.QueryEscape(symbol), p.apiKey)
	resp, err := p.doGET(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var body struct {
		C  float64 `json:"c"`
		D  float64 `json:"d"`
		DP float64 `json:"dp"`
		H  float64 `json:"h"`
		L  float64 `json:"l"`
		O  float64 `json:"o"`
		PC float64 `json:"pc"`
	}
	if err := json.Unmarshal(resp, &body); err != nil {
		return nil, err
	}

	if body.C == 0 && body.PC == 0 {
		return nil, fmt.Errorf("finnhub: no quote data for %s", symbol)
	}

	return map[string]float64{
		"c":  body.C,
		"d":  body.D,
		"dp": body.DP,
		"h":  body.H,
		"l":  body.L,
		"o":  body.O,
		"pc": body.PC,
	}, nil
}

func (p *Provider) GetCandles(ctx context.Context, symbol, resolution string, from, to int64) ([]model.Candle, error) {
	endpoint := fmt.Sprintf(
		"%s/stock/candle?symbol=%s&resolution=%s&from=%d&to=%d&token=%s",
		restBase,
		url.QueryEscape(symbol),
		url.QueryEscape(resolution),
		from,
		to,
		p.apiKey,
	)
	resp, err := p.doGET(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var body struct {
		S string    `json:"s"`
		T []int64   `json:"t"`
		O []float64 `json:"o"`
		H []float64 `json:"h"`
		L []float64 `json:"l"`
		C []float64 `json:"c"`
		V []float64 `json:"v"`
	}
	if err := json.Unmarshal(resp, &body); err != nil {
		return nil, err
	}
	if body.S != "ok" || len(body.T) == 0 {
		return nil, nil
	}

	candles := make([]model.Candle, len(body.T))
	for i := range body.T {
		vol := 0.0
		if i < len(body.V) {
			vol = body.V[i]
		}
		candles[i] = model.Candle{
			Time:   body.T[i],
			Open:   body.O[i],
			High:   body.H[i],
			Low:    body.L[i],
			Close:  body.C[i],
			Volume: vol,
		}
	}
	return candles, nil
}

func (p *Provider) doGET(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	res, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("finnhub status %d: %s", res.StatusCode, string(body))
	}
	return body, nil
}
