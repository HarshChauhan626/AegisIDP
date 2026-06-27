"use client";

import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/services/api";
import type { WorkflowExecution } from "@/types";

export function useWorkflows(environmentId?: string) {
  return useQuery<WorkflowExecution[]>({
    queryKey: ["workflows", environmentId],
    queryFn: async () => {
      const params = environmentId
        ? `?environment_id=${environmentId}&limit=50`
        : "?limit=50";
      const response = await apiFetch<{
        data: WorkflowExecution[];
        meta: { total: number };
      }>(`/api/workflows${params}`);
      // apiFetch returns body.data, but paginated endpoints nest further
      return (response as any)?.data ?? response ?? [];
    },
  });
}

export function useWorkflow(id: string) {
  return useQuery<WorkflowExecution>({
    queryKey: ["workflow", id],
    queryFn: () => apiFetch<WorkflowExecution>(`/api/workflows/${id}`),
    enabled: !!id,
  });
}
