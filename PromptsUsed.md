# Reverse-Engineered Prompts for AegisIDP

Below is a curated sequence of high-quality, comprehensive prompts that I have used to generate this project from scratch, moving from the initial ideation down to the final polish. 

---

## 1. Initial Ideation & Readme Generation(Used Claude)
> **Prompt:**
> "I want to build a local-first Internal Developer Platform (IDP) called AegisIDP (Platform Orchestrator). The goal is to create a production-grade workflow orchestration engine in Go that simulates realistic environment provisioning (e.g., Kubernetes Namespaces, PostgreSQL, Redis, RabbitMQ). It should have real-time execution visibility via an SSE-powered UI, use the Saga pattern for rollbacks, and have robust state management. 
> 
> Please act as a Staff Engineer and draft a comprehensive `README.md`. It should outline the project overview, core features, technology stack (Go, Next.js, SQLite, Tailwind, shadcn/ui), and the high-level architecture."

## 2. Project Planning (Used Claude)
> **Prompt:**
> "Based on the `README.md` we just created, please write a detailed `Plan.md`. Break down the project into logical, sequential phases (e.g., Foundation, Workflow Engine, Provisioning Executors, Observability, Real-Time UI, Additional Features, and Polish). 
> 
> For each phase, list the specific backend and frontend deliverables, technical requirements, and how the features interlock."

## 3. System Design & Architecture
> **Prompt:**
> "Now, let's create a detailed `SystemDesign.md` to guide our implementation. I need you to document:
> 1. The high-level architecture and component diagram.
> 2. The database schema (Users, Projects, Environments, WorkflowExecutions, WorkflowSteps, Events, AuditLogs).
> 3. The REST API endpoints structure.
> 4. The event bus mechanism.
> 5. The workflow state machine transitions (Pending -> Queued -> Running -> Completed/Failed -> Rolling Back). 
> Ensure you detail how the worker pool and job queue will function."

## 4. Phase 1 — Foundation (Scaffolding & Auth)
> **Prompt:**
> "Let's start executing Phase 1 (Foundation). Please provide the code to scaffold the Go backend. 
> 
> Use Gin for routing, Zap for logging, and GORM with SQLite for the database. Set up the database models based on our system design. Implement JWT authentication middleware and an RBAC middleware supporting Admin, Developer, and Viewer roles. 
> 
> Finally, provide a `docker-compose.yml` to run the backend and an empty Next.js frontend service, along with the basic Next.js frontend setup using Tailwind CSS and shadcn/ui."

## 5. Phase 2 — Workflow Engine Core
> **Prompt:**
> "Moving on to Phase 2 (Workflow Engine Core). I need you to implement the core workflow engine in Go. 
> 
> First, create a YAML parser for workflow definitions (e.g., `create-environment.yaml`). Then, build a generic workflow executor that handles sequential step execution, step dependency resolution, and passing output context between steps. 
> 
> Implement an in-memory job queue using Go channels and a worker pool with configurable concurrency and context cancellation. Ensure the state machine correctly tracks workflow and step states."

## 6. Phase 3 — Provisioning Executors & Saga Pattern
> **Prompt:**
> "For Phase 3 (Provisioning Executors), let's implement the mock resource executors for Namespace, PostgreSQL, Redis, RabbitMQ, Deployment, and DNS. 
> 
> Each executor should simulate real infrastructure behavior with configurable random latency and failure rates. 
> 
> Most importantly, implement a retry engine with exponential backoff for transient failures, and a rollback engine that executes compensation actions in reverse order (Saga pattern) upon permanent failure."

## 7. Phase 4 — Event System + Observability
> **Prompt:**
> "Let's tackle Phase 4 (Event System + Observability). 
> 
> Implement an in-process event bus using Go channels that emits typed events (e.g., WorkflowStarted, ResourceCreated, RollbackStarted). Persist these events to SQLite for audit logging. 
> 
> Integrate Zap structured logging with workflow correlation IDs so we can trace specific executions. Finally, expose a `GET /api/metrics` endpoint that aggregates real-time operational data like workflow throughput, queue size, and active workers."

## 8. Phase 5 — Real-Time UI (SSE & Frontend)
> **Prompt:**
> "For Phase 5 (Real-Time UI), let's bridge the backend and frontend. 
> 
> On the backend, implement a Server-Sent Events (SSE) endpoint (`/api/workflows/{id}/stream`) that subscribes to the event bus and streams live step updates. 
> 
> On the frontend (Next.js 15, React 19, TanStack Query), build the workflow execution detail page. It should feature a step-by-step progress timeline, live status badges, a real-time log stream panel, and a metrics dashboard using Recharts."

## 9. Phase 6 — Additional Features (Monaco, Cron, RBAC)
> **Prompt:**
> "Now for Phase 6 (Additional Features). 
> 
> 1. Implement the granular RBAC permission matrix in the backend middleware to restrict specific actions (like user management to Admins).
> 2. Add a 'Workflow Templates' feature where users can author custom workflow YAMLs via a Monaco Editor integrated into the frontend. Include YAML schema validation.
> 3. Implement a cron-based scheduler using a library like `robfig/cron` to automate recurring workflows (e.g., nightly cleanups), persisting schedule configurations in SQLite."

## 10. Phase 7 — Hardening, Polish & Documentation
> **Prompt:**
> "Finally, let's wrap up with Phase 7 (Hardening + Polish). 
> 
> Please write a robust `Makefile` with targets for `dev`, `test`, `build`, and `migrate`. Add a standard API error handling envelope for consistent frontend consumption, and ensure pagination is applied to all list endpoints. 
> 
> Additionally, create a comprehensive `RunningGuide.md` that provides clear, step-by-step instructions for a new developer to set up, build, and run the AegisIDP project locally."
