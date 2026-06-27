# Platform Orchestrator

> A local-first Internal Developer Platform (IDP) that simulates environment provisioning workflows with real-time execution, workflow orchestration, state management, retries, rollbacks, and observability.

---

## Overview

Platform Orchestrator is a workflow-driven platform engineering tool inspired by modern Internal Developer Platforms (IDPs) such as Backstage, Humanitec, and Harness.

The goal of the project is to simulate how platform teams provision application environments while demonstrating the engineering concepts behind workflow orchestration, background job execution, state machines, retries, rollbacks, event-driven systems, and observability.

Unlike a simple CRUD application, every user action starts a long-running workflow consisting of multiple dependent steps. The platform tracks execution in real time, persists workflow state, handles failures gracefully, and automatically performs rollback when required.

The project is completely local and requires no cloud accounts or paid services. Infrastructure components are mocked while maintaining behavior similar to real provisioning systems.

---

# Problem Statement

Modern platform teams receive requests such as:

* Create a development environment
* Provision a PostgreSQL database
* Allocate Redis cache
* Deploy an application
* Configure networking
* Delete environments after expiration

These requests involve multiple dependent operations that must execute in order while providing visibility into progress and handling failures safely.

This project simulates those workflows through a generic orchestration engine.

---

# Example Workflow

A developer submits a request to create an environment.

```
Developer Request
        │
        ▼
Validate Configuration
        │
        ▼
Reserve Resources
        │
        ▼
Create Namespace
        │
        ▼
Provision PostgreSQL
        │
        ▼
Provision Redis
        │
        ▼
Provision RabbitMQ
        │
        ▼
Deploy Application
        │
        ▼
Health Check
        │
        ▼
Environment Ready
```

If any stage fails:

```
Deployment Failed
        │
        ▼
Rollback Deployment
        │
        ▼
Delete Redis
        │
        ▼
Delete PostgreSQL
        │
        ▼
Delete Namespace
        │
        ▼
Workflow Failed
```

---

# Key Features

## Workflow Engine

* Generic workflow execution engine
* Sequential and dependent step execution
* Workflow definitions stored as YAML
* Dynamic workflow loading
* Step outputs shared across workflow
* Execution history

---

## Environment Provisioning

Provision mock infrastructure including:

* Namespace
* PostgreSQL
* Redis
* RabbitMQ
* Application Deployment
* DNS Entry

Every resource behaves similarly to a real cloud resource while remaining local.

---

## State Machine

Every workflow and every step maintains its own lifecycle.

Workflow states:

```
Pending

↓

Queued

↓

Running

↓

Completed

↓

Failed

↓

Rolling Back

↓

Rolled Back
```

Step states:

```
Pending

↓

Running

↓

Succeeded

↓

Failed

↓

Skipped
```

---

## Background Job Processing

User requests never execute synchronously.

Instead:

```
REST API

↓

Job Queue

↓

Worker Pool

↓

Workflow Engine
```

This allows multiple workflows to execute concurrently.

---

## Worker Pool

Multiple workers process provisioning jobs concurrently.

Supports:

* Configurable worker count
* Graceful shutdown
* Context cancellation
* Retry support
* Timeout handling

---

## Retry Engine

Transient failures are retried automatically.

Example:

```
Provision Redis

↓

Attempt 1

↓

Failed

↓

Retry

↓

Attempt 2

↓

Success
```

Retry policies include:

* Fixed retry
* Exponential backoff
* Maximum retry limit

---

## Rollback Engine

Every provisioning step defines an optional compensation action.

Example:

```
Create Database

↓

Create Redis

↓

Deploy Application

↓

Failure

↓

Delete Application

↓

Delete Redis

↓

Delete Database
```

This mimics Saga-style distributed transactions.

---

## Event System

Every workflow action emits events.

Examples:

* Workflow Started
* Resource Created
* Deployment Started
* Retry Triggered
* Rollback Started
* Workflow Completed

Events are persisted for auditing and displayed in the UI timeline.

---

## Real-Time Progress

The frontend receives live execution updates using Server-Sent Events (or WebSockets).

Operators can watch workflows execute step-by-step without refreshing the page.

---

## Logs

Every workflow generates structured logs.

Example:

```
12:00 Workflow Started

12:01 Namespace Created

12:02 PostgreSQL Provisioned

12:03 Redis Provisioned

12:04 Deployment Started

12:05 Health Check Passed
```

