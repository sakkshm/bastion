

# Bastion Design Document

## 1. Overview

Bastion is a lightweight, self-hosted, **policy-driven workspace runtime for AI agents**. It provides isolated execution environments using containers, enabling agents to execute commands, interact with filesystems, and access terminal interfaces under strict governance and observability.

The system is designed to behave as a **persistent development environment**. Workspaces are long-lived, support concurrent execution, and expose interactive terminals.

## 2. Design Goals

* **Isolation**: Strong sandboxing using containerized environments
* **Persistence**: Workspaces retain state across executions and restarts
* **Concurrency**: Multiple commands execute concurrently within a workspace
* **Governance**: Fine-grained policy enforcement (commands, filesystem, network)
* **Observability**: Full audit logs and execution metadata
* **Extensibility**: Modular architecture enabling future scaling (multi-tenant, pooling, distributed runtime)



## 3. Extended Goals

* Distributed scheduling across multiple nodes
* Kubernetes-native orchestration
* Snapshot-based prebuild systems
* Multi-region deployment
* Full IDE-like interface



## 4. System Architecture

### 4.1 High-Level Architecture

```
Client (CLI / Agent / UI)
        │
        ▼
API Layer (HTTP + WebSocket)
        │
        ▼
Workspace Manager (Orchestration)
        │
        ▼
Docker Runtime Layer
        │
        ▼
Filesystem Layer (Bind Mounts)
        │
        ▼
Data Layer (SQLite / Logs)
```



### 4.2 Layer Responsibilities

#### 1. API Layer

* HTTP endpoints for lifecycle, execution, files
* WebSocket endpoints for:
  * Terminal (TTY)
  * Streaming execution output
* Middleware for authentication and authorization

#### 2. Orchestration Layer (Workspace Manager)

* Manages workspace lifecycle
* Tracks state and metadata
* Enforces concurrency limits
* Applies policies before execution
* Handles reconciliation with runtime

#### 3. Runtime Layer

* Docker-based container execution
* Handles:
  * Container creation and lifecycle
  * Command execution (`exec`)
  * TTY attachment
* Enforces CPU/memory limits

#### 4. Filesystem Layer

* Provides persistent workspace storage
* Uses bind mounts:
  ```
  /var/workspaces/{workspace_id} → /workspace
  ```
* Ensures strict path isolation

#### 5. Data Layer

* Postgres:
  * Workspace metadata
  * API keys and roles
* Log storage:
  * Structured JSON logs
* Future:
  * Redis (caching, pub/sub)



## 5. Core Concepts

### 5.1 Workspace

A workspace is a **long-lived containerized environment**.

* Backed by a Docker container
* Has persistent filesystem
* Supports concurrent execution
* Represents the primary unit of isolation

```go
type Workspace struct {
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



### 5.2 Execution Model

#### Types of Execution

| Type         | Behavior                    |
|  |  |
| Interactive  | WebSocket + TTY             |
| Synchronous  | HTTP streaming response     |
| Asynchronous | Returns `exec_id`, pollable |

#### Key Properties

* Each execution maps to a Docker `exec` process
* Execution is stateless relative to workspace
* Multiple executions can run simultaneously


### 5.3 Terminal (TTY)

Interactive shell access is provided via WebSocket.

#### Data Flow

```
Browser/Agent WS
    ↔
Go WebSocket Handler
    ↔
Docker Exec (TTY mode)
```

#### Features

* Bidirectional input/output
* Terminal resize support
* Persistent session within container



### 5.4 Container Model

* Each workspace corresponds to a container
* Containers are persistent and run:

```bash
sleep infinity
```

* Execution occurs via `docker exec`
* Containers are treated as **ephemeral runtime instances**, while workspace metadata is authoritative



### 5.5 Filesystem Model

* Bind mounts provide persistence:

```
/var/workspaces/{id} → /workspace
```

* Guarantees:

  * Persistence across container restarts
  * Isolation per workspace
* Path validation prevents directory traversal



## 6. API Design

### 6.1 Workspace APIs

```
POST   /workspaces
GET    /workspaces
GET    /workspaces/{id}

