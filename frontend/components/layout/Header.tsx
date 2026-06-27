"use client";

import { Bell, LogOut, User } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";

interface HeaderProps {
  title: string;
}

export function Header({ title }: HeaderProps) {
  const { user, logout } = useAuth();

  return (
    <header className="flex items-center justify-between h-16 px-6 border-b border-gray-800 bg-gray-950">
      <h1 className="text-lg font-semibold text-white">{title}</h1>

      <div className="flex items-center gap-4">
        {/* Notification bell */}
        <button
          id="notifications-btn"
          className="relative p-2 rounded-lg text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
        >
          <Bell className="w-5 h-5" />
        </button>

        {/* User info */}
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gradient-to-br from-violet-500 to-indigo-600">
            <User className="w-4 h-4 text-white" />
          </div>
          <div className="hidden md:block">
            <p className="text-sm font-medium text-white">{user?.id ?? "User"}</p>
            <p className="text-xs text-gray-400 capitalize">{user?.role ?? "—"}</p>
          </div>
        </div>

        {/* Logout */}
        <button
          id="logout-btn"
          onClick={logout}
          className="p-2 rounded-lg text-gray-400 hover:text-red-400 hover:bg-red-400/10 transition-colors"
          title="Logout"
        >
          <LogOut className="w-5 h-5" />
        </button>
      </div>
    </header>
  );
}
