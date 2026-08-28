package candles

import "time"

// NormalizeRange clamps TradingView/Finnhub candle requests to valid windows.
func NormalizeRange(from, to int64, resolution string) (int64, int64) {
	now := time.Now().Unix()
	if to <= 0 || to > now {
		to = now
	}
	if from <= 0 {
		from = to - defaultWindow(resolution)
	}
	if from >= to {
		from = to - defaultWindow(resolution)
	}

	maxWindow := maxWindow(resolution)
	if to-from > maxWindow {
		from = to - maxWindow
	}

	return from, to
}

func defaultWindow(resolution string) int64 {
	switch resolution {
	case "1":
		return 5 * 24 * 3600
	case "5":
		return 10 * 24 * 3600
	case "15", "30":
		return 30 * 24 * 3600
	case "60":
		return 90 * 24 * 3600
	case "W":
		return 5 * 365 * 24 * 3600
	default:
		return 365 * 24 * 3600
	}
}

func maxWindow(resolution string) int64 {
	switch resolution {
	case "1":
		return 7 * 24 * 3600
	case "5":
		return 30 * 24 * 3600
	case "15", "30":
		return 60 * 24 * 3600
	case "60":
		return 180 * 24 * 3600
	case "W":
		return 10 * 365 * 24 * 3600
	default:
		return 365 * 24 * 3600
	}
}
