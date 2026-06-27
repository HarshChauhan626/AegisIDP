# Platform Orchestrator — Project Plan

## Overview

Platform Orchestrator is a local-first Internal Developer Platform (IDP) that simulates environment provisioning workflows with real-time execution, orchestration, state management, retries, rollbacks, and observability. This plan outlines phased delivery across backend, frontend, infrastructure, and additional features.

---

## Goals

- Build a production-grade workflow orchestration engine in Go
- Simulate realistic environment provisioning (Namespace, PostgreSQL, Redis, RabbitMQ, DNS)
- Deliver real-time execution visibility via SSE-powered UI
- Demonstrate platform engineering concepts: Saga pattern, state machines, worker pools, retry/rollback
- Run entirely locally with no cloud dependencies

---

## Phases

### Phase 1 — Foundation (Week 1–2)

**Backend**

- Project scaffold: Go module, Gin router, Zap logger, GORM + SQLite
- Database schema: Users, Projects, Environments, WorkflowExecutions, WorkflowSteps, Events, AuditLogs, Resources, Configurations
- JWT authentication middleware (login, refresh, protected routes)
- RBAC middleware (Admin, Developer, Viewer roles)
- Basic REST API skeleton with health check endpoint
- Docker Compose setup: backend + frontend services

**Frontend**

- Next.js 15 + React 19 project scaffold
- TanStack Query + Zod + React Hook Form setup
- Tailwind CSS + shadcn/ui component library integration
- Authentication flow: Login page, JWT token storage, protected routes
- Layout shell: sidebar navigation, header, notification area

**Deliverables**

- Authenticated API with role enforcement
- Frontend authenticates and renders a dashboard shell
- Docker Compose brings up both services

---

### Phase 2 — Workflow Engine Core (Week 3–4)

**Workflow Engine**

- YAML workflow definition parser (`create-environment.yaml`, `delete-environment.yaml`, `rollback.yaml`)
- Generic workflow executor: sequential step execution, step dependency resolution
- State machine implementation for workflows (Pending → Queued → Running → Completed → Failed → Rolling Back → Rolled Back)
- State machine for steps (Pending → Running → Succeeded → Failed → Skipped)
- Step output context: outputs from one step passed to subsequent steps
- Execution history persistence

**Job Queue + Worker Pool**

- In-memory job queue using Go channels
- Worker pool with configurable worker count
- Context cancellation + graceful shutdown
- Job dispatcher: REST API enqueues, workers dequeue and invoke workflow engine

**Deliverables**

- Submit a workflow via REST API → executes asynchronously through worker pool → state persisted in SQLite
- Workflow steps execute in order with output sharing

---

### Phase 3 — Provisioning Executors (Week 5)

**Mock Resource Executors**

Each executor simulates real infrastructure behaviour (random latency, configurable failure rates):

| Executor | Provision Action | Rollback Action |
|---|---|---|
| Namespace | Create namespace | Delete namespace |
| PostgreSQL | Provision database | Drop database |
| Redis | Provision cache | Delete cache |
| RabbitMQ | Provision message broker | Delete broker |
| Deployment | Deploy application | Rollback deployment |
| DNS | Register DNS entry | Remove DNS entry |

**Retry Engine**

- Fixed retry policy
- Exponential backoff retry policy
- Maximum retry limit enforcement
- Per-step retry configuration

**Rollback Engine**

- Compensation actions registered per step
- On workflow failure: execute rollback in reverse order (Saga pattern)
- Rollback state tracked independently per step

**Deliverables**

- Full `create-environment` workflow executes end-to-end with all six provisioners
- Transient failures retried automatically
- On permanent failure, compensation actions roll back completed steps

---

### Phase 4 — Event System + Observability (Week 6)

**Event Bus**

- In-process event bus using Go channels
- Event types: WorkflowStarted, ResourceCreated, DeploymentStarted, RetryTriggered, RollbackStarted, WorkflowCompleted, WorkflowFailed
- Event persistence to SQLite
- Audit log recording for all user-initiated actions

**Structured Logging**

- Zap logger integrated throughout workflow engine and executors
- Per-workflow log streams stored to SQLite
- Log levels: INFO, WARN, ERROR with workflow correlation IDs

**Metrics**

- In-memory metrics aggregation
- `GET /api/metrics` endpoint: running workflows, completed, failed, success rate, queue size, average provisioning time, retry count, active workers

**Deliverables**

- All workflow actions emit typed events
- Audit trail queryable via API
- Metrics endpoint returns real-time operational data

---

### Phase 5 — Real-Time UI (Week 7)

**SSE Streaming**

- `GET /api/workflows/{id}/stream` — SSE endpoint streaming live step updates
- Event-to-SSE bridge: event bus → SSE response writer
- Reconnection handling

**Frontend — Workflow Views**

- Environments list page: create, delete, view status
- Workflow execution detail page:
  - Step-by-step progress timeline
  - Live status badges (Pending / Running / Succeeded / Failed)
  - Real-time log stream panel
  - Retry and Cancel action buttons
- Workflow history list with filters

**Frontend — Observability**

- Metrics dashboard with Recharts: workflow throughput, success/failure rates, queue depth, worker utilisation
- Event timeline feed (auto-updates via polling or SSE)
- Audit log table with pagination and filters

**Deliverables**

- Operators can watch a workflow execute step-by-step in real time without refreshing
- Full metrics and audit views operational

---

### Phase 6 — Additional Features (Week 8)

#### Feature 1 — Role-Based Access Control (RBAC)

Granular permission enforcement beyond basic role assignment.

**Roles and Permissions**

