# Platform Orchestrator — System Design

## 1. Overview

Platform Orchestrator is a local-first Internal Developer Platform (IDP) that executes multi-step environment provisioning workflows asynchronously. Every user action enters a job queue, gets picked up by a worker pool, and passes through a workflow engine that drives a state machine, handles retries, emits events, and performs Saga-style rollbacks on failure.

---

## 2. High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                          Browser / Client                            │
│                                                                      │
│              Next.js 15  ·  React 19  ·  TanStack Query             │
│              shadcn/ui  ·  Recharts  ·  Monaco Editor               │
└────────────────────────────┬─────────────────────────────────────────┘
                             │  HTTP / SSE
┌────────────────────────────▼─────────────────────────────────────────┐
│                        Go / Gin REST API                             │
│                                                                      │
│    Auth Middleware (JWT)  ·  RBAC Middleware  ·  Request Validation  │
│                                                                      │
│    /api/environments   /api/workflows   /api/templates               │
│    /api/schedules      /api/metrics     /api/events                  │
│    /api/logs           /api/audit       /api/users                   │
└────────────────────────────┬─────────────────────────────────────────┘
                             │
              ┌──────────────┼────────────────┐
              │              │                │
     ┌────────▼──────┐  ┌───▼────────┐  ┌───▼──────────────┐
     │   Job Queue   │  │  Scheduler │  │   SSE Broker     │
     │  (channels)   │  │ (cron lib) │  │  (event → HTTP)  │
     └────────┬──────┘  └───┬────────┘  └──────────────────┘
              │              │ enqueue
              └──────┬───────┘
                     │
            ┌────────▼─────────┐
            │   Worker Pool    │
            │  (goroutines)    │
            └────────┬─────────┘
                     │
            ┌────────▼──────────────┐
            │   Workflow Engine     │
            │                       │
            │  ┌─────────────────┐  │
            │  │  State Machine  │  │
            │  └─────────────────┘  │
            │  ┌─────────────────┐  │
            │  │  Retry Engine   │  │
            │  └─────────────────┘  │
            │  ┌─────────────────┐  │
            │  │ Rollback Engine │  │
            │  └─────────────────┘  │
            │  ┌─────────────────┐  │
            │  │   Event Bus     │  │
            │  └─────────────────┘  │
            └────────┬──────────────┘
                     │
      ┌──────────────┼──────────────────────────┐
      │              │              │            │
┌─────▼──────┐ ┌────▼─────┐ ┌─────▼──────┐ ┌──▼───────┐
│ Namespace  │ │PostgreSQL│ │   Redis    │ │RabbitMQ  │
│ Executor   │ │ Executor │ │  Executor  │ │ Executor │
└────────────┘ └──────────┘ └────────────┘ └──────────┘
┌─────────────────────┐   ┌──────────────────────────┐
│  Deployment Executor│   │      DNS Executor        │
└─────────────────────┘   └──────────────────────────┘
                     │
            ┌────────▼─────────┐
            │     SQLite       │
            │                  │
            │  Users           │
            │  Projects        │
            │  Environments    │
            │  WorkflowExec    │
            │  WorkflowSteps   │
            │  Events          │
            │  AuditLogs       │
            │  Resources       │
            │  Configurations  │
            │  Templates       │
            │  Schedules       │
            └──────────────────┘
