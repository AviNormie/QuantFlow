package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"market-service/internal/model"
	"market-service/internal/service"
	candlesutil "market-service/internal/service/candles"

	"github.com/gin-gonic/gin"
)

type MarketHandler struct {
	marketService *service.MarketService
}

func NewMarketHandler(marketService *service.MarketService) *MarketHandler {
	return &MarketHandler{marketService: marketService}
}

func (h *MarketHandler) SearchSymbols(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	results, err := h.marketService.SearchSymbols(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

func (h *MarketHandler) ResolveSymbol(c *gin.Context) {
	symbol := strings.TrimSpace(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	info, err := h.marketService.ResolveSymbol(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "resolve failed"})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *MarketHandler) GetQuote(c *gin.Context) {
	symbol := strings.TrimSpace(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	quote, err := h.marketService.GetQuote(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "quote failed"})
		return
	}

	c.JSON(http.StatusOK, quote)
}

func (h *MarketHandler) GetCandles(c *gin.Context) {
	symbol := strings.TrimSpace(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	resolution := c.DefaultQuery("resolution", "D")
	from, to, err := parseRange(c, resolution)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bars, err := h.marketService.GetCandles(c.Request.Context(), symbol, resolution, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "candles failed"})
		return
	}

	if len(bars) == 0 {
		c.JSON(http.StatusOK, gin.H{"s": "no_data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"s": "ok",
		"t": pluckTimes(bars),
		"o": pluckOpen(bars),
		"h": pluckHigh(bars),
		"l": pluckLow(bars),
		"c": pluckClose(bars),
		"v": pluckVolume(bars),
	})
}

func parseRange(c *gin.Context, resolution string) (int64, int64, error) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr != "" && toStr != "" {
		from, err := strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		to, err := strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		from, to = candlesutil.NormalizeRange(from, to, resolution)
		return from, to, nil
	}

	now := time.Now().Unix()
	from, to := candlesutil.NormalizeRange(now-30*24*3600, now, resolution)
	return from, to, nil
}

func pluckTimes(candles []model.Candle) []int64 {
	out := make([]int64, len(candles))
	for i, c := range candles {
		out[i] = c.Time
	}
	return out
}

func pluckOpen(candles []model.Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Open
	}
	return out
}

func pluckHigh(candles []model.Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.High
	}
	return out
}

func pluckLow(candles []model.Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Low
	}
	return out
}

func pluckClose(candles []model.Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Close
	}
	return out
}

func pluckVolume(candles []model.Candle) []float64 {
	out := make([]float64, len(candles))
	for i, c := range candles {
		out[i] = c.Volume
	}
	return out
}
