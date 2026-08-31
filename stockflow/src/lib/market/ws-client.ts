import { getPublicWebSocketUrl } from "@/lib/env";
import { getAccessToken } from "@/lib/auth";

export type WsTrade = {
  s: string;
  p: number;
  v: number;
  t: number;
};

export type WsConnectionStatus =
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected"
  | "error";

type TradeListener = (trade: WsTrade) => void;
type StatusListener = (status: WsConnectionStatus) => void;

const INITIAL_RECONNECT_MS = 1000;
const MAX_RECONNECT_MS = 30000;

function buildWsUrl(): string {
  const base = getPublicWebSocketUrl();
  const url = new URL("/ws", base.endsWith("/") ? base : `${base}/`);
  const token = getAccessToken();
  if (token) {
    url.searchParams.set("token", token);
  }
  return url.toString();
}

function normalizeTimestampMs(t: number): number {
  return t < 1e12 ? t * 1000 : t;
}

/**
 * Single shared WebSocket for the charts page — one connection for header
 * price and TradingView live bars, with auto-reconnect.
 */
class StockFlowWsClient {
  private ws: WebSocket | null = null;
  private subscriptions = new Set<string>();
  private subscriptionCounts = new Map<string, number>();
  private tradeListeners = new Set<TradeListener>();
  private statusListeners = new Set<StatusListener>();
  private status: WsConnectionStatus = "disconnected";
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelay = INITIAL_RECONNECT_MS;
  private connecting = false;
  private intentionalClose = false;
  private refCount = 0;

  private setStatus(status: WsConnectionStatus) {
    this.status = status;
    for (const listener of this.statusListeners) {
      listener(status);
    }
  }

  private scheduleReconnect() {
    if (this.intentionalClose || this.refCount === 0) return;
    if (this.reconnectTimer) return;

    this.setStatus("reconnecting");
    const delay = this.reconnectDelay;
    this.reconnectDelay = Math.min(this.reconnectDelay * 2, MAX_RECONNECT_MS);
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.ensureConnected();
    }, delay);
  }

  private resubscribeAll() {
    if (
      this.subscriptions.size === 0 ||
      !this.ws ||
      this.ws.readyState !== WebSocket.OPEN
    ) {
      return;
    }
    this.ws.send(
      JSON.stringify({
        action: "subscribe",
        symbols: Array.from(this.subscriptions),
      }),
    );
  }

  private handleMessage(event: MessageEvent) {
    try {
      const payload = JSON.parse(event.data as string) as {
        type?: string;
        data?: WsTrade[];
      };
      if (payload.type !== "trade" || !payload.data?.length) return;
      const raw = payload.data[payload.data.length - 1];
      const trade: WsTrade = {
        ...raw,
        t: normalizeTimestampMs(raw.t),
      };
      for (const listener of this.tradeListeners) {
        listener(trade);
      }
    } catch {
      // ignore malformed payloads
    }
  }

  private ensureConnected() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.setStatus("connected");
      return;
    }
    if (this.connecting) return;

    this.intentionalClose = false;
    this.connecting = true;
    this.setStatus("connecting");

    const ws = new WebSocket(buildWsUrl());
    this.ws = ws;

    ws.onopen = () => {
      this.connecting = false;
      this.reconnectDelay = INITIAL_RECONNECT_MS;
      this.setStatus("connected");
      this.resubscribeAll();
    };

    ws.onmessage = (event) => this.handleMessage(event);

    ws.onerror = () => {
      this.setStatus("error");
    };

    ws.onclose = () => {
      this.connecting = false;
      this.ws = null;
      if (!this.intentionalClose) {
        this.scheduleReconnect();
      } else {
        this.setStatus("disconnected");
      }
    };
  }

  connect() {
    this.refCount += 1;
    this.ensureConnected();
  }

  release() {
    this.refCount = Math.max(0, this.refCount - 1);
    if (this.refCount > 0) return;

    this.intentionalClose = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.setStatus("disconnected");
  }

  subscribe(symbols: string[]) {
    const normalized = symbols
      .map((s) => s.trim().toUpperCase())
      .filter(Boolean);
    const toSubscribe: string[] = [];

    for (const symbol of normalized) {
      const count = this.subscriptionCounts.get(symbol) ?? 0;
      this.subscriptionCounts.set(symbol, count + 1);
      if (count === 0) {
        this.subscriptions.add(symbol);
        toSubscribe.push(symbol);
      }
    }

    if (this.ws?.readyState === WebSocket.OPEN && toSubscribe.length > 0) {
      this.ws.send(
        JSON.stringify({ action: "subscribe", symbols: toSubscribe }),
      );
    }
  }

  unsubscribe(symbols: string[]) {
    const normalized = symbols
      .map((s) => s.trim().toUpperCase())
      .filter(Boolean);
    const toUnsubscribe: string[] = [];

    for (const symbol of normalized) {
      const count = this.subscriptionCounts.get(symbol) ?? 0;
      if (count <= 1) {
        this.subscriptionCounts.delete(symbol);
        this.subscriptions.delete(symbol);
        toUnsubscribe.push(symbol);
      } else {
        this.subscriptionCounts.set(symbol, count - 1);
      }
    }

    if (this.ws?.readyState === WebSocket.OPEN && toUnsubscribe.length > 0) {
      this.ws.send(
        JSON.stringify({ action: "unsubscribe", symbols: toUnsubscribe }),
      );
    }
  }

  addTradeListener(listener: TradeListener): () => void {
    this.tradeListeners.add(listener);
    return () => this.tradeListeners.delete(listener);
  }

  addStatusListener(listener: StatusListener): () => void {
    this.statusListeners.add(listener);
    listener(this.status);
    return () => this.statusListeners.delete(listener);
  }

  getStatus(): WsConnectionStatus {
    return this.status;
  }

  forceReconnect() {
    this.reconnectDelay = INITIAL_RECONNECT_MS;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.ensureConnected();
  }
}

export const stockFlowWs = new StockFlowWsClient();
