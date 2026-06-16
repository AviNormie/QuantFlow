"use client";

import { Navbar } from "@/components/layout/navbar";
import { AuroraBackground } from "@/components/landing/aurora-background";
import { CtaSection } from "@/components/landing/cta-section";
import { FeaturesBento } from "@/components/landing/features-bento";
import { Hero } from "@/components/landing/hero";
import { StatsRow } from "@/components/landing/stats-row";
import { TerminalPreview } from "@/components/landing/terminal-preview";

export default function HomePage() {
  return (
    <div className="min-h-screen overflow-x-hidden bg-zinc-950 text-zinc-100">
      <AuroraBackground />
      <Navbar />
      <Hero />
      <TerminalPreview />
      <StatsRow />
      <FeaturesBento />
      <CtaSection />
      <footer className="border-t border-white/[0.06] py-8 text-center text-xs text-zinc-600">
        © {new Date().getFullYear()} StockFlow. Market data for demonstration.
      </footer>
    </div>
  );
}
