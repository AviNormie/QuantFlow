"use client";

import { motion } from "framer-motion";
import { easeOut } from "./motion";

const candles = [
  { x: 0, o: 72, c: 58, h: 78, l: 52, bull: false },
  { x: 1, o: 58, c: 65, h: 70, l: 55, bull: true },
  { x: 2, o: 65, c: 52, h: 68, l: 48, bull: false },
  { x: 3, o: 52, c: 60, h: 64, l: 50, bull: true },
  { x: 4, o: 60, c: 72, h: 76, l: 58, bull: true },
  { x: 5, o: 72, c: 68, h: 80, l: 64, bull: false },
  { x: 6, o: 68, c: 78, h: 82, l: 66, bull: true },
  { x: 7, o: 78, c: 70, h: 82, l: 66, bull: false },
  { x: 8, o: 70, c: 82, h: 86, l: 68, bull: true },
  { x: 9, o: 82, c: 76, h: 88, l: 72, bull: false },
  { x: 10, o: 76, c: 88, h: 92, l: 74, bull: true },
  { x: 11, o: 88, c: 84, h: 94, l: 80, bull: false },
  { x: 12, o: 84, c: 92, h: 96, l: 82, bull: true },
  { x: 13, o: 92, c: 86, h: 96, l: 82, bull: false },
  { x: 14, o: 86, c: 94, h: 98, l: 84, bull: true },
  { x: 15, o: 94, c: 90, h: 100, l: 86, bull: false },
  { x: 16, o: 90, c: 98, h: 102, l: 88, bull: true },
  { x: 17, o: 98, c: 92, h: 102, l: 88, bull: false },
  { x: 18, o: 92, c: 100, h: 104, l: 90, bull: true },
  { x: 19, o: 100, c: 96, h: 106, l: 92, bull: false },
];

function CandlestickChart() {
  const w = 520;
  const h = 140;
  const pad = 8;
  const candleW = (w - pad * 2) / candles.length;

  return (
    <svg
      viewBox={`0 0 ${w} ${h}`}
      className="h-full w-full"
      preserveAspectRatio="none"
    >
      <defs>
        <linearGradient id="chartFade" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="rgba(16,185,129,0.08)" />
          <stop offset="100%" stopColor="rgba(16,185,129,0)" />
        </linearGradient>
      </defs>

      {[20, 40, 60, 80, 100, 120].map((y) => (
        <line
          key={y}
          x1={pad}
          y1={y}
          x2={w - pad}
          y2={y}
          stroke="rgba(255,255,255,0.04)"
          strokeWidth="1"
        />
      ))}

      <motion.path
        d={`M ${pad} ${h - pad} ${candles
          .map((c, i) => {
            const cx = pad + i * candleW + candleW / 2;
            return `L ${cx} ${h - c.c}`;
          })
          .join(" ")} L ${w - pad} ${h - pad} Z`}
        fill="url(#chartFade)"
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        viewport={{ once: true }}
        transition={{ duration: 1.2, ease: easeOut }}
      />

      {candles.map((c, i) => {
        const cx = pad + i * candleW + candleW / 2;
        const bodyTop = Math.min(c.o, c.c);
        const bodyH = Math.abs(c.c - c.o) || 2;
        const color = c.bull ? "#34d399" : "#f87171";

        return (
          <motion.g
            key={i}
            initial={{ opacity: 0, scaleY: 0 }}
            whileInView={{ opacity: 1, scaleY: 1 }}
            viewport={{ once: true }}
            transition={{
              delay: 0.02 * i,
              duration: 0.5,
              ease: easeOut,
            }}
            style={{ transformOrigin: `${cx}px ${h}px` }}
          >
            <line
              x1={cx}
              y1={h - c.h}
              x2={cx}
              y2={h - c.l}
              stroke={color}
              strokeWidth="1"
              opacity="0.7"
            />
            <rect
              x={cx - candleW * 0.28}
              y={h - bodyTop - bodyH}
              width={candleW * 0.56}
              height={bodyH}
              fill={color}
              rx="1"
            />
          </motion.g>
        );
      })}

      <motion.polyline
        points={candles
          .map((c, i) => {
            const cx = pad + i * candleW + candleW / 2;
            return `${cx},${h - c.c}`;
          })
          .join(" ")}
        fill="none"
        stroke="rgba(52,211,153,0.4)"
        strokeWidth="1.5"
        initial={{ pathLength: 0, opacity: 0 }}
        whileInView={{ pathLength: 1, opacity: 1 }}
        viewport={{ once: true }}
        transition={{ duration: 1.4, ease: easeOut, delay: 0.3 }}
      />
    </svg>
  );
}

