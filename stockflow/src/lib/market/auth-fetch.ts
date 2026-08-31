import { getPublicApiUrl } from "@/lib/env";
import { getAccessToken, getRefreshToken, refreshTokens } from "@/lib/auth";

/** Authenticated fetch for market API routes with refresh-token retry. */
export async function marketAuthFetch(url: string, options: RequestInit = {}) {
  const headers = new Headers(options.headers);
  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  let res = await fetch(url, { ...options, headers });
  if (res.status === 401 && getRefreshToken()) {
    const refreshed = await refreshTokens();
    if (refreshed) {
      headers.set("Authorization", `Bearer ${getAccessToken()}`);
      res = await fetch(url, { ...options, headers });
    }
  }

  return res;
}

export function getMarketApiUrl(path: string) {
  const base = getPublicApiUrl().replace(/\/$/, "");
  return `${base}${path.startsWith("/") ? path : `/${path}`}`;
}