```

---

## 3. Component Design

### 3.1 REST API Layer

Built with Gin. Every request passes through:

1. **JWT Middleware** — validates Bearer token, extracts `user_id` and `role` into request context
2. **RBAC Middleware** — checks the role against a permission matrix for the route's required permission
3. **Request Validation** — Gin binding + custom validators; returns `400` with structured error envelope on failure

Response envelope:

```json
{
  "data": { ... },
  "error": null,
  "meta": { "request_id": "uuid", "timestamp": "..." }
}
```

Error envelope:

```json
{
  "data": null,
  "error": { "code": "WORKFLOW_NOT_FOUND", "message": "..." }
}
```

---

### 3.2 Authentication & RBAC

**JWT**

- Access token: 15-minute expiry, signed with HS256
- Refresh token: 7-day expiry, stored in SQLite for rotation tracking
- Claims: `{ user_id, role, exp }`

**Role Permission Matrix**

| Permission | Admin | Developer | Viewer |
|---|---|---|---|
| `environment:create` | ✅ | ✅ | ❌ |
| `environment:delete` | ✅ | ✅ | ❌ |
| `workflow:view` | ✅ | ✅ | ✅ |
| `workflow:cancel` | ✅ | ✅ | ❌ |
| `workflow:retry` | ✅ | ✅ | ❌ |
| `template:manage` | ✅ | ✅ | ❌ |
| `schedule:manage` | ✅ | ✅ | ❌ |
| `metrics:view` | ✅ | ✅ | ✅ |
| `audit:view` | ✅ | ❌ | ❌ |
| `user:manage` | ✅ | ❌ | ❌ |

The RBAC middleware reads the required permission from a route registration map and enforces it before the handler executes.

---

### 3.3 Job Queue

An in-process Go channel acting as the job queue:

```go
type Job struct {
    ID           string
    WorkflowType string
    Payload      map[string]interface{}
    EnqueuedAt   time.Time
}

type Queue struct {
    ch chan Job
}
```

- Buffered channel with configurable capacity (default 100)
- API handler writes to channel; returns `202 Accepted` immediately with workflow ID
- Worker pool reads from channel

The queue is intentionally in-memory to keep the project local-first. Persistence of the *workflow execution record* happens in SQLite before the job is enqueued, so a restart can replay unfinished workflows on boot.

---

### 3.4 Worker Pool

```go
type WorkerPool struct {
    workers    int
    queue      *Queue
    engine     *WorkflowEngine
    ctx        context.Context
    cancelFunc context.CancelFunc
    wg         sync.WaitGroup
}
```

- Spawns `N` goroutines (configurable; default 5) on startup
- Each worker loops on `queue.ch` until context is cancelled
- Graceful shutdown: cancel context → drain in-flight jobs → `wg.Wait()`
- Each worker carries its own timeout context derived from the parent

---

### 3.5 Workflow Engine

The engine is the central orchestrator. It receives a `Job`, resolves the workflow definition, and drives execution.

```go
type WorkflowEngine struct {
    definitions map[string]*WorkflowDefinition  // loaded from YAML
    repo        WorkflowRepository
    eventBus    *EventBus
    retrier     *RetryEngine
    rollbacker  *RollbackEngine
}

func (e *WorkflowEngine) Execute(ctx context.Context, job Job) error
```

**Execution loop:**

```
1. Load workflow definition for job.WorkflowType
2. Create WorkflowExecution record (state: Queued)
3. Set state → Running, persist
4. For each step in definition:
   a. Create WorkflowStep record (state: Pending)
   b. Set step state → Running
   c. Call executor.Execute(ctx, stepInput)
   d. On success: set step state → Succeeded, store output, continue
   e. On failure: invoke RetryEngine
      - If retry succeeds: step → Succeeded, continue
      - If retries exhausted: step → Failed → invoke RollbackEngine
5. On all steps succeeded: workflow → Completed
6. On rollback complete: workflow → Rolled Back
```

**Step context (output sharing):**

Each step's output is merged into a shared `map[string]interface{}` passed to subsequent steps. For example, `create_namespace` outputs `namespace_id`, which `provision_database` uses as input.

---

### 3.6 State Machine

States are modelled as typed constants with a transition guard:

```go
type WorkflowState string

const (
    StatePending    WorkflowState = "pending"
    StateQueued     WorkflowState = "queued"
    StateRunning    WorkflowState = "running"
    StateCompleted  WorkflowState = "completed"
    StateFailed     WorkflowState = "failed"
    StateRollingBack WorkflowState = "rolling_back"
    StateRolledBack WorkflowState = "rolled_back"
)

var validWorkflowTransitions = map[WorkflowState][]WorkflowState{
    StatePending:     {StateQueued},
    StateQueued:      {StateRunning},
    StateRunning:     {StateCompleted, StateFailed},
    StateFailed:      {StateRollingBack},
    StateRollingBack: {StateRolledBack},
}
```

`Transition(from, to)` validates against the map before persisting; invalid transitions return an error and abort execution.

---

### 3.7 Retry Engine

```go
type RetryPolicy struct {
    MaxAttempts int
    Strategy    RetryStrategy  // Fixed | ExponentialBackoff
    BaseDelay   time.Duration
}

