"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { isAuthenticated } from "@/lib/auth";
import { Activity, Radio, Wifi, WifiOff } from "lucide-react";
import { SymbolSearch } from "@/components/charts/symbol-search";
import { TradingViewChart } from "@/components/charts/tradingview-chart";
import { Navbar } from "@/components/layout/navbar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useStockWebSocket } from "@/hooks/use-stock-websocket";
import { fetchQuote } from "@/lib/market/api";
import { resolutionToTradingViewInterval } from "@/lib/market/tradingview";
import { POPULAR_SYMBOLS, type CandleResolution } from "@/lib/market/types";
import { cn, formatChange, formatPrice } from "@/lib/utils";

export default function ChartsPage() {
  const router = useRouter();
  const [symbol, setSymbol] = useState("AAPL");
  const [input, setInput] = useState("AAPL");
  const [resolution, setResolution] = useState<CandleResolution>("D");
  const [quote, setQuote] = useState<{
    c: number;
    d: number;
    dp: number;
    h: number;
    l: number;
    o: number;
    pc: number;
  } | null>(null);
  const [quoteLoading, setQuoteLoading] = useState(true);
  const [quoteError, setQuoteError] = useState<string | null>(null);

  const { lastTrade, status, reconnect } = useStockWebSocket(symbol);
  const tradingViewInterval = useMemo(
    () => resolutionToTradingViewInterval(resolution),
    [resolution],
  );

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace("/login");
    }
  }, [router]);

  const loadQuote = useCallback(async () => {
    setQuoteLoading(true);
    setQuoteError(null);
    try {
      const q = await fetchQuote(symbol);
      setQuote(q);
    } catch (e) {
      setQuote(null);
      setQuoteError(e instanceof Error ? e.message : "Failed to load quote");
    } finally {
      setQuoteLoading(false);
    }
  }, [symbol]);

  useEffect(() => {
    loadQuote();
  }, [loadQuote]);

  const livePrice = lastTrade?.price ?? quote?.c ?? null;
  const change = quote?.dp ?? 0;
  const isPositive = change >= 0;

  const handleSymbolSelect = (next: string) => {
    const symbolValue = next.trim().toUpperCase();
    if (symbolValue) {
      setSymbol(symbolValue);
      setInput(symbolValue);
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100">
      <div className="pointer-events-none fixed inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-20%,rgba(16,185,129,0.12),transparent)]" />
      <Navbar />

      <main className="relative mx-auto max-w-7xl px-6 pb-12 pt-24">
        <div className="mb-6 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <div className="mb-2 flex items-center gap-2">
              <Badge variant="secondary" className="gap-1.5">
                <Radio className="h-3 w-3 text-emerald-400" />
                Live feed
              </Badge>
              <Badge
                variant={status === "connected" ? "positive" : "outline"}
                className="gap-1.5"
              >
                {status === "connected" ? (
                  <Wifi className="h-3 w-3" />
                ) : (
                  <WifiOff className="h-3 w-3" />
                )}
                {status}
              </Badge>
            </div>
            <h1 className="font-display text-4xl font-semibold tracking-tight text-white">
              {symbol}
            </h1>
            <div className="mt-2 flex items-baseline gap-3">
              <span className="font-mono text-3xl font-medium tabular-nums text-white">
                {livePrice != null ? formatPrice(livePrice) : "—"}
              </span>
              {quote && (
                <span
                  className={cn(
                    "font-mono text-sm tabular-nums",
                    isPositive ? "text-emerald-400" : "text-rose-400",
                  )}
                >
                  {formatChange(change)}
                </span>
              )}
              {quoteError && !quoteLoading && (
                <span className="text-xs text-amber-400/90">
                  Quote unavailable — chart still loads
                </span>
              )}
            </div>
          </div>

          <div className="w-full max-w-md">
            <SymbolSearch
              value={input}
              onChange={setInput}
              onSelect={handleSymbolSelect}
            />
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-[240px_1fr]">
          <aside className="space-y-4">
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium text-zinc-300">
                  Watchlist
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-1 p-3 pt-0">
                {POPULAR_SYMBOLS.map((item) => (
                  <button
                    key={item.symbol}
                    type="button"
                    onClick={() => {
                      setSymbol(item.symbol);
                      setInput(item.symbol);
                    }}
                    className={cn(
                      "flex w-full items-center justify-between rounded-lg px-3 py-2.5 text-left text-sm transition",
                      symbol === item.symbol
                        ? "bg-emerald-500/10 text-emerald-300 ring-1 ring-emerald-500/20"
                        : "text-zinc-400 hover:bg-white/5 hover:text-zinc-100",
                    )}
                  >
                    <span className="font-medium">{item.symbol}</span>
                    <span className="truncate pl-2 text-xs text-zinc-500">
                      {item.name}
                    </span>
                  </button>
                ))}
              </CardContent>
            </Card>

            {quote && (
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium text-zinc-300">
                    Session
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 p-4 pt-0 text-sm">
                  {[
                    ["Open", formatPrice(quote.o)],
                    ["High", formatPrice(quote.h)],
                    ["Low", formatPrice(quote.l)],
                    ["Prev close", formatPrice(quote.pc)],
                  ].map(([label, value]) => (
                    <div key={label} className="flex justify-between">
                      <span className="text-zinc-500">{label}</span>
                      <span className="font-mono tabular-nums text-zinc-200">
                        {value}
                      </span>
                    </div>
                  ))}
                </CardContent>
              </Card>
            )}
          </aside>

          <section className="space-y-4">
            <Card className="overflow-hidden">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 border-b border-white/[0.06] py-4">
                <div className="flex items-center gap-2">
                  <Activity className="h-4 w-4 text-emerald-400" />
                  <CardTitle className="text-base">Price chart</CardTitle>
                </div>
                <Tabs
                  value={resolution}
                  onValueChange={(v) => setResolution(v as CandleResolution)}
                >
                  <TabsList>
                    {(["1", "5", "15", "60", "D"] as const).map((r) => (
                      <TabsTrigger key={r} value={r}>
                        {r === "D" ? "1D" : `${r}m`}
                      </TabsTrigger>
                    ))}
                  </TabsList>
                </Tabs>
              </CardHeader>
              <CardContent className="p-0">
                <div className="relative h-[min(62vh,560px)] min-h-[420px]">
                  {quoteLoading && (
                    <div className="absolute left-4 top-4 z-10 flex items-center gap-2 rounded-md bg-zinc-950/80 px-2 py-1 text-xs text-zinc-400">
                      <div className="h-3 w-3 animate-spin rounded-full border-2 border-emerald-500/30 border-t-emerald-400" />
                      Loading quote…
                    </div>
                  )}
                  <TradingViewChart
                    symbol={symbol}
                    interval={tradingViewInterval}
                    className="h-full"
                  />
                </div>
              </CardContent>
            </Card>

            <div className="flex flex-wrap items-center justify-between gap-3 text-xs text-zinc-500">
              <span>
                {/* Powered by TradingView Advanced Charts · Finnhub WebSocket quotes */}
              </span>
              {status !== "connected" && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 text-xs"
                  onClick={reconnect}
                >
                  Reconnect feed
                </Button>
              )}
            </div>
          </section>
        </div>

        <Separator className="my-8" />
        <p className="text-center text-xs text-zinc-600">
          Market data delayed. For demonstration purposes only.
        </p>
      </main>
    </div>
  );
}
