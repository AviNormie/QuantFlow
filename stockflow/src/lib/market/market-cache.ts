/** Simple in-memory TTL cache for faster chart/quote reloads. */

const TTL_MS = 60_000;

type Entry<T> = { data: T; at: number };

const quotes = new Map<string, Entry<unknown>>();
const bars = new Map<string, Entry<unknown>>();

function get<T>(store: Map<string, Entry<unknown>>, key: string): T | null {
  const entry = store.get(key);
  if (!entry) return null;
  if (Date.now() - entry.at > TTL_MS) {
    store.delete(key);
    return null;
  }
  return entry.data as T;
}

function set(store: Map<string, Entry<unknown>>, key: string, data: unknown) {
  store.set(key, { data, at: Date.now() });
}

export function getCachedQuote<T>(symbol: string): T | null {
  return get<T>(quotes, symbol.toUpperCase());
}

export function setCachedQuote(symbol: string, data: unknown) {
  set(quotes, symbol.toUpperCase(), data);
}

export function getCachedBars<T>(key: string): T | null {
  return get<T>(bars, key);
}

export function setCachedBars(key: string, data: unknown) {
  set(bars, key, data);
}