func (r *RetryEngine) Execute(ctx context.Context, fn func() error, policy RetryPolicy) error
```

Strategies:

| Strategy | Delay formula |
|---|---|
| Fixed | `BaseDelay` (constant) |
| ExponentialBackoff | `BaseDelay * 2^(attempt-1)`, capped at `MaxDelay` |

On each failed attempt, a `RetryTriggered` event is emitted before sleeping. Attempts are logged with attempt number and error.

---

### 3.8 Rollback Engine (Saga Pattern)

Every executor implements an optional `Rollback(ctx, input)` method:

```go
type Executor interface {
    Execute(ctx context.Context, input StepInput) (StepOutput, error)
    Rollback(ctx context.Context, input StepInput) error  // optional
}
```

When the workflow engine detects a terminal failure:

1. Collect all steps that reached `Succeeded`
2. Reverse their order
3. For each: call `executor.Rollback()`, update step state → `Rolled Back` (or `Rollback Failed`)
4. Emit `RollbackStarted` and `RollbackCompleted` events
5. Set workflow state → `Rolled Back`

This is a Saga compensation pattern — each forward action has a defined undo. Rollback failures are logged but do not block subsequent rollback steps.

---

### 3.9 Event Bus

In-process pub/sub using Go channels:

```go
type EventBus struct {
    subscribers map[EventType][]chan Event
    mu          sync.RWMutex
}

type Event struct {
    ID           string
    Type         EventType
    WorkflowID   string
    StepID       string
    Payload      map[string]interface{}
    OccurredAt   time.Time
}
```

Event types: `WorkflowStarted`, `WorkflowCompleted`, `WorkflowFailed`, `StepStarted`, `StepSucceeded`, `StepFailed`, `RetryTriggered`, `RollbackStarted`, `RollbackCompleted`, `ResourceCreated`, `ResourceDeleted`

Subscribers:

- **Persistence subscriber** — writes every event to SQLite `events` table
- **Audit subscriber** — writes user-action events to `audit_logs`
- **SSE subscriber** — pushes events to the SSE broker for live UI updates
- **Metrics subscriber** — increments in-memory counters

---

### 3.10 SSE Broker (Real-Time Updates)

```
Client connects to GET /api/workflows/{id}/stream
         │
         ▼
SSE Broker creates a per-workflow channel
         │
         ▼
Event Bus delivers events for that workflow ID
         │
         ▼
Broker serialises events as SSE: data: {...}\n\n
         │
         ▼
Client receives live step updates
```

Clients receive events as they happen. On disconnect, the broker cleans up the channel. The frontend reconnects automatically using the browser's native `EventSource` API.

---

### 3.11 Scheduler

Built on `robfig/cron`:

```go
type Schedule struct {
    ID                 string
    WorkflowTemplateID string
    CronExpression     string
    Enabled            bool
    LastRunAt          *time.Time
    NextRunAt          *time.Time
}
```

On each tick, the scheduler constructs a `Job` from the linked template and enqueues it into the existing `Queue`. Scheduled runs are tagged `trigger: scheduled` in the `WorkflowExecution` record so they appear distinctly in history.

Schedules are loaded from SQLite on startup and re-synced on create/update/delete via the API.

---

### 3.12 Workflow Template Engine

Workflow definitions are stored as YAML:

```yaml
name: create-environment
steps:
  - id: validate
    executor: validate
    retry:
      max_attempts: 3
      strategy: fixed
      base_delay: 1s
  - id: provision_database
    executor: postgresql
    depends_on: [create_namespace]
    retry:
      max_attempts: 5
      strategy: exponential_backoff
      base_delay: 2s
    rollback: true
