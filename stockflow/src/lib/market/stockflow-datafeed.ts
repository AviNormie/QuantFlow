import { getPublicApiUrl, getPublicWebSocketUrl } from "@/lib/env";
import { getAccessToken } from "@/lib/auth";
import {
  BarAggregator,
  tradingViewResolutionToMarket,
  tradingViewResolutionToSeconds,
} from "@/lib/market/bar-aggregator";

type DatafeedConfiguration = {
  supported_resolutions: string[];
  supports_search: boolean;
  supports_group_request: boolean;
  supports_marks: boolean;
  supports_timescale_marks: boolean;
  supports_time: boolean;
};

type LibrarySymbolInfo = {
  name: string;
  ticker: string;
  description: string;
  type: string;
  session: string;
  timezone: string;
  exchange: string;
  minmov: number;
  pricescale: number;
  has_intraday: boolean;
  has_daily: boolean;
  has_weekly_and_monthly: boolean;
  supported_resolutions: string[];
  volume_precision: number;
  data_status: string;
};

type PeriodParams = {
  from: number;
  to: number;
  firstDataRequest: boolean;
};

type Bar = {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume?: number;
};

type SubscribeState = {
  symbol: string;
  resolution: string;
  onTick: (bar: Bar) => void;
  aggregator: BarAggregator;
  ws: WebSocket | null;
};

export type StockFlowDatafeedOptions = {
  apiUrl?: string;
  wsUrl?: string;
  getAccessToken?: () => string | null;
};

const CONFIG: DatafeedConfiguration = {
  supported_resolutions: ["1", "5", "15", "60", "1D", "1W"],
  supports_search: true,
  supports_group_request: false,
  supports_marks: false,
  supports_timescale_marks: false,
  supports_time: true,
};

export class StockFlowDatafeed {
  private apiUrl: string;
  private wsUrl: string;
  private getToken: () => string | null;
  private subscriptions = new Map<string, SubscribeState>();

  constructor(options: StockFlowDatafeedOptions = {}) {
    this.apiUrl = (options.apiUrl ?? getPublicApiUrl()).replace(/\/$/, "");
    this.wsUrl = options.wsUrl ?? getPublicWebSocketUrl();
    this.getToken = options.getAccessToken ?? getAccessToken;
  }

  onReady(callback: (config: DatafeedConfiguration) => void) {
    setTimeout(() => callback(CONFIG), 0);
  }

  searchSymbols(
    userInput: string,
    _exchange: string,
    _symbolType: string,
    onResult: (items: Array<Record<string, string>>) => void,
  ) {
    const trimmed = userInput.trim();
    const params = new URLSearchParams({ limit: "30" });
    if (trimmed) {
      params.set("q", trimmed);
    }
    this.authFetch(`${this.apiUrl}/api/market/symbols/search?${params}`)
      .then(async (res) => {
        if (!res.ok) {
          onResult([]);
          return;
        }
        const data = (await res.json()) as {
          results?: Array<{
            symbol: string;
            name: string;
            description: string;
            exchange: string;
            type: string;
          }>;
        };
        onResult(
          (data.results ?? []).map((item) => ({
            symbol: item.symbol,
            full_name: item.symbol,
            description: item.description || item.name,
            exchange: item.exchange,
            ticker: item.symbol,
            type: item.type || "stock",
          })),
        );
      })
      .catch(() => onResult([]));
  }

  resolveSymbol(
    symbolName: string,
    onResolve: (info: LibrarySymbolInfo) => void,
    onError: (reason: string) => void,
  ) {
    this.authFetch(`${this.apiUrl}/api/market/symbols/${encodeURIComponent(symbolName)}`)
      .then(async (res) => {
        if (!res.ok) {
          onError("symbol not found");
          return;
        }
        const item = (await res.json()) as {
          symbol: string;
          name: string;
          description: string;
          exchange: string;
          type: string;
        };
        onResolve({
          name: item.symbol,
          ticker: item.symbol,
          description: item.description || item.name,
          type: item.type || "stock",
          session: "0930-1600",
          timezone: "America/New_York",
          exchange: item.exchange || "US",
          minmov: 1,
          pricescale: 100,
          has_intraday: true,
          has_daily: true,
          has_weekly_and_monthly: true,
          supported_resolutions: CONFIG.supported_resolutions,
          volume_precision: 2,
          data_status: "streaming",
        });
      })
      .catch(() => onError("resolve failed"));
  }

  getBars(
    symbolInfo: LibrarySymbolInfo,
    resolution: string,
    periodParams: PeriodParams,
    onResult: (bars: Bar[], meta: { noData?: boolean }) => void,
    onError: (reason: string) => void,
  ) {
    const marketResolution = tradingViewResolutionToMarket(resolution);
    const params = new URLSearchParams({
      resolution: marketResolution,
      from: String(periodParams.from),
      to: String(periodParams.to),
    });
    const url = `${this.apiUrl}/api/market/candles/${encodeURIComponent(symbolInfo.ticker)}?${params}`;

    this.authFetch(url)
      .then(async (res) => {
        if (!res.ok) {
          onError("failed to load bars");
          return;
        }
        const data = (await res.json()) as {
          s: string;
          t?: number[];
          o?: number[];
          h?: number[];
          l?: number[];
          c?: number[];
          v?: number[];
        };
        if (data.s !== "ok" || !data.t?.length) {
          onResult([], { noData: true });
          return;
        }
        const bars = data.t.map((time, i) => ({
          time: time * 1000,
          open: data.o![i],
          high: data.h![i],
          low: data.l![i],
          close: data.c![i],
          volume: data.v?.[i],
        }));
        onResult(bars, { noData: false });
      })
      .catch(() => onError("failed to load bars"));
  }

  subscribeBars(
    symbolInfo: LibrarySymbolInfo,
    resolution: string,
    onTick: (bar: Bar) => void,
    listenerGuid: string,
  ) {
    const seconds = tradingViewResolutionToSeconds(resolution);
    const state: SubscribeState = {
      symbol: symbolInfo.ticker,
      resolution,
      onTick,
      aggregator: new BarAggregator(seconds),
      ws: null,
    };

    const base = this.wsUrl.endsWith("/") ? this.wsUrl : `${this.wsUrl}/`;
    const wsUrl = new URL("/ws", base);
    const token = this.getToken();
    if (token) {
      wsUrl.searchParams.set("token", token);
    }

    const ws = new WebSocket(wsUrl.toString());
    state.ws = ws;

    ws.onopen = () => {
      ws.send(
        JSON.stringify({
          action: "subscribe",
          symbols: [symbolInfo.ticker],
        }),
      );
    };

    ws.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data as string) as {
          type?: string;
          data?: Array<{ s: string; p: number; v: number; t: number }>;
        };
        if (payload.type !== "trade" || !payload.data?.length) return;
        const trade = payload.data[payload.data.length - 1];
        if (trade.s !== symbolInfo.ticker) return;
        const bar = state.aggregator.onTick(trade.p, trade.v, trade.t);
        if (bar) onTick(bar);
      } catch {
        // ignore malformed payloads
      }
    };

    this.subscriptions.set(listenerGuid, state);
  }

  unsubscribeBars(listenerGuid: string) {
    const state = this.subscriptions.get(listenerGuid);
    if (!state) return;
    if (state.ws) {
      state.ws.close();
    }
    this.subscriptions.delete(listenerGuid);
  }

  private authFetch(url: string) {
    const headers: Record<string, string> = {};
    const token = this.getToken();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    return fetch(url, { headers });
  }
}
