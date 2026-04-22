# Bastion Design Document

## 1. Overview

Bastion is a self-hosted, policy-governed **sandboxed execution environment** designed for running of untrusted workloads, particularly AI-generated code and automation tasks.

It provides persistent, container-backed execution environments (sessions) with controlled access to system resources, filesystems, and networks, exposed via REST and WebSocket APIs.

The system is designed for **long-lived, concurrent, and observable execution environments** rather than ephemeral job execution.


## 2. Objectives

- Provide isolated execution environments using containerization
- Support persistent workspace state across restarts
- Enable concurrent execution of multiple processes per workspace
- Enforce fine-grained execution policies
- Provide full observability of execution lifecycle
- Support interactive and non-interactive execution modes
- Deterministic workspace lifecycle management
- Secure-by-default execution isolation
- Horizontal scalability of orchestration layer
- Minimal execution overhead for command dispatch
- Crash recovery and state reconciliation


## 3. System Architecture

### 3.1 High-Level Architecture

```id="9c0s9z"
Client (CLI / Agent / UI)
        │
        ▼
API Layer (REST + WebSocket)
        │
        ▼
Workspace Orchestrator
        │
        ▼
Container Runtime (Docker Engine)
        │
        ▼
Filesystem Layer (Bind Mounts)
        │
        ▼
Persistence Layer (SQLite + Logs)
```

 

### 3.2 Component Responsibilities

#### 3.2.1 API Layer

- Exposes REST endpoints for workspace lifecycle, execution, and file operations
- Exposes WebSocket endpoints for:
  - Interactive terminal sessions (TTY)
  - Streaming execution output
- Performs authentication and request validation

#### 3.2.2 Workspace Orchestrator

- Maintains workspace lifecycle state machine
- Tracks active executions per workspace
- Enforces concurrency limits
- Applies execution policies prior to dispatch
- Reconciles runtime state with persisted state

#### 3.2.3 Container Runtime Layer

- Executes workloads inside Docker containers
- Manages:
  - Container lifecycle (create/start/stop/delete)
  - Command execution via `docker exec`
  - TTY attachment for interactive sessions

- Enforces CPU, memory, and process limits

#### 3.2.4 Filesystem Layer

- Implements persistent storage via host bind mounts:

  ```
  /var/workspaces/{workspace_id} → /workspace
  ```

- Ensures:
  - Workspace isolation
  - Persistent state across container restarts
  - Path traversal prevention

#### 3.2.5 Persistence Layer

- SQLite used for workspace metadata persistence
- Structured logs stored for audit and replay

## 4. Core Abstractions

### 4.1 Session

A Session represents an isolated, persistent execution environment backed by a container.

```go id="g6k8qp"
type Session struct {
    ID           string
    Name         string
    Image        string
    ContainerID  string
    Status       string
    Mounts       []Mount
    EnvVars      map[string]string
    CreatedAt    time.Time
    LastActiveAt time.Time
}
```

#### Properties

- Long-lived lifecycle independent of executions
- Persistent filesystem via bind mounts
- Supports concurrent execution sessions

### 4.2 Execution Model

Execution is modeled as a stateless invocation over a session. Each execution maps to a Docker `exec` instance

| Type         | Semantics                        |
| ------------ | -------------------------------- |
| Interactive  | WebSocket-based TTY session      |
| Asynchronous | Returns execution ID for polling |

### 4.3 Terminal Subsystem

The terminal subsystem provides an interactive shell over WebSocket.

```id="7f5m3d"
Client ↔ WebSocket ↔ API Handler ↔ Docker Exec (TTY mode)
```

#### Capabilities

- Bidirectional streaming I/O
- Terminal resize events
- Persistent session bound to workspace container

 

### 4.4 Container Model

- One container per workspace
- Containers are persistent and long-running
- Idle state maintained via `sleep infinity`
- Execution occurs via `docker exec`

 

### 4.5 Filesystem Model

- Host bind mount provides persistence:

  ```
  /var/workspaces/{id} → /workspace
  ```

#### Guarantees

- Persistent across container restarts
- Strict workspace-level isolation
- No cross-workspace filesystem access
- Path validation enforced at API boundary

 

## 5. API Specification

### 5.1 Workspace Lifecycle APIs

```id="j9c1kq"
POST   /session
GET    /session
GET    /session/{id}
POST   /session/{id}/start
POST   /session/{id}/stop
DELETE /session/{id}
```

 

### 5.2 Execution APIs

```id="3x8v9a"
POST /session/{id}/exec
GET  /session/{id}/exec/{exec_id}
```

- Asynchronous execution returns execution handle

 

### 5.3 Terminal API

```id="v3m0qk"
GET /session/{id}/terminal (WebSocket Upgrade)
```

 

### 5.4 File APIs

```id="k2d9sn"
GET    /session/{id}/download
POST   /session/{id}/upload
DELETE /session/{id}/delete
GET    /session/{id}/list
```

 

## 6. Policy Engine

### 6.1 Policy Model

Execution behavior is governed by declarative policies. Example of config file.

```toml id="r3n8qp"
[server]
host = "0.0.0.0"
port = 8080

[execution]
mode = "sandbox"
max_concurrent_sessions = 10
session_ttl_minutes = 10
session_cleanup_interval_sec = 60
working_directory_base = "./sessions"
env_path="./config/.env"

[sandbox]
enabled = true
image = "python:3.11.15"
load_env = true
network_enabled = true
memory_mbs = 512
cpus = 0.5
pids = 128
job_ttl = 60

[filesystem]
max_upload_size_mbs = 10

[logging]
level = "info"
format = "text"
file = "./logs.json"

```

 

### 6.2 Enforcement Points

Policies are enforced at:

- Container initialization (resource constraints)
- Runtime execution (timeouts)
- Filesystem access layer
- Network configuration layer

 

## 7. Concurrency Model

- Multiple concurrent executions per workspace are supported
- Per-workspace execution limits enforced
- Container-level CPU/memory constraints enforced

 

## 8. Persistence & Reconciliation

### 8.1 Persistence Model

- Workspace state persisted in SQLite
- Required for crash recovery and restart consistency

 

### 8.2 Reconciliation Process

On system startup:

1. Load workspace state from SQLite
2. Fetch active Docker containers
3. Reconcile differences:
   - Missing container → mark workspace failed
   - Orphan container → cleanup or attach

4. Restore runtime consistency

## 9. Observability

Structured JSON logs include:

- Command execution
- Exit codes
- Execution duration
- Resource usage
- Policy evaluation results

## 10. Security Model

### 10.1 Authentication

- API key-based authentication

| Role   | Permissions               |
| ------ | ------------------------- |
| Admin  | Full system access        |
| Agent  | Execute + terminal access |
| Viewer | Read-only access          |

 

### 10.2 Isolation Guarantees

- Container-level process isolation
- Filesystem isolation via bind mounts
- Network isolation per policy

 

## 11. Resource Management

Resources enforced at container level:
- Timeout enforcement via execution context
- CPU/memory limits enforced by Docker runtime


## 12. Future Work

- Container pooling (warm execution environments)
- Snapshot/restore system for workspaces
- Distributed execution layer
- Multi-tenant isolation model
- Redis-backed coordination layer
- OpenTelemetry integration
- SSH / reverse proxy gateway
- Git-native workspace workflows
