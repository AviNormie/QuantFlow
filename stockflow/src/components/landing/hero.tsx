"use client";

import Link from "next/link";
import { motion, useScroll, useTransform } from "framer-motion";
import { ArrowRight } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { fadeUp } from "./motion";
import { TickerStrip } from "./ticker-strip";

export function Hero() {
  const { scrollY } = useScroll();
  const heroY = useTransform(scrollY, [0, 500], [0, 100]);
  const heroOpacity = useTransform(scrollY, [0, 400], [1, 0.3]);

  return (
    <motion.section
      style={{ y: heroY, opacity: heroOpacity }}
      className="relative flex min-h-screen flex-col"
    >
      <div className="flex flex-1 flex-col items-center justify-center px-6 pt-24 pb-12">
        <div className="mx-auto max-w-7xl text-center">
          <motion.div
            custom={0}
            initial="hidden"
            animate="visible"
            variants={fadeUp}
          >
            <Badge
              variant="secondary"
              className="mb-8 border border-emerald-500/20 bg-emerald-500/5 px-4 py-1.5 font-mono text-[11px] uppercase tracking-widest text-emerald-400/90"
            >
              Real-time market intelligence
            </Badge>
          </motion.div>

          <motion.h1
            custom={1}
            initial="hidden"
            animate="visible"
            variants={fadeUp}
            className="font-display mx-auto max-w-5xl text-[clamp(2.75rem,8vw,5.5rem)] font-semibold leading-[1.02] tracking-tight text-white"
          >
            Trade with
            <span className="relative mx-3 inline-block">
              <span className="relative z-10">clarity</span>
              <motion.span
                className="absolute -inset-x-2 bottom-1 h-3 bg-emerald-500/20"
                initial={{ scaleX: 0 }}
                animate={{ scaleX: 1 }}
                transition={{ delay: 0.6, duration: 0.8, ease: [0.22, 1, 0.36, 1] }}
                style={{ originX: 0 }}
              />
            </span>
            .
          </motion.h1>

          <motion.p
            custom={2}
            initial="hidden"
            animate="visible"
            variants={fadeUp}
            className="mx-auto mt-8 max-w-2xl text-lg leading-relaxed text-zinc-400 sm:text-xl"
          >
            <span className="bg-gradient-to-r from-emerald-300 via-teal-200 to-emerald-400 bg-clip-text font-medium text-transparent">
              Charts that move with the market.
            </span>{" "}
            Institutional-grade charting meets live WebSocket data in a terminal
            built for focused analysis.
          </motion.p>

          <motion.div
            custom={3}
            initial="hidden"
            animate="visible"
            variants={fadeUp}
            className="mt-12 flex flex-col items-center justify-center gap-4 sm:flex-row"
          >
            <Button
              size="lg"
              asChild
              className="group relative overflow-hidden shadow-lg shadow-emerald-500/25 transition-shadow hover:shadow-emerald-500/40"
            >
              <Link href="/charts">
                <span className="absolute inset-0 bg-gradient-to-r from-emerald-400/0 via-emerald-300/20 to-emerald-400/0 opacity-0 transition-opacity group-hover:opacity-100" />
                Launch charts
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild>
              <Link href="/charts">Explore watchlist</Link>
            </Button>
          </motion.div>

          <motion.div
            custom={4}
            initial="hidden"
            animate="visible"
            variants={fadeUp}
            className="mt-16 flex flex-wrap items-center justify-center gap-x-8 gap-y-3 font-mono text-[11px] uppercase tracking-wider text-zinc-600"
          >
            <span className="flex items-center gap-2">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.6)]" />
              Live feed
            </span>
            <span>TradingView charts</span>
            <span>Sub-50ms latency</span>
          </motion.div>
        </div>
      </div>

      <TickerStrip />
    </motion.section>
  );
}
