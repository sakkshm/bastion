# Bastion Roadmap

### Goal:

Build a lightweight, self-hosted, **policy-driven workspace runtime for AI agents** with sandboxed execution, resource governance, and auditability.


## ~~Phase 0: Project Setup~~

**Tasks:**

* ~~Set up project structure~~
* ~~Choose web framework: **Chi**~~
* ~~Choose TOML parser~~
* ~~Logging (structured JSON)~~

**Additions:**

* Define layered architecture early:

  * API Layer
  * Workspace Manager (Orchestration)
  * Runtime Layer (Docker)
  * Filesystem Layer
  * Data Layer
* Establish structured logging format compatible with audit + replay



## Phase 1: Core Workspace Runtime

**Goal:** Persistent, concurrent execution environments.

**Features:**

* REST API:
  * ~~CreateSessionEndpoint  = "/session/create"~~ → `/workspaces`
  * ~~GetSessionStatusEndpoint  = "/session/{id}/status"~~ → `/workspaces/{id}`
  * ~~StartSessionEndpoint  = "/session/{id}/start"~~ → `/workspaces/{id}/start`
  * ~~StopSessionEndpoint  = "/session/{id}/start"~~ → `/workspaces/{id}/stop`
  * ~~DeleteSessionEndpoint  = "/session/{id}"~~ → `/workspaces/{id}`
  * ~~SessionExecuteEndpoint = "/session/{id}/exec"~~ → `/workspaces/{id}/exec`
  * ~~Polling endpoint: `/workspaces/{id}/exec/{exec_id}`~~ 
  ~~* NEW: `/workspaces/{id}/terminal` (WebSocket TTY)~~
~~* WebSocket endpoint for:~~
~~  * Streaming stdout/stderr (exec)~~
~~  * Interactive terminal (TTY)~~
* ~~Docker sandbox (default)~~
* ~~Persistent containers~~

**Additions:**

~~* Execution model explicitly supports:~~
~~  * Interactive (TTY over WS)~~
  * ~~Async (exec_id-based)~~
*~~ Each exec maps to independent Docker process ~~
*~~ Workspace lifecycle is independent of execution lifecycle~~
* Terminal uses bidirectional streaming + resize events

**Implementation Notes (Go):**

~~* Stream stdout/stderr via:~~
~~  * WebSocket (interactive)~~
~~  * HTTP streaming (non-interactive)~~
~~* Maintain per-workspace:~~
~~  * ContainerID~~
~~  * Active exec counter (atomic)~~


**Additions:**

~~* Execution must be stateless:~~
~~  * No shared buffers~~
~~  * No per-workspace locks blocking exec~~

**Test Checklist:**

~~* Commands execute successfully in Docker containers~~
~~* Multiple commands run concurrently in same workspace~~
~~* Interactive terminal works (TTY + resize)~~
~~* Timeout and CPU/memory limits enforced~~
~~* Containers persist across exec calls~~
* Containers restart correctly after stop/start
~~* Exec does not block other execs in same workspace~~


## Phase 1.5: File Management API

**Features:**

* ~~Workspace-scoped file access~~
* ~~Bind-mounted host directories (`/sessions/{session_id}`)~~
* ~~Upload, Download, Delete files~~
* ~~List directories~~
* ~~Enforce workspace scoping~~

**Additions:**

* Filesystem guarantees:
  * ~~Isolation per workspace~~
  * ~~No directory traversal~~


**Test Checklist:**

* [ ]~~ Upload only in workspace~~
* [ ] ~~Cannot escape workspace~~
* [ ]~~ Directory listing works~~
* [ ]~~ Download returns correct content~~
* [ ] ~~Files persist across container restarts~~
* [ ] ~~Delete removes files~~
* [ ] ~~Path traversal attacks blocked~~


## Phase 2: Workspace Persistence (CRITICAL)

**Goal:** Survive restarts, avoid orphan containers.

**Features:**

* Persistent workspace store (SQLite for MVP)
* Store:
  * ID, Image, ContainerID, Status
  * Mounts, EnvVars, CreatedAt
* Startup reconciliation with Docker
* Container lifecycle tracking

**Additions:**

* Workspace record is the **source of truth**, not container state

* Containers are treated as recoverable/replaceable runtime artifacts

* Reconciliation logic required:

  * Missing container → mark workspace failed
  * Orphan container → cleanup or attach

* Event-driven updates via Docker events:

  * start / stop / die

**Test Checklist:**

* [ ] Server restart preserves workspace state
* [ ] Orphan containers detected and handled
* [ ] Status reflects actual container state
* [ ] Container crash updates workspace status



## Phase 3: Policy Engine

**Goal:** Governance over commands, filesystem, network.

**Features:**

* TOML policy file (per workspace or global)
* Allow/deny commands
* Workspace read/write paths
* Network enable/disable
* Resource limits per workspace
* Policy validation endpoint (`/validate`)

**Additions:**

* Policy enforcement points:

  * Pre-execution (command validation)
  * Container creation (resource + network)
  * Runtime (timeouts)
