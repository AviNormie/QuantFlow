import { getPublicApiUrl } from "@/lib/env";
import { getAccessToken, refreshTokens } from "@/lib/auth";
import type { CandleBar, CandleResolution } from "./types";

export type SymbolInfo = {
  symbol: string;
  name: string;
  description: string;
  exchange: string;
  type: string;
};

type FinnhubCandleResponse = {
  s: "ok" | "no_data";
  t?: number[];
  o?: number[];
  h?: number[];
  l?: number[];
  c?: number[];
  v?: number[];
};

async function authFetch(url: string, options: RequestInit = {}) {
  const headers = new Headers(options.headers);
  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  let res = await fetch(url, { ...options, headers });

  if (res.status === 401) {
    const refreshed = await refreshTokens();
    if (refreshed) {
      headers.set("Authorization", `Bearer ${getAccessToken()}`);
      res = await fetch(url, { ...options, headers });
    }
  }

  return res;
}

export async function fetchCandles(
  symbol: string,
  resolution: CandleResolution = "D",
  from?: number,
  to?: number,
): Promise<CandleBar[]> {
  const params = new URLSearchParams({ resolution });
  if (from) params.set("from", String(from));
  if (to) params.set("to", String(to));

  const res = await authFetch(
    `${getPublicApiUrl()}/api/market/candles/${encodeURIComponent(symbol)}?${params}`,
  );

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
  const res = await authFetch(
    `${getPublicApiUrl()}/api/market/quotes/${encodeURIComponent(symbol)}`,
  );

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

export async function searchSymbols(query: string, limit = 30): Promise<SymbolInfo[]> {
  const params = new URLSearchParams({ limit: String(limit) });
  const trimmed = query.trim();
  if (trimmed) {
    params.set("q", trimmed);
  }
  const res = await authFetch(
    `${getPublicApiUrl()}/api/market/symbols/search?${params}`,
  );
  if (!res.ok) {
    throw new Error("Failed to search symbols");
  }
  const data = (await res.json()) as { results: SymbolInfo[] };
  return data.results ?? [];
}

export async function resolveSymbol(symbol: string): Promise<SymbolInfo> {
  const res = await authFetch(
    `${getPublicApiUrl()}/api/market/symbols/${encodeURIComponent(symbol)}`,
  );
  if (!res.ok) {
    throw new Error("Failed to resolve symbol");
  }
  return res.json() as Promise<SymbolInfo>;
}
