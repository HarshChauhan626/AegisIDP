"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "@/services/api";
import type { Environment } from "@/types";
import { Plus, Server, Circle, X, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { useRouter } from "next/navigation";

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
  const router = useRouter();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [envName, setEnvName] = useState("");
  const [capacity, setCapacity] = useState("small");

  const { data: environments = [], isLoading } = useQuery<Environment[]>({
    queryKey: ["environments"],
    queryFn: () => apiFetch<Environment[]>("/api/environments"),
  });

  const createMutation = useMutation({
    mutationFn: (name: string) =>
      apiFetch<{ environment: Environment; workflow_id: string }>("/api/environments", {
        method: "POST",
        body: JSON.stringify({ name, project_id: "default-project", config: { capacity } }),
      }),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["environments"] });
      setIsModalOpen(false);
      setEnvName("");
      if (data.workflow_id) {
        router.push(`/workflows/${data.workflow_id}`);
      }
    },
  });

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    if (envName.trim().length >= 2) {
      createMutation.mutate(envName.trim());
    }
  };

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
          onClick={() => setIsModalOpen(true)}
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
              onClick={() => router.push(`/environments/${env.id}`)}
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

      {/* Create Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
          <div className="bg-gray-900 border border-gray-800 rounded-xl p-6 w-full max-w-md shadow-2xl">
            <div className="flex items-center justify-between mb-5">
              <h3 className="text-lg font-semibold text-white">New Environment</h3>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-gray-500 hover:text-white transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <form onSubmit={handleCreate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">
                  Environment Name
                </label>
                <input
                  type="text"
                  value={envName}
                  onChange={(e) => setEnvName(e.target.value)}
                  placeholder="e.g. staging-env"
                  className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-white placeholder:text-gray-600 focus:outline-none focus:border-violet-500 transition-colors"
                  autoFocus
                  required
                  minLength={2}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-1">
                  Capacity
                </label>
                <select
                  value={capacity}
                  onChange={(e) => setCapacity(e.target.value)}
                  className="w-full bg-gray-950 border border-gray-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-violet-500 transition-colors"
                >
                  <option value="small">Small (Dev)</option>
                  <option value="medium">Medium (Staging)</option>
                  <option value="large">Large (Production)</option>
                </select>
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-400 hover:text-white transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={createMutation.isPending || envName.trim().length < 2}
                  className="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-500 disabled:opacity-50 disabled:cursor-not-allowed text-white text-sm font-medium rounded-lg transition-colors"
                >
                  {createMutation.isPending ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : null}
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