```

Templates authored via Monaco Editor are validated server-side:

- Required fields present
- All `depends_on` references resolve to defined step IDs
- No dependency cycles (topological sort check)
- Valid executor names

---

## 4. Data Model

### workflows (WorkflowExecution)

| Column | Type | Notes |
|---|---|---|
| id | TEXT | UUID |
| type | TEXT | e.g. `create-environment` |
| state | TEXT | State machine value |
| trigger | TEXT | `manual` or `scheduled` |
| input | JSON | Initial payload |
| started_at | DATETIME | |
| completed_at | DATETIME | Nullable |
| created_by | TEXT | User UUID |

### workflow_steps

| Column | Type | Notes |
|---|---|---|
| id | TEXT | UUID |
| workflow_id | TEXT | FK → workflows |
| name | TEXT | Step identifier |
| state | TEXT | Step state machine value |
| attempt | INT | Current attempt number |
| input | JSON | |
| output | JSON | Nullable; fed to next steps |
| error | TEXT | Nullable |
| started_at | DATETIME | |
| completed_at | DATETIME | |

### events

| Column | Type | Notes |
|---|---|---|
| id | TEXT | UUID |
| type | TEXT | EventType |
| workflow_id | TEXT | |
| step_id | TEXT | Nullable |
| payload | JSON | |
| occurred_at | DATETIME | |

### audit_logs

| Column | Type | Notes |
|---|---|---|
| id | TEXT | UUID |
| user_id | TEXT | |
| action | TEXT | e.g. `environment.create` |
| resource_type | TEXT | |
| resource_id | TEXT | |
| metadata | JSON | |
| occurred_at | DATETIME | |

### templates

| Column | Type | Notes |
|---|---|---|
| id | TEXT | UUID |
| name | TEXT | |
| definition | TEXT | YAML source |
| version | INT | Incremented on update |
| created_by | TEXT | |
| created_at | DATETIME | |

### schedules

| Column | Type | Notes |
|---|---|---|
| id | TEXT | UUID |
| template_id | TEXT | FK → templates |
| cron_expression | TEXT | |
| enabled | BOOL | |
| last_run_at | DATETIME | Nullable |
| next_run_at | DATETIME | Nullable |

---

## 5. Request Lifecycle

```
1.  Client: POST /api/environments  { name, config }
2.  RBAC: check environment:create permission
3.  Handler: validate request body
4.  Handler: create Environment record (state: pending)
5.  Handler: create WorkflowExecution record (state: queued)
6.  Handler: enqueue Job to Queue channel
7.  Handler: return 202 { workflow_id }

8.  Worker: dequeue Job
9.  Engine: load YAML definition
10. Engine: set WorkflowExecution → running
11. Engine: for each step:
      - set step → running
      - call executor.Execute()
      - on success: set step → succeeded, store output
      - on failure: retry via RetryEngine
        - if retries exhausted: set step → failed
          - invoke RollbackEngine (reverse compensation)
          - set workflow → rolling_back → rolled_back
12. Engine: all steps succeeded → set workflow → completed
13. Events emitted at each transition → persisted + SSE → client

14. Client: GET /api/workflows/{id}/stream (SSE)
    Receives live updates as steps complete
```

---

## 6. Concurrency Model

```
Main Goroutine
    │
    ├── HTTP Server (Gin)           — handles incoming requests
    │       └── enqueues Jobs → channel
    │
    ├── Worker 1 (goroutine)        — reads from channel, runs workflow
    ├── Worker 2 (goroutine)
    ├── Worker N (goroutine)
    │
    ├── Event Bus (goroutine)       — fan-out events to subscribers
    │
    ├── Scheduler (goroutine)       — fires cron ticks, enqueues jobs
    │
    └── SSE Broker (goroutine)      — manages per-workflow SSE channels
```

Shared state is protected:

- `WorkerPool` uses `sync.WaitGroup` for lifecycle
- `EventBus` uses `sync.RWMutex` on subscriber map
- `MetricsStore` uses `sync.Mutex` on counters
- SQLite accessed via a single GORM connection with WAL mode enabled (allows concurrent reads)

---

## 7. Executor Interface

```go
type StepInput struct {
    WorkflowID string
    StepName   string
    Context    map[string]interface{}  // shared output from prior steps
    Config     map[string]interface{}  // from workflow definition
}

