package candles

import (
	"testing"

	"market-service/internal/model"
)

func TestNormalizeBarTimesDailyUTC(t *testing.T) {
	ts := int64(1750019400) // 2025-06-15 18:30:00 UTC
	bars := NormalizeBarTimes([]model.Candle{
		{Time: ts, Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100},
	}, "D")
	if len(bars) != 1 {
		t.Fatalf("expected 1 bar")
	}
	want := utcDayStart(ts)
	if bars[0].Time != want {
		t.Fatalf("got %d want %d", bars[0].Time, want)
	}
}
