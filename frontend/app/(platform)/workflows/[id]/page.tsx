"use client";

import { use } from "react";
import { useWorkflow } from "@/hooks/useWorkflows";
import type { WorkflowStep } from "@/types";
import {
  CheckCircle2,
  XCircle,
  Circle,
  Loader2,
  ArrowLeft,
  RotateCcw,
  Ban,
} from "lucide-react";
import Link from "next/link";
import { cn } from "@/lib/utils";

function StepIcon({ state }: { state: WorkflowStep["state"] }) {
  if (state === "succeeded") return <CheckCircle2 className="w-5 h-5 text-emerald-400" />;
  if (state === "failed" || state === "rollback_failed") return <XCircle className="w-5 h-5 text-red-400" />;
  if (state === "running") return <Loader2 className="w-5 h-5 text-indigo-400 animate-spin" />;
  if (state === "rolled_back") return <RotateCcw className="w-5 h-5 text-orange-400" />;
  return <Circle className="w-5 h-5 text-gray-600" />;
}

const stepStateLabel: Record<string, string> = {
  pending: "Pending",
  running: "Running",
  succeeded: "Succeeded",
  failed: "Failed",
  skipped: "Skipped",
  rolled_back: "Rolled Back",
  rollback_failed: "Rollback Failed",
};

export default function WorkflowDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data: workflow, isLoading, error } = useWorkflow(id);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 text-violet-400 animate-spin" />
      </div>
    );
  }

  if (error || !workflow) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center">
        <XCircle className="w-12 h-12 text-red-400 mb-4" />
        <h3 className="text-lg font-medium text-white">Workflow not found</h3>
        <Link href="/workflows" className="text-violet-400 text-sm mt-2 hover:underline">
          ← Back to workflows
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-3xl">
      {/* Back */}
      <Link
        href="/workflows"
        className="inline-flex items-center gap-2 text-sm text-gray-400 hover:text-white transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        Workflows
      </Link>

      {/* Header */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl p-6">
        <div className="flex items-start justify-between">
          <div>
            <h2 className="text-xl font-bold text-white capitalize">{workflow.type.replace(/-/g, " ")}</h2>
            <p className="text-xs text-gray-500 font-mono mt-1">{workflow.id}</p>
          </div>
          <div className="flex items-center gap-2">
            {(workflow.state === "failed" || workflow.state === "rolled_back") && (
              <button
                id="retry-workflow-btn"
                className="flex items-center gap-1.5 px-3 py-1.5 bg-amber-600/20 hover:bg-amber-600/30 text-amber-400 text-sm rounded-lg transition-colors border border-amber-600/30"
              >
                <RotateCcw className="w-3.5 h-3.5" />
                Retry
              </button>
            )}
            {(workflow.state === "running" || workflow.state === "queued") && (
              <button
                id="cancel-workflow-btn"
                className="flex items-center gap-1.5 px-3 py-1.5 bg-red-600/20 hover:bg-red-600/30 text-red-400 text-sm rounded-lg transition-colors border border-red-600/30"
              >
                <Ban className="w-3.5 h-3.5" />
                Cancel
              </button>
            )}
          </div>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-5 pt-5 border-t border-gray-800">
          <div>
            <p className="text-xs text-gray-500">State</p>
            <p className="text-sm font-medium text-white capitalize mt-0.5">{workflow.state.replace(/_/g, " ")}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500">Trigger</p>
            <p className="text-sm font-medium text-white capitalize mt-0.5">{workflow.trigger}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500">Started</p>
            <p className="text-sm font-medium text-white mt-0.5">
              {workflow.started_at ? new Date(workflow.started_at).toLocaleTimeString() : "—"}
            </p>
          </div>
          <div>
            <p className="text-xs text-gray-500">Completed</p>
            <p className="text-sm font-medium text-white mt-0.5">
              {workflow.completed_at ? new Date(workflow.completed_at).toLocaleTimeString() : "—"}
            </p>
          </div>
        </div>
      </div>

      {/* Steps timeline */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl p-6">
        <h3 className="text-base font-semibold text-white mb-5">Execution Steps</h3>
        {!workflow.steps?.length ? (
          <p className="text-gray-500 text-sm">No steps recorded yet.</p>
        ) : (
          <div className="space-y-0">
            {workflow.steps.map((step, i) => (
              <div key={step.id} className="flex gap-4">
                {/* Timeline line */}
                <div className="flex flex-col items-center">
                  <StepIcon state={step.state} />
                  {i < (workflow.steps?.length ?? 0) - 1 && (
                    <div className="w-px flex-1 bg-gray-800 my-1" />
                  )}
                </div>

                <div
                  className={cn(
                    "flex-1 pb-4",
                    i < (workflow.steps?.length ?? 0) - 1 && "border-b border-gray-800/50"
                  )}
                >
                  <div className="flex items-center justify-between">
                    <p className="font-medium text-white text-sm">{step.name}</p>
                    <span className="text-xs text-gray-500">{stepStateLabel[step.state]}</span>
                  </div>
                  <p className="text-xs text-gray-600 font-mono mt-0.5">{step.executor_key}</p>
                  {step.error && (
                    <p className="text-xs text-red-400 mt-1 bg-red-400/10 px-2 py-1 rounded">
                      {step.error}
                    </p>
                  )}
                  {step.attempt > 1 && (
                    <p className="text-xs text-amber-400 mt-1">Attempt #{step.attempt}</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {workflow.error && (
        <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4">
          <p className="text-sm font-medium text-red-400 mb-1">Workflow Error</p>
          <p className="text-sm text-red-300">{workflow.error}</p>
        </div>
      )}
    </div>
  );
}