type StepOutput struct {
    Resources map[string]interface{}  // e.g. { "database_url": "..." }
    Metadata  map[string]interface{}
}

type Executor interface {
    Name() string
    Execute(ctx context.Context, input StepInput) (StepOutput, error)
    Rollback(ctx context.Context, input StepInput) error
}
```

Each mock executor sleeps for a random duration (50–500ms) to simulate real infrastructure latency. A configurable `failure_rate` (0.0–1.0) causes random failures to exercise the retry and rollback paths.

---

## 8. Frontend Architecture

```
app/
├── (auth)/
│   └── login/          — public
├── (platform)/
│   ├── dashboard/       — metrics overview
│   ├── environments/    — list + create
│   ├── workflows/
│   │   ├── page.tsx     — history list
│   │   └── [id]/        — execution detail + SSE stream
│   ├── templates/       — Monaco Editor YAML authoring
│   ├── schedules/       — cron schedule management
│   ├── events/          — event timeline
│   ├── audit/           — audit log (Admin)
│   └── users/           — user management (Admin)

hooks/
├── useWorkflowStream    — wraps EventSource for SSE
├── useWorkflows         — TanStack Query for workflow list
├── useMetrics           — polling metrics endpoint

services/
├── api.ts               — typed fetch wrapper with auth headers
├── auth.ts              — token storage and refresh logic
```

State management is intentionally minimal — TanStack Query handles server state (caching, refetching, optimistic updates). Local UI state (form values, modal open/close) uses `useState`.

---

## 9. Observability

| Signal | Implementation | Storage |
|---|---|---|
| Structured logs | Zap logger, JSON format, workflow correlation ID | SQLite `logs` table |
| Events | Typed event structs, emitted at every state transition | SQLite `events` table |
| Metrics | In-memory atomic counters, polled by frontend | In-memory (reset on restart) |
| Audit logs | Written by event bus subscriber on user-action events | SQLite `audit_logs` table |

Metrics exposed at `GET /api/metrics`:

```json
{
  "running_workflows": 3,
  "completed_workflows": 142,
  "failed_workflows": 7,
  "success_rate": 0.953,
  "queue_depth": 2,
  "active_workers": 5,
  "avg_provisioning_time_ms": 4820,
  "total_retries": 23
}
```

---

## 10. Failure Scenarios & Handling

| Scenario | Behaviour |
|---|---|
| Transient executor failure | RetryEngine retries with backoff; emits `RetryTriggered` event |
| All retries exhausted | Step → Failed; RollbackEngine compensates completed steps in reverse |
| Worker panic | `recover()` in worker goroutine; workflow marked failed; worker continues |
| Context timeout | `ctx.Done()` checked in executor loop; workflow marked failed |
| SSE client disconnect | Broker closes client channel; no further events delivered; no impact on execution |
| Scheduler overlap | If previous run still active, scheduler skips tick and logs a warning |
| DB write failure | Logged; execution continues; consistency restored on next successful write |

---

## 11. Local Development

```
make dev
```

Starts via Docker Compose:

```
frontend  (Next.js, port 3000)
backend   (Go/Gin, port 8080)
```

SQLite database file persisted at `./data/platform.db` via Docker volume mount.

Environment variables:

```
JWT_SECRET=dev-secret
DB_PATH=./data/platform.db
WORKER_COUNT=5
QUEUE_CAPACITY=100
LOG_LEVEL=debug
```

No cloud accounts, no external APIs, no paid services required.

---

## 12. Future Architecture Extensions

| Extension | Notes |
|---|---|
| DAG-based execution | Replace sequential loop with topological sort; enable parallel steps |
| Prometheus metrics | Replace in-memory counters with `prometheus/client_golang` |
| OpenTelemetry tracing | Instrument executor calls with spans for distributed tracing |
| Kubernetes executor | Real `kubectl` calls instead of mock sleep |
| Terraform executor | Wrap `terraform apply` as a workflow step |
| Postgres backend | Replace SQLite for multi-instance deployments |
| Workflow versioning | Immutable template versions; executions reference version at run time |
| Plugin SDK | Define executor interface as a Go plugin; load `.so` at runtime |