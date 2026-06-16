"use client";

import { motion } from "framer-motion";

const tickers = [
  { symbol: "AAPL", price: "189.42", change: "+1.24%", up: true },
  { symbol: "NVDA", price: "875.30", change: "+2.87%", up: true },
  { symbol: "MSFT", price: "415.18", change: "-0.32%", up: false },
  { symbol: "TSLA", price: "248.91", change: "+0.95%", up: true },
  { symbol: "AMZN", price: "178.65", change: "+0.41%", up: true },
  { symbol: "META", price: "512.33", change: "-0.18%", up: false },
  { symbol: "GOOGL", price: "171.22", change: "+1.05%", up: true },
  { symbol: "SPY", price: "523.80", change: "+0.67%", up: true },
];

function TickerItem({
  symbol,
  price,
  change,
  up,
}: (typeof tickers)[number]) {
  return (
    <div className="flex shrink-0 items-center gap-3 px-6 font-mono text-xs">
      <span className="font-medium text-zinc-300">{symbol}</span>
      <span className="text-zinc-500">{price}</span>
      <span className={up ? "text-emerald-400" : "text-rose-400"}>
        {change}
      </span>
    </div>
  );
}

export function TickerStrip() {
  const doubled = [...tickers, ...tickers];

  return (
    <div className="relative w-full overflow-hidden border-y border-white/[0.06] bg-zinc-950/60 py-3 backdrop-blur-md">
      <div className="pointer-events-none absolute inset-y-0 left-0 z-10 w-16 bg-gradient-to-r from-zinc-950 to-transparent" />
      <div className="pointer-events-none absolute inset-y-0 right-0 z-10 w-16 bg-gradient-to-l from-zinc-950 to-transparent" />

      <motion.div
        className="flex w-max"
        animate={{ x: ["0%", "-50%"] }}
        transition={{
          duration: 40,
          repeat: Infinity,
          ease: "linear",
        }}
      >
        {doubled.map((t, i) => (
          <TickerItem key={`${t.symbol}-${i}`} {...t} />
        ))}
      </motion.div>
    </div>
  );
}
