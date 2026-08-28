"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { getPublicWebSocketUrl } from "@/lib/env";
import { getAccessToken } from "@/lib/auth";
import type { LiveTrade } from "@/lib/market/types";

type ConnectionStatus = "connecting" | "connected" | "disconnected" | "error";

function buildWsUrl(symbol: string) {
  const base = getPublicWebSocketUrl();
  const url = new URL("/ws", base.endsWith("/") ? base : `${base}/`);
  const token = getAccessToken();
  if (token) {
    url.searchParams.set("token", token);
  }
  url.searchParams.set("symbol", symbol);
  return url.toString();
}

export function useStockWebSocket(symbol: string) {
  const [lastTrade, setLastTrade] = useState<LiveTrade | null>(null);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const wsRef = useRef<WebSocket | null>(null);
  const symbolRef = useRef(symbol);

  const connect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    symbolRef.current = symbol;
    setStatus("connecting");
    setLastTrade(null);

    const ws = new WebSocket(buildWsUrl(symbol));
    wsRef.current = ws;

    ws.onopen = () => {
      setStatus("connected");
      ws.send(
        JSON.stringify({
          action: "subscribe",
          symbols: [symbolRef.current],
        }),
      );
    };

    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data as string) as {
          type?: string;
          data?: Array<{
            s: string;
            p: number;
            v: number;
            t: number;
          }>;
        };

        if (payload.type === "trade" && payload.data?.length) {
          const trade = payload.data[payload.data.length - 1];
          if (trade.s === symbolRef.current) {
            setLastTrade({
              symbol: trade.s,
              price: trade.p,
              volume: trade.v,
              timestamp: trade.t,
            });
          }
        }
      } catch {
        // ignore malformed messages
      }
    };

    ws.onerror = () => setStatus("error");
    ws.onclose = () => setStatus("disconnected");
  }, [symbol]);

  useEffect(() => {
    connect();
    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [connect]);

  return { lastTrade, status, reconnect: connect };
}
