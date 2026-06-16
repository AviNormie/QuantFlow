export type CandleResolution = "1" | "5" | "15" | "30" | "60" | "D" | "W";

export type CandleBar = {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume?: number;
};

export type LiveTrade = {
  symbol: string;
  price: number;
  volume: number;
  timestamp: number;
};

export const POPULAR_SYMBOLS = [
  { symbol: "AAPL", name: "Apple Inc." },
  { symbol: "MSFT", name: "Microsoft" },
  { symbol: "GOOGL", name: "Alphabet" },
  { symbol: "AMZN", name: "Amazon" },
  { symbol: "NVDA", name: "NVIDIA" },
  { symbol: "TSLA", name: "Tesla" },
  { symbol: "META", name: "Meta" },
  { symbol: "JPM", name: "JPMorgan" },
] as const;

export function resolutionToSeconds(resolution: CandleResolution): number {
  switch (resolution) {
    case "1":
      return 60;
    case "5":
      return 300;
    case "15":
      return 900;
    case "30":
      return 1800;
    case "60":
      return 3600;
    case "D":
      return 86400;
    case "W":
      return 604800;
    default:
      return 86400;
  }
}
