"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/services/api";
import type { AuditLog } from "@/types";
import { ClipboardList } from "lucide-react";

export default function AuditPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["audit"],
    queryFn: () => apiFetch<{ data: AuditLog[] }>("/api/audit?limit=50"),
  });

  const logs = (data as any)?.data ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Audit Log</h2>
        <p className="text-gray-400 text-sm mt-1">Every user-initiated action recorded for compliance</p>
      </div>
      <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        {isLoading ? (
          <div className="h-48 animate-pulse" />
        ) : logs.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20">
            <ClipboardList className="w-10 h-10 text-gray-600 mb-3" />
            <p className="text-gray-400 text-sm">No audit records yet</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-800">
            {logs.map((log: AuditLog) => (
              <div key={log.id} className="flex items-center gap-4 px-5 py-3.5">
                <span className="text-xs text-gray-500 w-36 flex-shrink-0">
                  {new Date(log.occurred_at).toLocaleString()}
                </span>
                <span className="text-sm text-violet-400 font-mono w-48 flex-shrink-0">{log.action}</span>
                <span className="text-sm text-white">{log.resource_type}</span>
                <span className="text-xs text-gray-500 font-mono truncate">{log.resource_id}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
