import type { CandleResolution } from "./types";

type YahooChartResponse = {
  chart?: {
    result?: Array<{
      timestamp?: number[];
      indicators?: {
        quote?: Array<{
          open?: (number | null)[];
          high?: (number | null)[];
          low?: (number | null)[];
          close?: (number | null)[];
          volume?: (number | null)[];
        }>;
      };
    }>;
  };
};

export type FinnhubCandlePayload = {
  s: "ok" | "no_data";
  t: number[];
  o: number[];
  h: number[];
  l: number[];
  c: number[];
  v: number[];
};

const EMPTY: FinnhubCandlePayload = {
  s: "no_data",
  t: [],
  o: [],
  h: [],
  l: [],
  c: [],
  v: [],
};

function resolutionToYahooParams(resolution: CandleResolution): {
  interval: string;
  range: string;
} {
  switch (resolution) {
    case "1":
      return { interval: "1m", range: "5d" };
    case "5":
      return { interval: "5m", range: "5d" };
    case "15":
      return { interval: "15m", range: "1mo" };
    case "30":
      return { interval: "30m", range: "1mo" };
    case "60":
      return { interval: "60m", range: "3mo" };
    case "D":
      return { interval: "1d", range: "1y" };
    case "W":
      return { interval: "1wk", range: "5y" };
    default:
      return { interval: "1d", range: "1y" };
  }
}

export async function fetchYahooCandles(
  symbol: string,
  resolution: CandleResolution,
): Promise<FinnhubCandlePayload> {
  const { interval, range } = resolutionToYahooParams(resolution);
  const url = new URL(
    `https://query1.finance.yahoo.com/v8/finance/chart/${encodeURIComponent(symbol)}`,
  );
  url.searchParams.set("interval", interval);
  url.searchParams.set("range", range);

  const res = await fetch(url.toString(), {
    headers: { "User-Agent": "Mozilla/5.0 (compatible; StockFlow/1.0)" },
    next: { revalidate: 60 },
  });

  if (!res.ok) return EMPTY;

  const data = (await res.json()) as YahooChartResponse;
  const result = data.chart?.result?.[0];
  const timestamps = result?.timestamp ?? [];
  const quote = result?.indicators?.quote?.[0];

  if (!quote || !timestamps.length) return EMPTY;

  const t: number[] = [];
  const o: number[] = [];
  const h: number[] = [];
  const l: number[] = [];
  const c: number[] = [];
  const v: number[] = [];

  for (let i = 0; i < timestamps.length; i++) {
    const open = quote.open?.[i];
    const high = quote.high?.[i];
    const low = quote.low?.[i];
    const close = quote.close?.[i];
    if (open == null || high == null || low == null || close == null) continue;

    t.push(timestamps[i]);
    o.push(open);
    h.push(high);
    l.push(low);
    c.push(close);
    v.push(quote.volume?.[i] ?? 0);
  }

  if (!t.length) return EMPTY;

  return { s: "ok", t, o, h, l, c, v };
}
