"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type { LiveTrade } from "@/lib/market/types";

type ConnectionStatus = "connecting" | "connected" | "disconnected" | "error";

export function useFinnhubWebSocket(symbol: string) {
  const [lastTrade, setLastTrade] = useState<LiveTrade | null>(null);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const wsRef = useRef<WebSocket | null>(null);
  const symbolRef = useRef(symbol);
  const prevSymbolRef = useRef(symbol);

  const connect = useCallback(() => {
    const token = process.env.NEXT_PUBLIC_FINNHUB_API_KEY;
    if (!token) {
      setStatus("error");
      return;
    }

    if (wsRef.current) {
      wsRef.current.close();
    }

    setStatus("connecting");
    const ws = new WebSocket(`wss://ws.finnhub.io?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => {
      setStatus("connected");
      ws.send(JSON.stringify({ type: "subscribe", symbol: symbolRef.current }));
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
  }, []);

  useEffect(() => {
    const prev = prevSymbolRef.current;
    prevSymbolRef.current = symbol;
    symbolRef.current = symbol;

    if (wsRef.current?.readyState === WebSocket.OPEN && prev !== symbol) {
      wsRef.current.send(JSON.stringify({ type: "unsubscribe", symbol: prev }));
      wsRef.current.send(JSON.stringify({ type: "subscribe", symbol }));
    }

    setLastTrade(null);
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
