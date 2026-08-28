import { NextResponse } from "next/server";

export async function GET() {
  return NextResponse.json(
    {
      error:
        "Deprecated route. Use API gateway /api/market/candles/{symbol} via market-service.",
    },
    { status: 501 },
  );
}
