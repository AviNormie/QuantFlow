import type { CandleResolution } from "./types";

/** Map StockFlow candle resolution to TradingView interval strings. */
export function resolutionToTradingViewInterval(
  resolution: CandleResolution,
): string {
  switch (resolution) {
    case "1":
      return "1";
    case "5":
      return "5";
    case "15":
      return "15";
    case "30":
      return "30";
    case "60":
      return "60";
    case "D":
      return "1D";
    case "W":
      return "1W";
    default:
      return "1D";
  }
}

export const TRADINGVIEW_DEMO_DATAFEED_URL =
  "https://demo-feed-data.tradingview.com";