POST   /workspaces/{id}/start
POST   /workspaces/{id}/stop
DELETE /workspaces/{id}
```



### 6.2 Execution APIs

```
POST   /workspaces/{id}/exec
GET    /workspaces/{id}/exec/{exec_id}   (async only)
```

* Streaming response for synchronous execution
* Async execution returns identifier for polling



### 6.3 Terminal API

```
GET /workspaces/{id}/terminal  (WebSocket upgrade)
```



### 6.4 File APIs

```
GET    /workspaces/{id}/files
POST   /workspaces/{id}/files
DELETE /workspaces/{id}/files
```



### 6.5 Observability APIs

```
GET /logs
```



## 7. Policy Engine

### 7.1 Overview

Policies define constraints on execution and resource usage.

### 7.2 Policy Structure

```go
type Policy struct {
    Commands struct {
        Allow []string
        Deny  []string
    }
    Filesystem struct {
        WritablePaths []string
        ReadOnlyPaths []string
    }
    Network  bool
    Limits   struct {
        CPU          float64
        MemoryMB     int
        TimeoutSecs  int
    }
}
```



### 7.3 Enforcement Points

* Before command execution
* At container creation:

  * CPU/memory limits
* At runtime:

  * Timeout enforcement
* Filesystem:

  * Path validation
* Network:

  * Enabled/disabled via container config



## 8. Concurrency Model

* Multiple executions per workspace
* Controlled via atomic counters:

```go
activeExecs atomic.Int32
maxExecs    int32
```

* Limits:

  * Per workspace exec limit
  * Container-level CPU/memory constraints



## 9. Persistence and Reconciliation

### 9.1 Persistence

* Workspace metadata stored in SQLite
* Required for:

  * Restart recovery
  * Lifecycle tracking



### 9.2 Reconciliation

On server startup:

* Fetch all workspaces from DB
* Fetch running containers from Docker
* Reconcile differences:

  * Missing container → mark failed
  * Orphan container → cleanup or attach



### 9.3 Event Monitoring

* Subscribe to Docker events:

  * Container start/stop/die
* Update workspace state accordingly



## 10. Observability

### 10.1 Logging

* Structured JSON logs
* Includes:

  * Command executed
  * Exit code
  * Duration
  * Resource usage
  * Policy decisions



### 10.2 Replay

* Store execution metadata
* Enable replay/debugging (non-deterministic for MVP)



## 11. Security Model

### 11.1 Authentication

* API key-based authentication

### 11.2 Authorization

* Role-based access:

| Role   | Permissions        |
|  |  |
| Admin  | Full access        |
| Agent  | Execute + terminal |
| Viewer | Read-only          |



### 11.3 Isolation

* Container-level isolation
* Filesystem isolation via mounts
* Optional network isolation



## 12. Resource Management

* CPU/memory limits applied at container level:

```go
container.Resources{
    Memory:   512 * 1024 * 1024,
    NanoCPUs: 1_000_000_000,
}
```

* Execution timeouts enforced via context



## 13. Failure Handling

* Container crashes → detected via events
* Workspace marked failed
* Optional restart policy

Failure cases handled:

* Container OOM
* Exec timeout
* Network failures
* API disconnects



## 14. Future Extensions

* Warm container pools (low-latency startup)
* Snapshotting and prebuilds
* S3/MinIO-backed storage
* Multi-tenancy (org/user isolation)
* Redis (caching, pub/sub, locks)
* OpenTelemetry-based observability
* Proxy and SSH gateway
* Git integration
* Distributed execution



## 15. Key Constraints

* No sequential job queue
* No in-memory-only state
* Workspace is the primary abstraction
* Execution must be concurrent and stateless
* Container lifecycle must be recoverable



## 16. Summary

Bastion is designed as a **workspace-native execution runtime** that prioritizes:

* Persistent environments over ephemeral jobs
* Direct execution over queued processing
* Interactive workflows over batch systems

This architecture enables AI agents to operate in a controlled, auditable, and reproducible environment while maintaining flexibility and performance.
