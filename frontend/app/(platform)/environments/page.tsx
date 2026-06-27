"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/services/api";
import type { Environment } from "@/types";
import { Plus, Server, Circle } from "lucide-react";
import { cn } from "@/lib/utils";

const statusColors: Record<string, string> = {
  pending: "text-gray-400",
  provisioning: "text-blue-400",
  ready: "text-emerald-400",
  failed: "text-red-400",
  deleting: "text-amber-400",
  deleted: "text-gray-600",
};

const statusBg: Record<string, string> = {
  pending: "bg-gray-400/10",
  provisioning: "bg-blue-400/10",
  ready: "bg-emerald-400/10",
  failed: "bg-red-400/10",
  deleting: "bg-amber-400/10",
  deleted: "bg-gray-600/10",
};

export default function EnvironmentsPage() {
  const { data: environments = [], isLoading } = useQuery<Environment[]>({
    queryKey: ["environments"],
    queryFn: () => apiFetch<Environment[]>("/api/environments"),
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Environments</h2>
          <p className="text-gray-400 text-sm mt-1">Manage provisioned application environments</p>
        </div>
        <button
          id="create-environment-btn"
          className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-sm font-medium rounded-lg transition-all shadow-lg shadow-violet-500/20"
        >
          <Plus className="w-4 h-4" />
          New Environment
        </button>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-36 bg-gray-900 border border-gray-800 rounded-xl animate-pulse" />
          ))}
        </div>
      ) : environments.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="flex items-center justify-center w-16 h-16 rounded-2xl bg-gray-900 border border-gray-800 mb-4">
            <Server className="w-8 h-8 text-gray-600" />
          </div>
          <h3 className="text-lg font-medium text-white mb-1">No environments yet</h3>
          <p className="text-gray-400 text-sm">
            Create your first environment to start provisioning infrastructure.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {environments.map((env) => (
            <div
              key={env.id}
              className="bg-gray-900 border border-gray-800 rounded-xl p-5 hover:border-gray-700 transition-colors cursor-pointer"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-indigo-600/20">
                  <Server className="w-5 h-5 text-indigo-400" />
                </div>
                <span
                  className={cn(
                    "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium",
                    statusBg[env.status],
                    statusColors[env.status]
                  )}
                >
                  <Circle className="w-1.5 h-1.5 fill-current" />
                  {env.status}
                </span>
              </div>
              <h3 className="font-semibold text-white">{env.name}</h3>
              <p className="text-xs text-gray-500 mt-1 font-mono">{env.id.slice(0, 8)}…</p>
              <p className="text-xs text-gray-600 mt-3">
                Created {new Date(env.created_at).toLocaleDateString()}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
