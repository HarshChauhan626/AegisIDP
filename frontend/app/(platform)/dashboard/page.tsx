"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/services/api";
import type { Metrics } from "@/types";
import {
  Activity,
  CheckCircle,
  XCircle,
  Clock,
  Layers,
  Users,
  RefreshCw,
  TrendingUp,
} from "lucide-react";

interface MetricCardProps {
  label: string;
  value: string | number;
  icon: React.ElementType;
  color: string;
  sub?: string;
}

function MetricCard({ label, value, icon: Icon, color, sub }: MetricCardProps) {
  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl p-5 flex items-start gap-4 hover:border-gray-700 transition-colors">
      <div className={`flex items-center justify-center w-10 h-10 rounded-lg ${color}`}>
        <Icon className="w-5 h-5 text-white" />
      </div>
      <div className="min-w-0">
        <p className="text-2xl font-bold text-white">{value}</p>
        <p className="text-sm text-gray-400 mt-0.5">{label}</p>
        {sub && <p className="text-xs text-gray-600 mt-1">{sub}</p>}
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const { data: metrics, isLoading } = useQuery<Metrics>({
    queryKey: ["metrics"],
    queryFn: () => apiFetch<Metrics>("/api/metrics"),
    refetchInterval: 5000,
  });

  const cards = [
    {
      label: "Running Workflows",
      value: metrics?.running_workflows ?? 0,
      icon: Activity,
      color: "bg-blue-600",
    },
    {
      label: "Completed",
      value: metrics?.completed_workflows ?? 0,
      icon: CheckCircle,
      color: "bg-emerald-600",
    },
    {
      label: "Failed",
      value: metrics?.failed_workflows ?? 0,
      icon: XCircle,
      color: "bg-red-600",
    },
    {
      label: "Success Rate",
      value: metrics ? `${(metrics.success_rate * 100).toFixed(1)}%` : "—",
      icon: TrendingUp,
      color: "bg-violet-600",
    },
    {
      label: "Queue Depth",
      value: metrics?.queue_depth ?? 0,
      icon: Layers,
      color: "bg-amber-600",
    },
    {
      label: "Active Workers",
      value: metrics?.active_workers ?? 0,
      icon: Users,
      color: "bg-indigo-600",
    },
    {
      label: "Avg Provisioning",
      value: metrics ? `${metrics.avg_provisioning_time_ms}ms` : "—",
      icon: Clock,
      color: "bg-pink-600",
    },
    {
      label: "Total Retries",
      value: metrics?.total_retries ?? 0,
      icon: RefreshCw,
      color: "bg-orange-600",
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h2 className="text-2xl font-bold text-white">Dashboard</h2>
        <p className="text-gray-400 text-sm mt-1">
          Real-time operational metrics — auto-refreshes every 5 seconds
        </p>
      </div>

      {/* Metric grid */}
      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="h-24 bg-gray-900 border border-gray-800 rounded-xl animate-pulse" />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {cards.map((card) => (
            <MetricCard key={card.label} {...card} />
          ))}
        </div>
      )}

      {/* Status notice for Phase 1 */}
      <div className="bg-violet-900/20 border border-violet-800/40 rounded-xl p-4">
        <p className="text-sm text-violet-300">
          <span className="font-semibold">Phase 1 Foundation</span> — Workflow engine, real-time SSE streaming, and
          full metrics aggregation will be active in Phase 2+. Metrics reflect live system state once workflows are
          executing.
        </p>
      </div>
    </div>
  );
}
