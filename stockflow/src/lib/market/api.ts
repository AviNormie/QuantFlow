import { getApiUrl } from "@/lib/auth";
import type { CandleBar, CandleResolution } from "./types";

type FinnhubCandleResponse = {
  s: "ok" | "no_data";
  t?: number[];
  o?: number[];
  h?: number[];
  l?: number[];
  c?: number[];
  v?: number[];
};

export async function fetchCandles(
  symbol: string,
  resolution: CandleResolution = "D",
): Promise<CandleBar[]> {
  const params = new URLSearchParams({ symbol, resolution });
  const res = await fetch(`${getApiUrl()}/api/market/candles?${params.toString()}`);

  if (!res.ok) {
    const body = await res.text();
    throw new Error(body || "Failed to load candle data");
  }

  const data = (await res.json()) as FinnhubCandleResponse;

  if (data.s !== "ok" || !data.t?.length) {
    return [];
  }

  return data.t.map((time, i) => ({
    time,
    open: data.o![i],
    high: data.h![i],
    low: data.l![i],
    close: data.c![i],
    volume: data.v?.[i],
  }));
}

export async function fetchQuote(symbol: string) {
  const params = new URLSearchParams({ symbol });
  const res = await fetch(`${getApiUrl()}/api/market/quote?${params.toString()}`);

  if (!res.ok) {
    throw new Error("Failed to load quote");
  }

  return res.json() as Promise<{
    c: number;
    d: number;
    dp: number;
    h: number;
    l: number;
    o: number;
    pc: number;
  }>;
}
