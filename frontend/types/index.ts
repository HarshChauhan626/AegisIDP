// Shared TypeScript types mirroring the backend models

export type WorkflowState =
  | "pending"
  | "queued"
  | "running"
  | "completed"
  | "failed"
  | "rolling_back"
  | "rolled_back"
  | "cancelled";

export type StepState =
  | "pending"
  | "running"
  | "succeeded"
  | "failed"
  | "skipped"
  | "rolled_back"
  | "rollback_failed";

export type EnvironmentStatus =
  | "pending"
  | "provisioning"
  | "ready"
  | "failed"
  | "deleting"
  | "deleted";

export type UserRole = "admin" | "developer" | "viewer";

export interface WorkflowStep {
  id: string;
  workflow_id: string;
  name: string;
  executor_key: string;
  state: StepState;
  attempt: number;
  input?: Record<string, unknown>;
  output?: Record<string, unknown>;
  error?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

export interface WorkflowExecution {
  id: string;
  environment_id: string;
  type: string;
  state: WorkflowState;
  trigger: "manual" | "scheduled";
  input?: Record<string, unknown>;
  error?: string;
  started_at?: string;
  completed_at?: string;
  created_by: string;
  created_at: string;
  steps?: WorkflowStep[];
}

export interface Environment {
  id: string;
  project_id: string;
  name: string;
  status: EnvironmentStatus;
  config?: Record<string, unknown>;
  created_by: string;
  created_at: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  active: boolean;
  created_at: string;
}

export interface Metrics {
  running_workflows: number;
  completed_workflows: number;
  failed_workflows: number;
  success_rate: number;
  queue_depth: number;
  active_workers: number;
  avg_provisioning_time_ms: number;
  total_retries: number;
}

export interface Event {
  id: string;
  type: string;
  workflow_id: string;
  step_id?: string;
  payload?: Record<string, unknown>;
  occurred_at: string;
}

export interface AuditLog {
  id: string;
  user_id: string;
  action: string;
  resource_type: string;
  resource_id: string;
  metadata?: Record<string, unknown>;
  occurred_at: string;
}
