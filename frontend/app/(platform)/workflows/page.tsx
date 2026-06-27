"use client";

import { useWorkflows } from "@/hooks/useWorkflows";
import type { WorkflowExecution } from "@/types";
import { GitBranch, Circle, RefreshCw } from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";

const stateColors: Record<string, string> = {
  pending: "text-gray-400",
  queued: "text-blue-400",
  running: "text-indigo-400",
  completed: "text-emerald-400",
  failed: "text-red-400",
  rolling_back: "text-amber-400",
  rolled_back: "text-orange-400",
  cancelled: "text-gray-500",
};

const stateBg: Record<string, string> = {
  pending: "bg-gray-400/10",
  queued: "bg-blue-400/10",
  running: "bg-indigo-400/10",
  completed: "bg-emerald-400/10",
  failed: "bg-red-400/10",
  rolling_back: "bg-amber-400/10",
  rolled_back: "bg-orange-400/10",
  cancelled: "bg-gray-500/10",
};

function WorkflowRow({ wf }: { wf: WorkflowExecution }) {
  return (
    <Link
      href={`/workflows/${wf.id}`}
      className="flex items-center justify-between px-5 py-4 hover:bg-gray-800/50 transition-colors cursor-pointer"
    >
      <div className="flex items-center gap-4 min-w-0">
        <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-violet-600/20 flex-shrink-0">
          <GitBranch className="w-5 h-5 text-violet-400" />
        </div>
        <div className="min-w-0">
          <p className="font-medium text-white truncate">{wf.type}</p>
          <p className="text-xs text-gray-500 font-mono">{wf.id.slice(0, 12)}…</p>
        </div>
      </div>
      <div className="flex items-center gap-6 flex-shrink-0 ml-4">
        <span
          className={cn(
            "inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium",
            stateBg[wf.state],
            stateColors[wf.state]
          )}
        >
          <Circle
            className={cn(
              "w-1.5 h-1.5 fill-current",
              wf.state === "running" && "animate-pulse"
            )}
          />
          {wf.state}
        </span>
        <span className="text-xs text-gray-500 hidden md:block">
          {wf.trigger === "scheduled" ? "⏰ scheduled" : "👤 manual"}
        </span>
        <span className="text-xs text-gray-600">
          {new Date(wf.created_at).toLocaleString()}
        </span>
      </div>
    </Link>
  );
}

export default function WorkflowsPage() {
  const { data: workflows = [], isLoading, refetch } = useWorkflows();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Workflows</h2>
          <p className="text-gray-400 text-sm mt-1">Execution history and live status</p>
        </div>
        <button
          id="refresh-workflows-btn"
          onClick={() => refetch()}
          className="flex items-center gap-2 px-3 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm rounded-lg transition-colors"
        >
          <RefreshCw className="w-4 h-4" />
          Refresh
        </button>
      </div>

      {/* Table */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        {isLoading ? (
          <div className="space-y-px">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="h-16 bg-gray-800 animate-pulse" style={{ opacity: 1 - i * 0.15 }} />
            ))}
          </div>
        ) : workflows.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <GitBranch className="w-10 h-10 text-gray-600 mb-4" />
            <h3 className="text-lg font-medium text-white mb-1">No workflows yet</h3>
            <p className="text-gray-400 text-sm">
              Create an environment to trigger your first workflow.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-800">
            {workflows.map((wf) => (
              <WorkflowRow key={wf.id} wf={wf} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
