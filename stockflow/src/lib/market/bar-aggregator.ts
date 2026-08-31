import type { CandleResolution } from "./types";

type Bar = {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume?: number;
};

export class BarAggregator {
  private bar: Bar | null = null;
  private resolutionSeconds: number;

  constructor(resolutionSeconds: number) {
    this.resolutionSeconds = resolutionSeconds;
  }

  setResolution(resolutionSeconds: number) {
    this.resolutionSeconds = resolutionSeconds;
    this.bar = null;
  }

  onTick(price: number, volume: number, timestampMs: number): Bar | null {
    const tsMs = timestampMs < 1e12 ? timestampMs * 1000 : timestampMs;
    const bucketSec =
      Math.floor(tsMs / 1000 / this.resolutionSeconds) *
      this.resolutionSeconds;

    if (!this.bar || this.bar.time !== bucketSec) {
      this.bar = {
        time: bucketSec,
        open: price,
        high: price,
        low: price,
        close: price,
        volume: volume,
      };
      return { ...this.bar };
    }

    this.bar.high = Math.max(this.bar.high, price);
    this.bar.low = Math.min(this.bar.low, price);
    this.bar.close = price;
    this.bar.volume = (this.bar.volume ?? 0) + volume;
    return { ...this.bar };
  }
}

/** Convert aggregator bucket (unix seconds) to TradingView bar time (ms). */
export function barBucketToTradingViewMs(
  bucketSec: number,
  resolution: string,
): number {
  const ms = bucketSec * 1000;
  if (
    resolution === "1D" ||
    resolution === "D" ||
    resolution === "1W" ||
    resolution === "W"
  ) {
    const d = new Date(ms);
    return Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate());
  }
  return ms;
}

export function tradingViewResolutionToSeconds(resolution: string): number {
  if (resolution === "1D" || resolution === "D") return 86400;
  if (resolution === "1W" || resolution === "W") return 604800;
  const minutes = Number.parseInt(resolution, 10);
  if (!Number.isNaN(minutes)) return minutes * 60;
  return 86400;
}

export function tradingViewResolutionToMarket(resolution: string): CandleResolution {
  if (resolution === "1D" || resolution === "D") return "D";
  if (resolution === "1W" || resolution === "W") return "W";
  if (resolution === "60") return "60";
  if (resolution === "30") return "30";
  if (resolution === "15") return "15";
  if (resolution === "5") return "5";
  if (resolution === "1") return "1";
  return "D";
}
