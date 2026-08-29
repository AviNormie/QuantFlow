"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { searchSymbols, type SymbolInfo } from "@/lib/market/api";
import { cn } from "@/lib/utils";

type SymbolSearchProps = {
  value: string;
  onChange: (value: string) => void;
  onSelect: (symbol: string) => void;
  className?: string;
};

export function SymbolSearch({
  value,
  onChange,
  onSelect,
  className,
}: SymbolSearchProps) {
  const [open, setOpen] = useState(false);
  const [results, setResults] = useState<SymbolInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const runSearch = useCallback(async (query: string) => {
    setLoading(true);
    try {
      const items = await searchSymbols(query, 30);
      setResults(items);
      setActiveIndex(0);
    } catch {
      setResults([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!open) return;

    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    debounceRef.current = setTimeout(() => {
      runSearch(value.trim());
    }, 250);

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, [value, open, runSearch]);

  useEffect(() => {
    const onPointerDown = (event: MouseEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", onPointerDown);
    return () => document.removeEventListener("mousedown", onPointerDown);
  }, []);

  const pick = (symbol: string) => {
    onChange(symbol);
    onSelect(symbol);
    setOpen(false);
  };

  const onKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (!open || results.length === 0) {
      if (event.key === "Enter") {
        event.preventDefault();
        const next = value.trim().toUpperCase();
        if (next) pick(next);
      }
      return;
    }

    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActiveIndex((i) => (i + 1) % results.length);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActiveIndex((i) => (i - 1 + results.length) % results.length);
    } else if (event.key === "Enter") {
      event.preventDefault();
      pick(results[activeIndex].symbol);
    } else if (event.key === "Escape") {
      setOpen(false);
    }
  };

  return (
    <div ref={containerRef} className={cn("relative w-full", className)}>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-500" />
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value.toUpperCase())}
          onFocus={() => setOpen(true)}
          onKeyDown={onKeyDown}
          placeholder="Search symbol (e.g. AAPL, Amazon)..."
          className="pl-9"
          autoComplete="off"
          aria-expanded={open}
          aria-autocomplete="list"
        />
      </div>

      {open && (
        <div
          className="absolute z-50 mt-2 w-full overflow-hidden rounded-lg border border-white/10 bg-zinc-900 shadow-xl"
          role="listbox"
        >
          <div className="border-b border-white/5 px-3 py-2 text-xs text-zinc-500">
            {loading
              ? "Searching..."
              : value.trim()
                ? `${results.length} match${results.length === 1 ? "" : "es"}`
                : "Popular US stocks — type to search all symbols"}
          </div>
          <ul className="max-h-72 overflow-y-auto py-1">
            {results.length === 0 && !loading ? (
              <li className="px-3 py-6 text-center text-sm text-zinc-500">
                No symbols found
              </li>
            ) : (
              results.map((item, index) => (
                <li key={item.symbol}>
                  <button
                    type="button"
                    role="option"
                    aria-selected={index === activeIndex}
                    className={cn(
                      "flex w-full items-start gap-3 px-3 py-2.5 text-left text-sm transition",
                      index === activeIndex
                        ? "bg-emerald-500/10 text-emerald-200"
                        : "text-zinc-300 hover:bg-white/5",
                    )}
                    onMouseEnter={() => setActiveIndex(index)}
                    onClick={() => pick(item.symbol)}
                  >
                    <span className="font-mono font-semibold text-white">
                      {item.symbol}
                    </span>
                    <span className="truncate text-xs text-zinc-500">
                      {item.description || item.name}
                    </span>
                  </button>
                </li>
              ))
            )}
          </ul>
        </div>
      )}
    </div>
  );
}
