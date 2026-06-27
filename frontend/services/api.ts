const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Token storage helpers
export const tokenStorage = {
  getAccessToken: () => {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("access_token");
  },
  setTokens: (accessToken: string, refreshToken: string) => {
    localStorage.setItem("access_token", accessToken);
    localStorage.setItem("refresh_token", refreshToken);
  },
  getRefreshToken: () => {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("refresh_token");
  },
  clear: () => {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
  },
};

// Typed API response envelope
export interface ApiResponse<T> {
  data: T;
  error: { code: string; message: string } | null;
  meta?: { total: number; limit: number; offset: number };
}

// Generic fetch wrapper with auth headers
export async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = tokenStorage.getAccessToken();

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  });

  // Handle token expiry — attempt refresh
  if (response.status === 401) {
    const refreshed = await tryRefreshToken();
    if (refreshed) {
      const newToken = tokenStorage.getAccessToken();
      const retryResponse = await fetch(`${API_URL}${path}`, {
        ...options,
        headers: {
          ...headers,
          Authorization: `Bearer ${newToken}`,
        },
      });
      return parseResponse<T>(retryResponse);
    }
    // Refresh failed — clear and redirect
    tokenStorage.clear();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
    throw new Error("Session expired");
  }

  return parseResponse<T>(response);
}

async function parseResponse<T>(response: Response): Promise<T> {
  const body: ApiResponse<T> = await response.json();
  if (!response.ok || body.error) {
    throw new Error(body.error?.message ?? `Request failed: ${response.status}`);
  }
  return body.data;
}

async function tryRefreshToken(): Promise<boolean> {
  const refreshToken = tokenStorage.getRefreshToken();
  if (!refreshToken) return false;

  try {
    const response = await fetch(`${API_URL}/api/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) return false;

    const body = await response.json();
    if (body.data?.access_token) {
      tokenStorage.setTokens(body.data.access_token, body.data.refresh_token);
      return true;
    }
    return false;
  } catch {
    return false;
  }
}
