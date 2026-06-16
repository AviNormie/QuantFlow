import { NextResponse } from "next/server";
import type { CandleResolution } from "@/lib/market/types";
import { fetchYahooCandles } from "@/lib/market/yahoo-candles";

const FINNHUB_BASE = "https://finnhub.io/api/v1";

function getToken() {
  return process.env.FINNHUB_API_KEY ?? process.env.NEXT_PUBLIC_FINNHUB_API_KEY;
}

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const symbol = searchParams.get("symbol")?.toUpperCase();
  const resolution = (searchParams.get("resolution") ?? "D") as CandleResolution;

  if (!symbol) {
    return NextResponse.json({ error: "symbol is required" }, { status: 400 });
  }

  const token = getToken();
  if (!token) {
    return NextResponse.json(
      { error: "FINNHUB_API_KEY is not configured" },
      { status: 500 },
    );
  }

  const now = Math.floor(Date.now() / 1000);
  const from =
    resolution === "D" || resolution === "W"
      ? now - 365 * 24 * 60 * 60
      : now - 7 * 24 * 60 * 60;

  const url = new URL(`${FINNHUB_BASE}/stock/candle`);
  url.searchParams.set("symbol", symbol);
  url.searchParams.set("resolution", resolution);
  url.searchParams.set("from", String(from));
  url.searchParams.set("to", String(now));
  url.searchParams.set("token", token);

  const res = await fetch(url, { next: { revalidate: 60 } });
  const data = await res.json();

  // Finnhub free tier does not include /stock/candle — fall back to Yahoo Finance.
  if (data.s === "ok" && Array.isArray(data.t) && data.t.length > 0) {
    return NextResponse.json(data);
  }

  const yahoo = await fetchYahooCandles(symbol, resolution);
  return NextResponse.json(yahoo);
}
