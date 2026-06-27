"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/services/api";
import type { User } from "@/types";
import { Users, Shield, Code2, Eye } from "lucide-react";
import { cn } from "@/lib/utils";

const roleIcons: Record<string, React.ElementType> = {
  admin: Shield,
  developer: Code2,
  viewer: Eye,
};

const roleColors: Record<string, string> = {
  admin: "text-violet-400 bg-violet-400/10",
  developer: "text-blue-400 bg-blue-400/10",
  viewer: "text-gray-400 bg-gray-400/10",
};

export default function UsersPage() {
  const { data: users = [], isLoading } = useQuery<User[]>({
    queryKey: ["users"],
    queryFn: () => apiFetch<User[]>("/api/users"),
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">User Management</h2>
          <p className="text-gray-400 text-sm mt-1">Manage platform users and roles</p>
        </div>
        <button
          id="create-user-btn"
          className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-sm font-medium rounded-lg transition-all"
        >
          <Users className="w-4 h-4" />
          Add User
        </button>
      </div>

      <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        {isLoading ? (
          <div className="h-48 animate-pulse" />
        ) : users.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20">
            <Users className="w-10 h-10 text-gray-600 mb-3" />
            <p className="text-gray-400 text-sm">No users found</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-800">
            {users.map((user) => {
              const RoleIcon = roleIcons[user.role] ?? Eye;
              return (
                <div key={user.id} className="flex items-center justify-between px-5 py-4">
                  <div className="flex items-center gap-3">
                    <div className="flex items-center justify-center w-9 h-9 rounded-full bg-gradient-to-br from-violet-500 to-indigo-600">
                      <span className="text-sm font-bold text-white">
                        {user.name.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <p className="font-medium text-white">{user.name}</p>
                      <p className="text-xs text-gray-500">{user.email}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span
                      className={cn(
                        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium",
                        roleColors[user.role]
                      )}
                    >
                      <RoleIcon className="w-3 h-3" />
                      {user.role}
                    </span>
                    <span
                      className={cn(
                        "text-xs px-2 py-0.5 rounded-full",
                        user.active
                          ? "text-emerald-400 bg-emerald-400/10"
                          : "text-gray-500 bg-gray-700"
                      )}
                    >
                      {user.active ? "Active" : "Inactive"}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
