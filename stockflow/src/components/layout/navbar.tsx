import Link from "next/link";
import { Button } from "@/components/ui/button";
import { TrendingUp } from "lucide-react";

export function Navbar() {
  return (
    <header className="fixed inset-x-0 top-0 z-50 border-b border-white/[0.06] bg-zinc-950/70 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
        <Link href="/" className="flex items-center gap-2.5 group">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-500/10 ring-1 ring-emerald-500/30 transition group-hover:bg-emerald-500/20">
            <TrendingUp className="h-4 w-4 text-emerald-400" />
          </div>
          <span className="font-display text-lg font-semibold tracking-tight text-white">
            StockFlow
          </span>
        </Link>

        <nav className="hidden items-center gap-8 md:flex">
          <Link
            href="/#features"
            className="text-sm text-zinc-400 transition hover:text-zinc-100"
          >
            Features
          </Link>
          <Link
            href="/charts"
            className="text-sm text-zinc-400 transition hover:text-zinc-100"
          >
            Charts
          </Link>
        </nav>

        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" asChild className="hidden sm:inline-flex">
            <Link href="/charts">View markets</Link>
          </Button>
          <Button size="sm" asChild>
            <Link href="/charts">Open terminal</Link>
          </Button>
        </div>
      </div>
    </header>
  );
}
