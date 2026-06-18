const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

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
  return API_URL.replace(/\/$/, "");
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

export function getUserEmail() {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(USER_EMAIL_KEY);
}

export function isAuthenticated() {
  return Boolean(getAccessToken());
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
