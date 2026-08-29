package symbols

import (
	"strings"

	"market-service/internal/model"
)

// PopularUSStocks is a curated list of major US equities for browse/search fallback.
var PopularUSStocks = []model.SymbolInfo{
	{Symbol: "AAPL", Name: "AAPL", Description: "Apple Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "MSFT", Name: "MSFT", Description: "Microsoft Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "GOOGL", Name: "GOOGL", Description: "Alphabet Inc Class A", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "GOOG", Name: "GOOG", Description: "Alphabet Inc Class C", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "AMZN", Name: "AMZN", Description: "Amazon.com Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "NVDA", Name: "NVDA", Description: "NVIDIA Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "TSLA", Name: "TSLA", Description: "Tesla Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "META", Name: "META", Description: "Meta Platforms Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "BRK.B", Name: "BRK.B", Description: "Berkshire Hathaway Class B", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "JPM", Name: "JPM", Description: "JPMorgan Chase & Co", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "V", Name: "V", Description: "Visa Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "UNH", Name: "UNH", Description: "UnitedHealth Group", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "XOM", Name: "XOM", Description: "Exxon Mobil Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "JNJ", Name: "JNJ", Description: "Johnson & Johnson", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "WMT", Name: "WMT", Description: "Walmart Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "MA", Name: "MA", Description: "Mastercard Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "PG", Name: "PG", Description: "Procter & Gamble Co", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "HD", Name: "HD", Description: "Home Depot Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "CVX", Name: "CVX", Description: "Chevron Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "MRK", Name: "MRK", Description: "Merck & Co Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "ABBV", Name: "ABBV", Description: "AbbVie Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "KO", Name: "KO", Description: "Coca-Cola Co", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "PEP", Name: "PEP", Description: "PepsiCo Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "COST", Name: "COST", Description: "Costco Wholesale Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "AVGO", Name: "AVGO", Description: "Broadcom Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "TMO", Name: "TMO", Description: "Thermo Fisher Scientific", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "MCD", Name: "MCD", Description: "McDonald's Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "CSCO", Name: "CSCO", Description: "Cisco Systems Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "ACN", Name: "ACN", Description: "Accenture PLC", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "ADBE", Name: "ADBE", Description: "Adobe Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "CRM", Name: "CRM", Description: "Salesforce Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "NFLX", Name: "NFLX", Description: "Netflix Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "AMD", Name: "AMD", Description: "Advanced Micro Devices", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "INTC", Name: "INTC", Description: "Intel Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "DIS", Name: "DIS", Description: "Walt Disney Co", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "PYPL", Name: "PYPL", Description: "PayPal Holdings Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "BAC", Name: "BAC", Description: "Bank of America Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "WFC", Name: "WFC", Description: "Wells Fargo & Co", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "GS", Name: "GS", Description: "Goldman Sachs Group", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "MS", Name: "MS", Description: "Morgan Stanley", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "VZ", Name: "VZ", Description: "Verizon Communications", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "T", Name: "T", Description: "AT&T Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "CMCSA", Name: "CMCSA", Description: "Comcast Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "NKE", Name: "NKE", Description: "Nike Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "ORCL", Name: "ORCL", Description: "Oracle Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "QCOM", Name: "QCOM", Description: "Qualcomm Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "TXN", Name: "TXN", Description: "Texas Instruments Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "AMAT", Name: "AMAT", Description: "Applied Materials Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "LIN", Name: "LIN", Description: "Linde PLC", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "HON", Name: "HON", Description: "Honeywell International", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "UPS", Name: "UPS", Description: "United Parcel Service", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "LOW", Name: "LOW", Description: "Lowe's Companies Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "SBUX", Name: "SBUX", Description: "Starbucks Corp", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "INTU", Name: "INTU", Description: "Intuit Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "AMGN", Name: "AMGN", Description: "Amgen Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "IBM", Name: "IBM", Description: "International Business Machines", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "CAT", Name: "CAT", Description: "Caterpillar Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "GE", Name: "GE", Description: "GE Aerospace", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "BA", Name: "BA", Description: "Boeing Co", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "F", Name: "F", Description: "Ford Motor Co", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "GM", Name: "GM", Description: "General Motors Co", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "UBER", Name: "UBER", Description: "Uber Technologies Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "ABNB", Name: "ABNB", Description: "Airbnb Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "COIN", Name: "COIN", Description: "Coinbase Global Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "SQ", Name: "SQ", Description: "Block Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "SHOP", Name: "SHOP", Description: "Shopify Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "SNOW", Name: "SNOW", Description: "Snowflake Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "PLTR", Name: "PLTR", Description: "Palantir Technologies", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "SOFI", Name: "SOFI", Description: "SoFi Technologies Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "RIVN", Name: "RIVN", Description: "Rivian Automotive Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
	{Symbol: "LCID", Name: "LCID", Description: "Lucid Group Inc", Exchange: "US", Type: "Common Stock", Currency: "USD"},
}

// FilterPopular returns symbols matching query (symbol or description), or all when query is empty.
func FilterPopular(query string, limit int) []model.SymbolInfo {
	if limit <= 0 {
		limit = 30
	}
	if limit > len(PopularUSStocks) {
		limit = len(PopularUSStocks)
	}

	q := strings.ToUpper(strings.TrimSpace(query))
	if q == "" {
		out := make([]model.SymbolInfo, limit)
		copy(out, PopularUSStocks[:limit])
		return out
	}

	out := make([]model.SymbolInfo, 0, limit)
	for _, item := range PopularUSStocks {
		if strings.Contains(item.Symbol, q) ||
			strings.Contains(strings.ToUpper(item.Description), q) ||
			strings.Contains(strings.ToUpper(item.Name), q) {
			out = append(out, item)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}