---

## Metrics Dashboard

Operational metrics include:

* Running workflows
* Completed workflows
* Failed workflows
* Success rate
* Queue size
* Average provisioning time
* Retry count
* Active workers

---

## Audit Log

Every user action is recorded.

Examples:

* Environment Created
* Workflow Cancelled
* Workflow Retried
* Environment Deleted

---

## Configuration Management

Workflow definitions are stored as YAML.

Example:

```yaml
name: Create Environment

steps:
  - validate
  - reserve_resources
  - create_namespace
  - provision_database
  - provision_cache
  - deploy_application
  - health_check
```

The workflow engine parses and executes these definitions dynamically.

---

# Architecture

```
                    Next.js UI
                         │
                         │
             REST API + Server Sent Events
                         │
                 Go Backend (Gin)
                         │
                Authentication (JWT)
                         │
                 Workflow Orchestrator
                         │
        ┌────────────────┼─────────────────┐
        │                │                 │
     Job Queue      State Machine      Event Bus
        │                │                 │
        └────────────────┼─────────────────┘
                         │
                    Worker Pool
                         │
     ┌───────────┬────────────┬───────────┬──────────┐
     │           │            │           │          │
 Namespace   PostgreSQL    Redis    Deployment   DNS
 Executor     Executor    Executor   Executor   Executor
                         │
                     SQLite
                         │
       Logs • Events • Audit • Resources
```

---

# Technology Stack

## Frontend

* Next.js 15
* React 19
* TypeScript
* Tailwind CSS
* shadcn/ui
* TanStack Query
* React Hook Form
* Zod
* Server-Sent Events (SSE)
* Recharts
* Monaco Editor

---

## Backend

* Go
* Gin
* GORM
* SQLite
* Context
* Goroutines
* Channels
* Worker Pool
* Zap Logger
* UUID

---

## Workflow Engine

* Generic workflow executor
* State machine
* Retry engine
* Rollback engine
* Dependency management
* Timeout handling
* Event publishing

---

## Persistence

SQLite stores:

* Users
* Projects
* Environments
* Workflow Executions
* Workflow Steps
* Events
* Audit Logs
* Resources
* Configurations

---

## Observability

* Structured Logging
* Event Timeline
* Metrics Dashboard
* Audit Logs
* Execution History

---

## Authentication

JWT-based authentication with Role-Based Access Control.

Roles:

* Admin
* Developer
* Viewer

---

## Local Development

Everything runs locally using Docker Compose.

```
Frontend

↓

Backend

↓

SQLite

↓

Mock Infrastructure Services
```

No cloud resources or external APIs are required.

---

# REST APIs

## Workflow APIs

```
POST   /api/workflows

GET    /api/workflows

GET    /api/workflows/{id}

POST   /api/workflows/{id}/retry

POST   /api/workflows/{id}/cancel
```

---

## Environment APIs

```
POST   /api/environments

GET    /api/environments

DELETE /api/environments/{id}
```

---

## Logs

```
GET /api/logs
```

---

## Events

```
GET /api/events
```

---

## Metrics

```
GET /api/metrics
```

---

# Project Structure

```
platform-orchestrator/

├── frontend/
│   ├── app/
│   ├── components/
│   ├── hooks/
│   ├── services/
│   └── lib/
│
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
│   ├── auth/
│   └── logger/
│
├── workflows/
│   ├── create-environment.yaml
│   ├── delete-environment.yaml
│   └── rollback.yaml
│
├── docker-compose.yml
├── Makefile
└── README.md
```

---

# Future Enhancements

* DAG-based workflow execution
* Parallel step execution
* Kubernetes executor
* Terraform executor
* Plugin SDK for custom resource executors
* Prometheus metrics integration
* Grafana dashboards
* OpenTelemetry distributed tracing
* Loki log aggregation
* GitOps workflow integration
* Workflow versioning
* Workflow templates
* Scheduled workflows
* Multi-project support

---

# Learning Objectives

This project demonstrates practical platform engineering concepts including:

* Workflow orchestration
* Distributed workflow execution
* Long-running asynchronous jobs
* Worker pool design
* Event-driven architecture
* State machine implementation
* Retry and rollback strategies
* Saga compensation pattern
* Concurrency with goroutines and channels
* Background job processing
* Observability and metrics
* REST API design
* Real-time UI updates
* Clean Architecture
* Local-first platform engineering development