| Action | Admin | Developer | Viewer |
|---|---|---|---|
| Create environment | ✅ | ✅ | ❌ |
| Delete environment | ✅ | ✅ | ❌ |
| Cancel workflow | ✅ | ✅ | ❌ |
| Retry workflow | ✅ | ✅ | ❌ |
| View workflows | ✅ | ✅ | ✅ |
| View metrics | ✅ | ✅ | ✅ |
| Manage users | ✅ | ❌ | ❌ |
| View audit log | ✅ | ❌ | ❌ |

**Implementation**

- Permission matrix stored in config; enforced via Go middleware on every protected route
- JWT claims carry role; middleware extracts and validates against required permission
- Frontend hides disallowed actions based on decoded role (non-authoritative; backend remains source of truth)
- User management API (Admin only): create, update role, deactivate users

---

#### Feature 2 — Workflow Templates + Monaco Editor

Allow teams to author and manage custom workflow definitions directly in the UI.

**Backend**

- `GET /api/templates` — list saved workflow templates
- `POST /api/templates` — save new template (validates YAML structure)
- `PUT /api/templates/{id}` — update existing template
- `DELETE /api/templates/{id}` — remove template
- YAML schema validation on save: required fields, valid step names, dependency cycle detection

**Frontend**

- Template library page: list, create, edit, delete templates
- Monaco Editor integration with YAML syntax highlighting and basic schema validation
- "Run Template" button: instantiate a workflow execution from a saved template
- Template version history (append-only; no overwrites)

**Why it matters:** Platform teams can extend the orchestrator with new provisioning workflows without modifying Go source code.

---

#### Feature 3 — Scheduled Workflows (Cron-based)

Automate recurring operations such as environment expiry cleanup, nightly health checks, and resource audits.

**Backend**

- Cron scheduler using a lightweight Go cron library (e.g. `robfig/cron`)
- Schedule configuration stored in SQLite: `{workflow_template_id, cron_expression, enabled, last_run_at, next_run_at}`
- Scheduler loop: on tick, enqueues a workflow job into the existing job queue — reuses the entire execution pipeline
- `GET /api/schedules` — list schedules
- `POST /api/schedules` — create schedule (validate cron expression)
- `PATCH /api/schedules/{id}` — enable/disable schedule
- `DELETE /api/schedules/{id}` — remove schedule
- Scheduled runs appear in workflow execution history tagged as `trigger: scheduled`

**Frontend**

- Schedules management page: create, enable/disable, delete schedules
- Next run time display with human-readable cron description
- Execution history filtered by trigger type (manual vs scheduled)

**Why it matters:** Enables automated environment lifecycle management — e.g. delete development environments every night, run health checks hourly — without manual intervention.

---

### Phase 7 — Hardening + Polish (Week 9)

- Integration tests: workflow execution end-to-end, retry and rollback paths
- API error handling standardisation (consistent error response envelope)
- Request validation on all POST/PUT endpoints
- Pagination on all list endpoints
- Rate limiting middleware
- README: setup guide, architecture overview, API reference
- Makefile targets: `make dev`, `make test`, `make build`, `make migrate`
- Docker Compose health checks and service dependency ordering

---

## API Summary

### Environments
```
POST   /api/environments
GET    /api/environments
DELETE /api/environments/{id}
```

### Workflows
```
POST   /api/workflows
GET    /api/workflows
GET    /api/workflows/{id}
GET    /api/workflows/{id}/stream   (SSE)
POST   /api/workflows/{id}/retry
POST   /api/workflows/{id}/cancel
```

### Templates
```
GET    /api/templates
POST   /api/templates
PUT    /api/templates/{id}
DELETE /api/templates/{id}
```

### Schedules
```
GET    /api/schedules
POST   /api/schedules
PATCH  /api/schedules/{id}
DELETE /api/schedules/{id}
```

### Observability
```
GET    /api/logs
GET    /api/events
GET    /api/metrics
GET    /api/audit
```

### Auth + Users
```
POST   /api/auth/login
POST   /api/auth/refresh
GET    /api/users          (Admin)
POST   /api/users          (Admin)
PATCH  /api/users/{id}     (Admin)
```

---

## Technology Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 15, React 19, TypeScript, Tailwind CSS, shadcn/ui, TanStack Query, Recharts, Monaco Editor |
| Backend | Go, Gin, GORM, SQLite, Zap, UUID |
| Workflow Engine | Custom executor, state machine, retry engine, rollback engine, Saga pattern |
| Real-time | Server-Sent Events (SSE) |
| Scheduler | robfig/cron |
| Auth | JWT + RBAC middleware |
| Dev | Docker Compose, Makefile |

---

## Project Structure

```
platform-orchestrator/
├── frontend/
│   ├── app/
│   ├── components/
│   ├── hooks/
│   ├── services/
│   └── lib/
├── backend/
│   ├── cmd/
│   ├── api/
│   ├── workflow/
│   ├── workers/
│   ├── queue/
│   ├── executors/
│   ├── services/
│   ├── repository/
│   ├── models/
│   ├── middleware/
│   ├── events/
│   ├── metrics/
│   ├── scheduler/
│   ├── auth/
│   └── logger/
├── workflows/
│   ├── create-environment.yaml
│   ├── delete-environment.yaml
│   └── rollback.yaml
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## Success Criteria

- A developer can submit an environment creation request and watch all six provisioning steps execute in real time
- Injected failures trigger automatic retries; permanent failures trigger full Saga-style rollback
- Admins can manage users and roles; Developers can manage environments; Viewers have read-only access
- Custom workflow templates can be authored in the Monaco Editor and executed on demand or on a cron schedule
- All workflow actions appear in the event timeline and audit log
- Metrics dashboard reflects live system state