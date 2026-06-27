import { apiFetch, tokenStorage } from "./api";

export interface User {
  id: string;
  email: string;
  name: string;
  role: "admin" | "developer" | "viewer";
  active: boolean;
  created_at: string;
}

export interface LoginResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

// Login — returns user + tokens and persists tokens to localStorage
export async function login(email: string, password: string): Promise<User> {
  const data = await apiFetch<LoginResponse>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });

  tokenStorage.setTokens(data.access_token, data.refresh_token);
  return data.user;
}

// Logout — clears tokens
export function logout(): void {
  tokenStorage.clear();
}

// getCurrentUser — decodes user info from the stored token (no extra network call)
export function getCurrentUser(): { id: string; role: string } | null {
  const token = tokenStorage.getAccessToken();
  if (!token) return null;

  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    if (payload.exp * 1000 < Date.now()) {
      tokenStorage.clear();
      return null;
    }
    return { id: payload.user_id, role: payload.role };
  } catch {
    return null;
  }
}

// isAuthenticated — checks token presence and validity
export function isAuthenticated(): boolean {
  return getCurrentUser() !== null;
}
