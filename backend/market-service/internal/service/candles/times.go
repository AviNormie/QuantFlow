package candles

import (
	"time"

	"market-service/internal/model"
)

// NormalizeBarTimes aligns daily/weekly bar timestamps to UTC midnight for TradingView.
func NormalizeBarTimes(candles []model.Candle, resolution string) []model.Candle {
	if len(candles) == 0 {
		return candles
	}

	switch resolution {
	case "D", "W":
		out := make([]model.Candle, len(candles))
		for i, bar := range candles {
			bar.Time = utcDayStart(bar.Time)
			if resolution == "W" {
				bar.Time = utcWeekStart(bar.Time)
			}
			out[i] = bar
		}
		return out
	default:
		return candles
	}
}

func utcDayStart(unixSec int64) int64 {
	t := time.Unix(unixSec, 0).UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix()
}

func utcWeekStart(unixSec int64) int64 {
	t := time.Unix(unixSec, 0).UTC()
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC).Unix()
}
