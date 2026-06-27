"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { getCurrentUser, isAuthenticated, login, logout } from "@/services/auth";

interface AuthUser {
  id: string;
  role: string;
}

export function useAuth() {
  const router = useRouter();
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const currentUser = getCurrentUser();
    setUser(currentUser);
    setLoading(false);
  }, []);

  const handleLogin = useCallback(
    async (email: string, password: string) => {
      const userData = await login(email, password);
      setUser({ id: userData.id, role: userData.role });
      router.push("/dashboard");
    },
    [router]
  );

  const handleLogout = useCallback(() => {
    logout();
    setUser(null);
    router.push("/login");
  }, [router]);

  const hasPermission = useCallback(
    (permission: string) => {
      if (!user) return false;
      const permissions: Record<string, string[]> = {
        admin: [
          "environment:create",
          "environment:delete",
          "workflow:view",
          "workflow:cancel",
          "workflow:retry",
          "template:manage",
          "schedule:manage",
          "metrics:view",
          "audit:view",
          "user:manage",
        ],
        developer: [
          "environment:create",
          "environment:delete",
          "workflow:view",
          "workflow:cancel",
          "workflow:retry",
          "template:manage",
          "schedule:manage",
          "metrics:view",
        ],
        viewer: ["workflow:view", "metrics:view"],
      };
      return permissions[user.role]?.includes(permission) ?? false;
    },
    [user]
  );

  return {
    user,
    loading,
    isAuthenticated: isAuthenticated(),
    login: handleLogin,
    logout: handleLogout,
    hasPermission,
  };
}
