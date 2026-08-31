"use client";

import { useCallback, useEffect, useState } from "react";
import { stockFlowWs, type WsConnectionStatus } from "@/lib/market/ws-client";
import type { LiveTrade } from "@/lib/market/types";

export function useStockWebSocket(symbol: string) {
  const [lastTrade, setLastTrade] = useState<LiveTrade | null>(null);
  const [status, setStatus] = useState<WsConnectionStatus>("connecting");

  const reconnect = useCallback(() => {
    stockFlowWs.forceReconnect();
  }, []);

  useEffect(() => {
    const normalized = symbol.toUpperCase();
    stockFlowWs.connect();
    stockFlowWs.subscribe([normalized]);

    const offStatus = stockFlowWs.addStatusListener(setStatus);
    const offTrade = stockFlowWs.addTradeListener((trade) => {
      if (trade.s === normalized) {
        setLastTrade({
          symbol: trade.s,
          price: trade.p,
          volume: trade.v,
          timestamp: trade.t,
        });
      }
    });

    return () => {
      offTrade();
      offStatus();
      stockFlowWs.unsubscribe([normalized]);
      stockFlowWs.release();
    };
  }, [symbol]);

  return { lastTrade, status, reconnect };
}