const orderBook = [
  { price: "189.48", size: "1.2K", side: "ask" },
  { price: "189.45", size: "840", side: "ask" },
  { price: "189.42", size: "2.1K", side: "bid" },
  { price: "189.40", size: "1.5K", side: "bid" },
  { price: "189.38", size: "920", side: "bid" },
];

export function TerminalPreview() {
  return (
    <section className="relative mx-auto max-w-7xl px-6 py-20 lg:py-28">
      <motion.div
        initial={{ opacity: 0, y: 40 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, margin: "-80px" }}
        transition={{ duration: 0.8, ease: easeOut }}
        className="relative"
      >
        <div className="absolute -inset-px rounded-2xl bg-gradient-to-b from-emerald-500/30 via-white/[0.06] to-transparent opacity-60" />

        <div className="relative overflow-hidden rounded-2xl border border-white/[0.08] bg-zinc-900/40 shadow-2xl shadow-black/50 backdrop-blur-2xl">
          <div className="flex items-center gap-2 border-b border-white/[0.06] bg-zinc-950/60 px-4 py-3">
            <div className="h-2.5 w-2.5 rounded-full bg-rose-500/80" />
            <div className="h-2.5 w-2.5 rounded-full bg-amber-500/80" />
            <div className="h-2.5 w-2.5 rounded-full bg-emerald-500/80" />
            <span className="ml-3 font-mono text-[11px] text-zinc-500">
              stockflow.app/charts — AAPL · 1D · live
            </span>
            <span className="ml-auto flex items-center gap-1.5 font-mono text-[10px] text-emerald-400">
              <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-emerald-500" />
              CONNECTED
            </span>
          </div>

          <div className="grid lg:grid-cols-[1fr_200px]">
            <div className="border-b border-white/[0.04] p-5 lg:border-b-0 lg:border-r">
              <div className="mb-4 flex flex-wrap items-baseline gap-x-4 gap-y-1">
                <span className="font-display text-2xl font-semibold text-white">
                  AAPL
                </span>
                <span className="font-mono text-2xl text-emerald-400">
                  189.42
                </span>
                <span className="font-mono text-sm text-emerald-400/80">
                  +2.34 (+1.25%)
                </span>
              </div>

              <div className="h-[140px] overflow-hidden rounded-lg bg-zinc-950/50 ring-1 ring-white/[0.04]">
                <CandlestickChart />
              </div>

              <div className="mt-4 flex gap-2 font-mono text-[10px] text-zinc-600">
                {["1m", "5m", "15m", "1H", "4H", "1D", "1W"].map((tf, i) => (
                  <span
                    key={tf}
                    className={
                      i === 5
                        ? "rounded bg-emerald-500/15 px-2 py-0.5 text-emerald-400"
                        : "px-2 py-0.5"
                    }
                  >
                    {tf}
                  </span>
                ))}
              </div>
            </div>

            <div className="hidden bg-zinc-950/30 p-4 lg:block">
              <p className="mb-3 font-mono text-[10px] uppercase tracking-wider text-zinc-600">
                Order book
              </p>
              <div className="space-y-1.5">
                {orderBook.map((row) => (
                  <div
                    key={row.price}
                    className="flex justify-between font-mono text-[11px]"
                  >
                    <span
                      className={
                        row.side === "bid" ? "text-emerald-400" : "text-rose-400"
                      }
                    >
                      {row.price}
                    </span>
                    <span className="text-zinc-600">{row.size}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="flex items-center justify-between border-t border-white/[0.04] bg-zinc-950/40 px-5 py-2.5 font-mono text-[10px] text-zinc-600">
            <span>Vol 48.2M</span>
            <span>H 190.12 · L 186.80</span>
            <span className="hidden sm:inline">Last tick 14:32:08.241</span>
          </div>
        </div>

        <motion.div
          className="absolute -right-4 -top-4 hidden rounded-lg border border-emerald-500/20 bg-zinc-900/90 px-3 py-2 font-mono text-[10px] text-emerald-400 shadow-lg backdrop-blur-xl lg:block"
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.5, duration: 0.6, ease: easeOut }}
        >
          ws://feed · 42ms
        </motion.div>
      </motion.div>
    </section>
  );
}
