"use client";

import { motion } from "framer-motion";
import {
  Activity,
  BarChart3,
  Layers,
  LineChart,
  Radio,
  Search,
  Zap,
} from "lucide-react";
import { easeOut } from "./motion";

const bentoItems = [
  {
    id: "charts",
    title: "Professional charts",
    description:
      "TradingView Advanced Charts with studies, drawing tools, multi-timeframe views, and a dark terminal aesthetic.",
    icon: LineChart,
    className: "md:col-span-2 md:row-span-2",
    accent: "from-emerald-500/20 via-teal-500/10 to-transparent",
    large: true,
  },
  {
    id: "websocket",
    title: "Live WebSocket feed",
    description: "Real-time trade ticks via Finnhub — prices update as markets move.",
    icon: Zap,
    className: "md:col-span-1",
    accent: "from-teal-500/15 to-transparent",
    stat: "<50ms",
    statLabel: "avg latency",
  },
  {
    id: "watchlist",
    title: "Watchlists",
    description: "Track symbols across sessions with persistent watchlists.",
    icon: Layers,
    className: "md:col-span-1",
    accent: "from-emerald-600/10 to-transparent",
  },
  {
    id: "search",
    title: "Symbol search",
    description: "Instant lookup across US equities.",
    icon: Search,
    className: "md:col-span-1",
    accent: "from-zinc-500/10 to-transparent",
  },
  {
    id: "stats",
    title: "Session stats",
    description: "Day range, volume, and change at a glance in the status bar.",
    icon: BarChart3,
    className: "md:col-span-1",
    accent: "from-emerald-500/10 to-transparent",
  },
  {
    id: "flow",
    title: "Built for flow",
    description:
      "Layout designed for focused analysis — minimal chrome, maximum chart real estate.",
    icon: Activity,
    className: "md:col-span-2",
    accent: "from-emerald-500/15 via-transparent to-teal-500/10",
    wide: true,
  },
];

function BentoCell({
  item,
  index,
}: {
  item: (typeof bentoItems)[number];
  index: number;
}) {
  const Icon = item.icon;

  return (
    <motion.div
      initial={{ opacity: 0, y: 24 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-40px" }}
      transition={{ delay: index * 0.06, duration: 0.65, ease: easeOut }}
      className={`group relative ${item.className}`}
    >
      <div className="absolute -inset-px rounded-2xl bg-gradient-to-br from-emerald-500/20 via-white/[0.04] to-transparent opacity-0 transition-opacity duration-500 group-hover:opacity-100" />

      <div className="relative flex h-full flex-col overflow-hidden rounded-2xl border border-white/[0.06] bg-zinc-900/30 p-6 backdrop-blur-sm transition-colors duration-500 group-hover:border-emerald-500/15 group-hover:bg-zinc-900/50">
        <div
          className={`pointer-events-none absolute inset-0 bg-gradient-to-br ${item.accent} opacity-60`}
        />

        <div className="relative flex flex-1 flex-col">
          <div className="mb-4 flex items-start justify-between">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10 ring-1 ring-emerald-500/20 transition-all duration-500 group-hover:bg-emerald-500/15 group-hover:shadow-[0_0_20px_rgba(16,185,129,0.15)]">
              <Icon className="h-5 w-5 text-emerald-400" />
            </div>
            {item.stat && (
              <div className="text-right">
                <p className="font-mono text-xl font-medium text-emerald-400">
                  {item.stat}
                </p>
                <p className="font-mono text-[10px] uppercase tracking-wider text-zinc-600">
                  {item.statLabel}
                </p>
              </div>
            )}
          </div>

          <h3
            className={`font-display font-semibold text-white ${item.large ? "text-2xl" : "text-lg"}`}
          >
            {item.title}
          </h3>
          <p
            className={`mt-2 leading-relaxed text-zinc-400 ${item.large ? "text-sm sm:max-w-md" : "text-sm"}`}
          >
            {item.description}
          </p>

          {item.large && (
            <div className="mt-auto pt-8">
              <div className="flex items-end gap-1 opacity-40">
                {Array.from({ length: 24 }).map((_, i) => (
                  <div
                    key={i}
                    className="w-1.5 rounded-sm bg-gradient-to-t from-emerald-500/30 to-emerald-400/70"
                    style={{ height: `${12 + ((i * 7) % 5) * 8}px` }}
                  />
                ))}
              </div>
            </div>
          )}

          {item.wide && (
            <div className="mt-4 flex items-center gap-2 font-mono text-[10px] text-zinc-600">
              <Radio className="h-3 w-3 text-emerald-500/60" />
              <span>Optimized for single-monitor workflows</span>
            </div>
          )}
        </div>
      </div>
    </motion.div>
  );
}

export function FeaturesBento() {
  return (
    <section id="features" className="relative mx-auto max-w-7xl px-6 py-24 lg:py-32">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, margin: "-80px" }}
        transition={{ duration: 0.6, ease: easeOut }}
        className="mb-14 max-w-2xl"
      >
        <p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-emerald-500/80">
          Capabilities
        </p>
        <h2 className="font-display text-3xl font-semibold tracking-tight text-white sm:text-4xl lg:text-5xl">
          Everything you need to watch the tape
        </h2>
        <p className="mt-4 text-zinc-400">
          Inspired by professional terminals — built with modern web tooling.
        </p>
      </motion.div>

      <div className="grid auto-rows-fr gap-4 md:grid-cols-3">
        {bentoItems.map((item, i) => (
          <BentoCell key={item.id} item={item} index={i} />
        ))}
      </div>
    </section>
  );
}
