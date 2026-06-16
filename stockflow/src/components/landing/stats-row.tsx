"use client";

import { useEffect, useRef, useState } from "react";
import { motion, useInView } from "framer-motion";
import { easeOut } from "./motion";

const stats = [
  { value: 50, suffix: "ms", label: "Avg tick latency", prefix: "<" },
  { value: 12, suffix: "K+", label: "US equities", prefix: "" },
  { value: 24, suffix: "/7", label: "Market hours coverage", prefix: "" },
  { value: 99.9, suffix: "%", label: "Feed uptime", prefix: "", decimals: 1 },
];

function AnimatedCounter({
  value,
  prefix = "",
  suffix = "",
  decimals = 0,
  inView,
}: {
  value: number;
  prefix?: string;
  suffix?: string;
  decimals?: number;
  inView: boolean;
}) {
  const [display, setDisplay] = useState(0);

  useEffect(() => {
    if (!inView) return;

    const duration = 1600;
    const start = performance.now();

    const tick = (now: number) => {
      const progress = Math.min((now - start) / duration, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      setDisplay(value * eased);
      if (progress < 1) requestAnimationFrame(tick);
    };

    requestAnimationFrame(tick);
  }, [inView, value]);

  const formatted =
    decimals > 0 ? display.toFixed(decimals) : Math.round(display).toString();

  return (
    <span className="font-mono text-3xl font-medium tabular-nums text-white sm:text-4xl">
      {prefix}
      {formatted}
      {suffix}
    </span>
  );
}

export function StatsRow() {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { once: true, margin: "-60px" });

  return (
    <section
      ref={ref}
      className="relative border-y border-white/[0.06] bg-zinc-900/20 py-16 backdrop-blur-sm"
    >
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,rgba(16,185,129,0.06),transparent_70%)]" />

      <motion.div
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        viewport={{ once: true }}
        transition={{ duration: 0.8, ease: easeOut }}
        className="relative mx-auto grid max-w-7xl grid-cols-2 gap-8 px-6 lg:grid-cols-4"
      >
        {stats.map((stat, i) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 16 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ delay: i * 0.08, duration: 0.6, ease: easeOut }}
            className="text-center"
          >
            <AnimatedCounter
              value={stat.value}
              prefix={stat.prefix}
              suffix={stat.suffix}
              decimals={stat.decimals}
              inView={inView}
            />
            <p className="mt-2 font-mono text-[11px] uppercase tracking-wider text-zinc-500">
              {stat.label}
            </p>
          </motion.div>
        ))}
      </motion.div>
    </section>
  );
}
