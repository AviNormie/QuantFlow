"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { easeOut } from "./motion";

export function CtaSection() {
  return (
    <section className="relative overflow-hidden py-28 lg:py-36">
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-emerald-950/30 to-transparent" />
      <motion.div
        className="absolute inset-x-0 top-1/2 h-px -translate-y-1/2 bg-gradient-to-r from-transparent via-emerald-500/40 to-transparent"
        initial={{ scaleX: 0, opacity: 0 }}
        whileInView={{ scaleX: 1, opacity: 1 }}
        viewport={{ once: true }}
        transition={{ duration: 1.2, ease: easeOut }}
      />

      <motion.div
        initial={{ opacity: 0, y: 24 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.7, ease: easeOut }}
        className="relative mx-auto max-w-4xl px-6 text-center"
      >
        <p className="mb-4 font-mono text-[11px] uppercase tracking-[0.2em] text-emerald-500/70">
          Start analyzing
        </p>
        <h2 className="font-display text-4xl font-semibold tracking-tight text-white sm:text-5xl lg:text-6xl">
          Ready to open
          <span className="block bg-gradient-to-r from-emerald-300 via-teal-200 to-emerald-400 bg-clip-text text-transparent">
            the terminal?
          </span>
        </h2>
        <p className="mx-auto mt-6 max-w-lg text-zinc-400">
          Jump into live charts with WebSocket-powered price updates. No setup
          required — just open and watch the tape.
        </p>

        <Button
          size="lg"
          asChild
          className="group mt-10 shadow-lg shadow-emerald-500/20 transition-all hover:shadow-[0_0_40px_rgba(16,185,129,0.25)]"
        >
          <Link href="/charts">
            Open StockFlow Charts
            <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
          </Link>
        </Button>
      </motion.div>
    </section>
  );
}
