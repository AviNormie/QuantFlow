import { NextResponse } from "next/server";

const FINNHUB_BASE = "https://finnhub.io/api/v1";

function getToken() {
  return process.env.FINNHUB_API_KEY ?? process.env.NEXT_PUBLIC_FINNHUB_API_KEY;
}

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const symbol = searchParams.get("symbol")?.toUpperCase();

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

  const url = new URL(`${FINNHUB_BASE}/quote`);
  url.searchParams.set("symbol", symbol);
  url.searchParams.set("token", token);

  const res = await fetch(url, { next: { revalidate: 15 } });
  const data = await res.json();

  return NextResponse.json(data);
}