* Policy must be deterministic and side-effect free
* Dry-run mode simulates execution without side effects

**Implementation Notes (Go):**

* Load TOML using `BurntSushi/toml`
* Define Go structs for policy:

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

* Validate command before exec (not via queue)
* Enforce limits at:

  * Container level (CPU/memory)
  * Exec level (timeout)

**Test Checklist:**

* [ ] Allowed commands succeed
* [ ] Denied commands blocked
* [ ] Workspace write restrictions enforced
* [ ] Network disabled correctly
* [ ] Dry-run simulation endpoint works
* [ ] Policy enforcement is consistent across exec types



## Phase 4: Audit Logging & Observability

**Features:**

* JSON structured logs
* Query logs via REST API (`/logs`)
* Execution metadata:
  * Command
  * Exit code
  * Duration
  * Resource usage
* Replay execution metadata
* Log rotation support (`lumberjack` or custom)

**Additions:**

* Logs must include:

  * Policy decision (allow/deny)
  * Workspace ID
  * Exec ID
* Logging must not block execution path (async or buffered)
* Replay is metadata-based (non-deterministic for MVP)

**Test Checklist:**

* [ ] Logs include resource usage and policy decision
* [ ] Replay session reproduces execution metadata
* [ ] Log rotation works
* [ ] Exec + terminal events logged
* [ ] Logs are consistent across restart



## Phase 5: Role-Based Access & Authentication

**Features:**

* API key auth
* Roles: `admin`, `agent`, `viewer`
* Admin: modify policy, manage workspaces, view logs
* Agent: execute commands, use terminal
* Viewer: read-only

**Additions:**

* Auth enforced at API layer via middleware
* Role-based restrictions applied per endpoint
* Future extension: multi-tenant isolation

**Go Notes:**

* Use middleware in Chi to enforce auth
* Store keys in config or database (SQLite for MVP)

**Test Checklist:**

* [ ] Role enforcement works correctly
* [ ] API key rotation works
* [ ] Unauthorized access blocked
* [ ] Terminal access restricted correctly



## Phase 6: Configuration & Deployment

**Features:**

* CLI flags (Cobra)
* Environment variables
* Config files (`~/.config/bastion/config.toml`)
* Docker deployment
* Custom Docker image support (user-provided)
* `go build` / `go install`

**Additions:**

* Config precedence:

  ```
  CLI > Env > Config file > Defaults
  ```

* Workspace image must be configurable per request

* Deployment must support reproducible environments

**Test Checklist:**

* [ ] CLI overrides config files
* [ ] Env vars override defaults
* [ ] Docker deployment works
* [ ] Custom Dockerfile supported
* [ ] Workspace image selection works
* [ ] Config precedence works correctly



## Phase 7: Bonus (Optional)

**Features:**

* Async exec jobs (background tasks only)
* Replay sessions deterministically
* Capability profiles (predefined policies)
* Dry-run simulation
* CPU/memory benchmarks
* WebSocket streaming performance test
* Container pooling (optional optimization)

**Additions:**

* Container pooling:

  * Pre-warmed containers for low latency
  * Optional optimization (not required for MVP)
* Async jobs must not affect interactive execution path

**Test Checklist:**

* [ ] Async exec works with polling
* [ ] Replay sessions reproduce output
* [ ] Profiles enforce limits
* [ ] Dry-run blocks destructive commands
* [ ] Streaming works reliably
* [ ] Pooling improves startup latency (if implemented)



## MVP Feature List (Golang Edition)

| Feature Category    | Feature                                                                                                                          |
| - | -- |
| Core Execution      | Workspace lifecycle, concurrent exec, WebSocket terminal, Docker sandbox, Bare-metal mode, Workspace-scoped file API             |
| Governance          | TOML policies, Command allow/deny, Filesystem scoping, Network toggle, CPU/memory limits, Timeout enforcement, Policy simulation |
| Observability       | JSON audit logs, Resource tracking, Execution metadata, Replay-ready sessions                                                    |
| Deployment & Config | Single Docker command, Go build/install, Config hierarchy, Custom Docker image support                                           |
| Security & Access   | API key auth, Role-based access (`admin`, `agent`, `viewer`)                                                                     |
| Optional / Bonus    | Async exec, Deterministic replay, Capability profiles, Dry-run simulation, Benchmarks, Streaming tests                           |



## Hackathon Submission Checklist

* Clean architecture diagram (**workspace-based, no queue**)
* REST + WebSocket endpoints documented
* Sample policy files included
* Demo workflow: create workspace → upload file → exec → terminal → check logs
* CLI + Docker instructions
* Clear README + non-goals section
* Prebuilt Docker image
* Automated tests (policy, execution, files, auth)

**Additions:**

* Demonstrate concurrent exec (multiple commands running simultaneously)
* Demonstrate terminal interaction (TTY)
* Show restart recovery (workspace persistence)
* Include failure scenarios (container crash, timeout)