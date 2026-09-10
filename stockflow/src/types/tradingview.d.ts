export {};

declare global {
  interface Window {
    TradingView: {
      widget: new (options: Record<string, unknown>) => TradingViewWidget;
    };
    Datafeeds: {
      UDFCompatibleDatafeed: new (url: string) => unknown;
    };
  }

  interface TradingViewWidget {
    onChartReady: (callback: () => void) => void;
    activeChart: () => TradingViewActiveChart;
    remove: () => void;
    setSymbol: (
      symbol: string,
      interval: string,
      callback?: () => void,
    ) => void;
  }

  interface TradingViewActiveChart {
    setSymbol: (
      symbol: string,
      callback?: () => void,
    ) => void | Promise<boolean>;
    setResolution: (
      interval: string,
      callback?: () => void,
    ) => void | Promise<boolean>;
    applyOverrides: (overrides: Record<string, string | number>) => void;
  }
}
