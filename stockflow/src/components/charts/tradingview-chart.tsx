"use client";

import { useEffect, useRef, useState } from "react";
import { getAccessToken } from "@/lib/auth";
import { StockFlowDatafeed } from "@/lib/market/stockflow-datafeed";

type TradingViewChartProps = {
  symbol: string;
  interval: string;
  className?: string;
};

const STOCKFLOW_OVERRIDES: Record<string, string | number> = {
  "paneProperties.background": "#09090b",
  "paneProperties.backgroundType": "solid",
  "paneProperties.backgroundGradientStartColor": "#09090b",
  "paneProperties.backgroundGradientEndColor": "#09090b",
  "mainSeriesProperties.background": "#09090b",
  "chartProperties.background": "#09090b",
  "chartProperties.backgroundType": "solid",
  "chartProperties.backgroundGradientStartColor": "#09090b",
  "chartProperties.backgroundGradientEndColor": "#09090b",
  "paneProperties.vertGridProperties.color": "rgba(255,255,255,0.04)",
  "paneProperties.horzGridProperties.color": "rgba(255,255,255,0.04)",
  "paneProperties.vertGridProperties.style": 0,
  "paneProperties.horzGridProperties.style": 0,
  "scalesProperties.backgroundColor": "#09090b",
  "scalesProperties.textColor": "#a1a1aa",
  "scalesProperties.lineColor": "rgba(255,255,255,0.08)",
  "scalesProperties.fontSize": 11,
  "timeScale.backgroundColor": "#09090b",
  "timeScale.textColor": "#a1a1aa",
  "timeScale.fontSize": 11,
  "crossHairProperties.color": "rgba(255,255,255,0.15)",
  "crossHairProperties.width": 1,
  "crossHairProperties.style": 2,
  "symbolWatermarkProperties.color": "#27272a",
  "symbolWatermarkProperties.transparency": 90,
  "mainSeriesProperties.candleStyle.upColor": "#34d399",
  "mainSeriesProperties.candleStyle.downColor": "#fb7185",
  "mainSeriesProperties.candleStyle.wickUpColor": "#34d399",
  "mainSeriesProperties.candleStyle.wickDownColor": "#fb7185",
  "mainSeriesProperties.candleStyle.borderUpColor": "#34d399",
  "mainSeriesProperties.candleStyle.borderDownColor": "#fb7185",
  "mainSeriesProperties.priceLineColor": "#a1a1aa",
  "mainSeriesProperties.priceLineWidth": 1,
};

function waitForTradingView(): Promise<void> {
  return new Promise((resolve) => {
    const check = () => {
      if (window.TradingView?.widget) {
        resolve();
        return;
      }
      setTimeout(check, 100);
    };
    check();
  });
}

export function TradingViewChart({
  symbol,
  interval,
  className,
}: TradingViewChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetRef = useRef<TradingViewWidget | null>(null);
  const datafeedRef = useRef<StockFlowDatafeed | null>(null);
  const propsRef = useRef({ symbol, interval });
  propsRef.current = { symbol, interval };

  const [isReady, setIsReady] = useState(false);
  const [chartReady, setChartReady] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    waitForTradingView()
      .then(() => {
        if (!cancelled) setIsReady(true);
      })
      .catch(() => {
        if (!cancelled) {
          setLoadError("Failed to load TradingView chart library.");
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!isReady || !containerRef.current) return;

    const container = containerRef.current;
    container.innerHTML = "";
    setChartReady(false);

    if (widgetRef.current) {
      try {
        widgetRef.current.remove();
      } catch {
        // Widget may already be removed.
      }
      widgetRef.current = null;
    }

    const height = Math.max(container.clientHeight || 420, 420);
    const datafeed = new StockFlowDatafeed({ getAccessToken });
    datafeedRef.current = datafeed;

    try {
      const widget = new window.TradingView.widget({
        container,
        locale: "en",
        library_path: "/charting_library/",
        datafeed,
        symbol,
        interval,
        height,
        width: "100%",
        fullscreen: false,
        autosize: true,
        debug: false,
        theme: "dark",
        disabled_features: ["chart_property_page_background"],
        enabled_features: [
          "study_templates",
          "side_toolbar_in_fullscreen_mode",
          "header_in_fullscreen_mode",
          "header_symbol_search",
          "symbol_search_hot_key",
          "allow_arbitrary_symbol_search_input",
        ],
        toolbar_bg: "#18181b",
        loading_screen: {
          backgroundColor: "#09090b",
          foregroundColor: "#a1a1aa",
        },
        overrides: STOCKFLOW_OVERRIDES,
        studies_overrides: {
          "volume.volume.color.0": "#fb7185",
          "volume.volume.color.1": "#34d399",
          "volume.volume.transparency": 50,
        },
      });

      widgetRef.current = widget;

      widget.onChartReady(() => {
        setChartReady(true);
        try {
          const chart = widget.activeChart();
          chart.applyOverrides(STOCKFLOW_OVERRIDES);
        } catch {
          // Chart overrides are best-effort.
        }
      });
    } catch (error) {
      setLoadError(
        error instanceof Error ? error.message : "Failed to create chart widget.",
      );
    }

    return () => {
      setChartReady(false);
      datafeedRef.current = null;
      if (widgetRef.current) {
        try {
          widgetRef.current.remove();
        } catch {
          // Widget may already be removed.
        }
        widgetRef.current = null;
      }
    };
    // Recreate only on library readiness; symbol/interval updates handled below.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isReady]);

  useEffect(() => {
    const widget = widgetRef.current;
    if (!widget || !chartReady) return;

    try {
      widget.activeChart().setSymbol(symbol, interval);
    } catch {
      // Widget may still be initializing.
    }
  }, [symbol, interval, chartReady]);

  if (loadError) {
    return (
      <div className={className} style={{ minHeight: 420 }}>
        <div className="flex h-full min-h-[420px] flex-col items-center justify-center gap-2 p-8 text-center">
          <p className="text-sm text-zinc-400">{loadError}</p>
          <p className="text-xs text-zinc-500">
            Ensure charting_library scripts are loaded in the app layout.
          </p>
        </div>
      </div>
    );
  }

  if (!isReady) {
    return (
      <div className={className} style={{ minHeight: 420 }}>
        <div className="flex h-full min-h-[420px] items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-emerald-500/30 border-t-emerald-400" />
        </div>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className={className}
      style={{ width: "100%", height: "100%", minHeight: 420 }}
    />
  );
}
