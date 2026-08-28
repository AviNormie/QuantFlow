/**
 * Runtime URL helpers for API and WebSocket clients.
 * Production: set NEXT_PUBLIC_* in Vercel/Render dashboard (see DEPLOYMENT.md).
 */

const DEFAULT_API = "http://localhost:8080";
const DEFAULT_WS = "ws://localhost:8083";

export function getPublicApiUrl(): string {
  const raw = process.env.NEXT_PUBLIC_API_URL ?? DEFAULT_API;
  return raw.replace(/\/$/, "");
}

/** Converts http(s) Render URLs to ws(s) and ensures a WebSocket scheme. */
export function getPublicWebSocketUrl(): string {
  const raw =
    process.env.NEXT_PUBLIC_WEBSOCKET_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    DEFAULT_WS;

  const trimmed = raw.replace(/\/$/, "");

  if (trimmed.startsWith("ws://") || trimmed.startsWith("wss://")) {
    return trimmed;
  }
  if (trimmed.startsWith("https://")) {
    return trimmed.replace(/^https:/, "wss:");
  }
  if (trimmed.startsWith("http://")) {
    return trimmed.replace(/^http:/, "ws:");
  }
  return `wss://${trimmed}`;
}
