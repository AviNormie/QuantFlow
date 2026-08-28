package candles

import "testing"

func TestNormalizeRangeClampsFutureTo(t *testing.T) {
	from, to := NormalizeRange(1748131200, 1788134400, "D")
	if to > 2000000000 {
		t.Fatalf("expected to clamped to near now, got %d", to)
	}
	if from >= to {
		t.Fatalf("from must be before to: from=%d to=%d", from, to)
	}
}

func TestNormalizeRangeMaxDailyWindow(t *testing.T) {
	from, to := NormalizeRange(0, 2000000000, "D")
	if to-from > 366*86400 {
		t.Fatalf("daily window too wide: %d seconds", to-from)
	}
}
