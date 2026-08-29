package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"market-service/internal/model"
)

const chartBase = "https://query1.finance.yahoo.com/v8/finance/chart"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (c *Client) GetCandles(ctx context.Context, symbol, resolution string, from, to int64) ([]model.Candle, error) {
	interval, period1, period2 := yahooParams(resolution, from, to)

	endpoint := fmt.Sprintf("%s/%s?interval=%s&period1=%d&period2=%d",
		chartBase,
		url.PathEscape(ToYahooSymbol(symbol)),
		url.QueryEscape(interval),
		period1,
		period2,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; StockFlow/1.0)")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("yahoo status %d", res.StatusCode)
	}

	var payload struct {
		Chart struct {
			Result []struct {
				Timestamp  []int64 `json:"timestamp"`
				Indicators struct {
					Quote []struct {
						Open   []*float64 `json:"open"`
						High   []*float64 `json:"high"`
						Low    []*float64 `json:"low"`
						Close  []*float64 `json:"close"`
						Volume []*float64 `json:"volume"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if len(payload.Chart.Result) == 0 {
		return nil, nil
	}

	result := payload.Chart.Result[0]
	if len(result.Timestamp) == 0 || len(result.Indicators.Quote) == 0 {
		return nil, nil
	}

	quote := result.Indicators.Quote[0]
	candles := make([]model.Candle, 0, len(result.Timestamp))

	for i, ts := range result.Timestamp {
		open := deref(quote.Open, i)
		high := deref(quote.High, i)
		low := deref(quote.Low, i)
		closeP := deref(quote.Close, i)
		if open == nil || high == nil || low == nil || closeP == nil {
			continue
		}
		vol := 0.0
		if v := deref(quote.Volume, i); v != nil {
			vol = *v
		}
		candles = append(candles, model.Candle{
			Time:   ts,
			Open:   *open,
			High:   *high,
			Low:    *low,
			Close:  *closeP,
			Volume: vol,
		})
	}

	return candles, nil
}

// GetQuote returns latest price stats from Yahoo chart meta (Finnhub-compatible shape).
func (c *Client) GetQuote(ctx context.Context, symbol string) (map[string]float64, error) {
	yahooSymbol := ToYahooSymbol(symbol)
	endpoint := fmt.Sprintf("%s/%s?interval=1d&range=5d",
		chartBase,
		url.PathEscape(yahooSymbol),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; StockFlow/1.0)")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("yahoo status %d", res.StatusCode)
	}

	var payload struct {
		Chart struct {
			Result []struct {
				Meta struct {
					RegularMarketPrice   float64 `json:"regularMarketPrice"`
					PreviousClose        float64 `json:"previousClose"`
					ChartPreviousClose   float64 `json:"chartPreviousClose"`
					RegularMarketOpen    float64 `json:"regularMarketOpen"`
					RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
					RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
				} `json:"meta"`
			} `json:"result"`
		} `json:"chart"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if len(payload.Chart.Result) == 0 {
		return nil, nil
	}

	meta := payload.Chart.Result[0].Meta
	price := meta.RegularMarketPrice
	prevClose := meta.PreviousClose
	if prevClose == 0 {
		prevClose = meta.ChartPreviousClose
	}
	if price <= 0 {
		return nil, nil
	}

	change := price - prevClose
	changePct := 0.0
	if prevClose > 0 {
		changePct = (change / prevClose) * 100
	}

	return map[string]float64{
		"c":  price,
		"d":  change,
		"dp": changePct,
		"h":  meta.RegularMarketDayHigh,
		"l":  meta.RegularMarketDayLow,
		"o":  meta.RegularMarketOpen,
		"pc": prevClose,
	}, nil
}

// ToYahooSymbol maps Finnhub-style tickers to Yahoo (e.g. BRK.B → BRK-B).
func ToYahooSymbol(symbol string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(symbol)), ".", "-")
}

func deref(values []*float64, i int) *float64 {
	if i >= len(values) {
		return nil
	}
	return values[i]
}

func yahooParams(resolution string, from, to int64) (interval string, period1, period2 int64) {
	period1, period2 = from, to
	switch resolution {
	case "1":
		interval = "1m"
	case "5":
		interval = "5m"
	case "15":
		interval = "15m"
	case "30":
		interval = "30m"
	case "60":
		interval = "60m"
	case "W":
		interval = "1wk"
	default:
		interval = "1d"
	}
	return interval, period1, period2
}
