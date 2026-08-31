import { getPublicApiUrl } from "@/lib/env";
import { getAccessToken } from "@/lib/auth";
import { marketAuthFetch } from "@/lib/market/auth-fetch";
import { stockFlowWs } from "@/lib/market/ws-client";
import {
  BarAggregator,
  barBucketToTradingViewMs,
  tradingViewResolutionToMarket,
  tradingViewResolutionToSeconds,
} from "@/lib/market/bar-aggregator";

function barTimeMs(unixSec: number, resolution: string): number {
  return barBucketToTradingViewMs(unixSec, resolution);
}

function buildSymbolInfo(item: {
  symbol: string;
  name: string;
  description: string;
  exchange: string;
  type: string;
}): LibrarySymbolInfo {
  return {
    name: item.symbol,
    ticker: item.symbol,
    description: item.description || item.name,
    type: item.type || "stock",
    session: "0930-1600",
    timezone: "America/New_York",
    exchange: item.exchange || "US",
    listed_exchange: item.exchange || "US",
    minmov: 1,
    pricescale: 100,
    format: "price",
    has_intraday: true,
    has_daily: true,
    has_weekly_and_monthly: true,
    has_empty_bars: false,
    visible_plots_set: "ohlcv",
    supported_resolutions: CONFIG.supported_resolutions,
    intraday_multipliers: ["1", "5", "15", "60"],
    daily_multipliers: ["1"],
    weekly_multipliers: ["1"],
    volume_precision: 2,
    data_status: "streaming",
  };
}

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
  listed_exchange?: string;
  minmov: number;
  pricescale: number;
  format?: string;
  has_intraday: boolean;
  has_daily: boolean;
  has_weekly_and_monthly: boolean;
  has_empty_bars?: boolean;
  visible_plots_set?: string;
  supported_resolutions: string[];
  intraday_multipliers?: string[];
  daily_multipliers?: string[];
  weekly_multipliers?: string[];
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
  offTrade?: () => void;
};

export type StockFlowDatafeedOptions = {
  apiUrl?: string;
  getAccessToken?: () => string | null;
};

const CONFIG: DatafeedConfiguration = {
  supported_resolutions: ["1", "5", "15", "60", "1D", "1W"],
  supports_search: true,
  supports_group_request: false,
  supports_marks: false,
  supports_timescale_marks: false,
  supports_time: false,
};

export class StockFlowDatafeed {
  private apiUrl: string;
  private getToken: () => string | null;
  private subscriptions = new Map<string, SubscribeState>();

  constructor(options: StockFlowDatafeedOptions = {}) {
    this.apiUrl = (options.apiUrl ?? getPublicApiUrl()).replace(/\/$/, "");
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
    this.authFetch(
      `${this.apiUrl}/api/market/symbols/${encodeURIComponent(symbolName)}`,
    )
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
        onResolve(buildSymbolInfo(item));
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
        const bars = data.t
          .map((time, i) => ({
            time: barTimeMs(time, resolution),
            open: data.o![i],
            high: data.h![i],
            low: data.l![i],
            close: data.c![i],
            volume: data.v?.[i],
          }))
          .filter(
            (bar) =>
              bar.open > 0 &&
              bar.high > 0 &&
              bar.low > 0 &&
              bar.close > 0 &&
              bar.high >= bar.low,
          )
          .sort((a, b) => a.time - b.time);
        onResult(bars, { noData: bars.length === 0 });
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
    const symbol = symbolInfo.ticker.toUpperCase();
    const state: SubscribeState = {
      symbol,
      resolution,
      onTick,
      aggregator: new BarAggregator(seconds),
    };

    stockFlowWs.connect();
    stockFlowWs.subscribe([symbol]);

    state.offTrade = stockFlowWs.addTradeListener((trade) => {
      if (trade.s !== symbol) return;
      const bucket = state.aggregator.onTick(trade.p, trade.v, trade.t);
      if (!bucket) return;
      onTick({
        time: barBucketToTradingViewMs(bucket.time, resolution),
        open: bucket.open,
        high: bucket.high,
        low: bucket.low,
        close: bucket.close,
        volume: bucket.volume,
      });
    });

    this.subscriptions.set(listenerGuid, state);
  }

  unsubscribeBars(listenerGuid: string) {
    const state = this.subscriptions.get(listenerGuid);
    if (!state) return;
    if (state.offTrade) {
      state.offTrade();
    }
    stockFlowWs.unsubscribe([state.symbol]);
    this.subscriptions.delete(listenerGuid);
  }

  private authFetch(url: string) {
    return marketAuthFetch(url);
  }
}
