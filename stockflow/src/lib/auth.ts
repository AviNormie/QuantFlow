import { getPublicApiUrl } from "@/lib/env";

const ACCESS_TOKEN_KEY = "stockflow_access_token";
const REFRESH_TOKEN_KEY = "stockflow_refresh_token";
const USER_EMAIL_KEY = "stockflow_user_email";

export type AuthUser = {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
};

export type AuthResponse = {
  user: AuthUser;
  access_token: string;
  refresh_token: string;
};

export function getApiUrl() {
  return getPublicApiUrl();
}

export function saveAuth(data: AuthResponse) {
  localStorage.setItem(ACCESS_TOKEN_KEY, data.access_token);
  localStorage.setItem(REFRESH_TOKEN_KEY, data.refresh_token);
  localStorage.setItem(USER_EMAIL_KEY, data.user.email);
}

export function clearAuth() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_EMAIL_KEY);
}

export function getAccessToken() {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getRefreshToken() {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function getUserEmail() {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(USER_EMAIL_KEY);
}

export function isAuthenticated() {
  return Boolean(getAccessToken());
}

async function authFetch(url: string, options: RequestInit = {}) {
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

export async function refreshTokens() {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  const res = await fetch(`${getPublicApiUrl()}/api/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!res.ok) {
    clearAuth();
    return false;
  }

  const data = (await res.json()) as AuthResponse;
  saveAuth(data);
  return true;
}

export async function logout() {
  const refreshToken = getRefreshToken();
  if (refreshToken) {
    await fetch(`${getApiUrl()}/api/auth/logout`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
  }
  clearAuth();
}

export async function fetchMe() {
  const res = await authFetch(`${getApiUrl()}/api/auth/me`);
  if (!res.ok) {
    throw new Error("Failed to load profile");
  }
  const data = (await res.json()) as { user: AuthUser };
  return data.user;
}

export async function register(email: string, password: string) {
  const res = await fetch(`${getApiUrl()}/api/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });

  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error ?? "Registration failed");
  }

  return data as AuthResponse;
}

export async function signup(email: string, password: string) {
  const res = await fetch(`${getApiUrl()}/api/auth/signup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });

  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error ?? "Registration failed");
  }

  return data as AuthResponse;
}

export async function login(email: string, password: string) {
  const res = await fetch(`${getApiUrl()}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });

  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error ?? "Login failed");
  }

  return data as AuthResponse;
}
