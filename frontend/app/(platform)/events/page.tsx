"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/services/api";
import type { Event } from "@/types";
import { Activity } from "lucide-react";

export default function EventsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["events"],
    queryFn: () => apiFetch<{ data: Event[] }>("/api/events?limit=50"),
    refetchInterval: 5000,
  });

  const events = (data as any)?.data ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Event Timeline</h2>
        <p className="text-gray-400 text-sm mt-1">Live stream of workflow events — auto-refreshes every 5s</p>
      </div>
      <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        {isLoading ? (
          <div className="h-48 animate-pulse" />
        ) : events.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20">
            <Activity className="w-10 h-10 text-gray-600 mb-3" />
            <p className="text-gray-400 text-sm">No events recorded yet</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-800">
            {events.map((event: Event) => (
              <div key={event.id} className="flex items-center gap-4 px-5 py-3.5">
                <span className="w-1.5 h-1.5 rounded-full bg-violet-400 flex-shrink-0" />
                <span className="text-xs text-gray-500 w-36 flex-shrink-0">
                  {new Date(event.occurred_at).toLocaleTimeString()}
                </span>
                <span className="text-sm font-medium text-white">{event.type}</span>
                <span className="text-xs text-gray-500 font-mono truncate">{event.workflow_id}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
