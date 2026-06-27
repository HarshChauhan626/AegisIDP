"use client";

import { use } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/services/api";
import type { Environment, WorkflowExecution } from "@/types";
import {
  ArrowLeft,
  Server,
  Circle,
  Clock,
  CheckCircle2,
  XCircle,
  Loader2,
  RotateCcw,
  Ban,
  Activity
} from "lucide-react";
import Link from "next/link";
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

const wfStateIcon = (state: string) => {
  switch (state) {
    case "completed": return <CheckCircle2 className="w-4 h-4 text-emerald-400" />;
    case "failed":
    case "rollback_failed": return <XCircle className="w-4 h-4 text-red-400" />;
    case "running": return <Loader2 className="w-4 h-4 text-blue-400 animate-spin" />;
    case "rolled_back": return <RotateCcw className="w-4 h-4 text-amber-400" />;
    case "cancelled": return <Ban className="w-4 h-4 text-gray-500" />;
    default: return <Circle className="w-4 h-4 text-gray-400" />;
  }
};

interface WorkflowListResponse {
  data: WorkflowExecution[];
  total: number;
  limit: number;
  offset: number;
}

export default function EnvironmentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);

  const { data: env, isLoading: isEnvLoading } = useQuery<Environment>({
    queryKey: ["environment", id],
    queryFn: () => apiFetch<Environment>(`/api/environments/${id}`),
  });

  const { data: workflowsRes, isLoading: isWfLoading } = useQuery<WorkflowListResponse>({
    queryKey: ["workflows", "env", id],
    queryFn: () => apiFetch<WorkflowListResponse>(`/api/workflows?environment_id=${id}`),
    refetchInterval: 3000, // Poll every 3s to keep workflow state somewhat fresh
  });

  const workflows = workflowsRes?.data || [];

  if (isEnvLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 text-violet-400 animate-spin" />
      </div>
    );
  }

  if (!env) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center">
        <XCircle className="w-12 h-12 text-red-400 mb-4" />
        <h3 className="text-lg font-medium text-white">Environment not found</h3>
        <Link href="/environments" className="text-violet-400 text-sm mt-2 hover:underline">
          ← Back to environments
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Back */}
      <Link
        href="/environments"
        className="inline-flex items-center gap-2 text-sm text-gray-400 hover:text-white transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        Environments
      </Link>

      {/* Header */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl p-6 flex flex-col md:flex-row md:items-start justify-between gap-4">
        <div className="flex items-start gap-4">
          <div className="flex items-center justify-center w-12 h-12 rounded-xl bg-indigo-600/20 shrink-0">
            <Server className="w-6 h-6 text-indigo-400" />
          </div>
          <div>
            <h2 className="text-2xl font-bold text-white">{env.name}</h2>
            <p className="text-sm text-gray-500 font-mono mt-1">{env.id}</p>
            <div className="flex items-center gap-4 mt-3">
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
              <span className="text-xs text-gray-500 flex items-center gap-1">
                <Clock className="w-3.5 h-3.5" />
                {new Date(env.created_at).toLocaleString()}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Workflows List */}
      <div>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Activity className="w-5 h-5 text-gray-400" />
          Execution History
        </h3>
        
        {isWfLoading ? (
          <div className="space-y-3">
            {[1, 2].map((i) => (
              <div key={i} className="h-20 bg-gray-900/50 border border-gray-800 rounded-xl animate-pulse" />
            ))}
          </div>
        ) : workflows.length === 0 ? (
          <div className="bg-gray-900 border border-gray-800 rounded-xl p-8 text-center">
            <p className="text-gray-400 text-sm">No workflow executions found for this environment.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {workflows.map((wf) => (
              <Link
                key={wf.id}
                href={`/workflows/${wf.id}`}
                className="flex items-center justify-between p-4 bg-gray-900 border border-gray-800 rounded-xl hover:border-gray-700 transition-all group"
              >
                <div className="flex items-center gap-4">
                  <div className="shrink-0 p-2 bg-gray-800 rounded-lg group-hover:bg-gray-750 transition-colors">
                    {wfStateIcon(wf.state)}
                  </div>
                  <div>
                    <h4 className="text-sm font-medium text-white capitalize">
                      {wf.type.replace(/-/g, " ")}
                    </h4>
                    <p className="text-xs text-gray-500 font-mono mt-0.5">{wf.id.slice(0, 8)}…</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-xs font-medium text-gray-300 capitalize">
                    {wf.state.replace(/_/g, " ")}
                  </p>
                  <p className="text-xs text-gray-500 mt-0.5">
                    {wf.started_at ? new Date(wf.started_at).toLocaleTimeString() : "Pending"}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
