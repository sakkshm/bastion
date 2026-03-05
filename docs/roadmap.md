# Bastion Roadmap

### Goal:

Build a lightweight, self-hosted, **policy-driven terminal for AI agents** with sandboxed execution, resource governance, and auditability.


## ~~Phase 0: Project Setup~~

**Tasks:**

* ~~Set up project structure~~
* ~~Choose web framework: **Chi**~~
* ~~Choose TOML parser~~
* ~~Logging (structured JSON)~~

## Phase 1: Core Execution Engine

**Goal:** Run commands safely and capture output.

**Features:**

* REST API:
    - ~~CreateSessionEndpoint  = "/session/create"~~
    - ~~GetSessionStatusEndpoint  = "/session/{id}/status"~~
    - ~~StartSessionEndpoint  = "/session/{id}/start"~~
	- SessionExecuteEndpoint = "/session/{id}/exec"
	- GetSessionLogsEndpoint = "/session/{id}/logs"
	- DeleteSessionEndpoint  = "/session/{id}"
* WebSocket endpoint for streaming stdout/stderr
* Docker sandbox (default)
* Bare-metal mode (optional)
* Workspace-scoped file access

**Implementation Notes (Go):**

* Use `os/exec` for bare-metal execution
* Use Docker SDK for Go (`github.com/docker/docker/client`) for sandboxed containers
* Use `context.Context` for timeout & cancellation
* Stream stdout/stderr over WebSocket via `gorilla/websocket`

**Test Checklist:**

* [ ] Commands execute successfully in Docker containers
* [ ] Commands execute in bare-metal mode
* [ ] Timeout and CPU/memory limits enforced
* [ ] Containers auto-cleaned after session


## Phase 2: Policy Engine

**Goal:** Governance over commands, filesystem, network.

**Features:**

* TOML policy file
* Allow/deny commands
* Workspace read/write paths
* Network enable/disable
* Resource limits per session
* Policy validation endpoint (`/validate`)

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

* Validate command against policy before execution

**Test Checklist:**

* [ ] Allowed commands succeed
* [ ] Denied commands blocked
* [ ] Workspace write restrictions enforced
* [ ] Network disabled correctly
* [ ] Dry-run simulation endpoint works



## Phase 3: File Management API

**Features:**

* Upload, Download, Delete files
* List directories
* Optional search
* Enforce workspace scoping

**Go Notes:**

* Use `io/ioutil` / `os` for file ops
* Validate paths to ensure they stay within workspace
* Optional search via `filepath.WalkDir`

**Test Checklist:**

* [ ] Upload only in workspace
* [ ] Cannot escape workspace
* [ ] Directory listing works
* [ ] Download returns correct content
* [ ] Delete removes files



## Phase 4: Audit Logging & Observability

**Features:**

* JSON structured logs
* Query logs via REST API (`/logs`)
* Replay execution metadata
* Log rotation support (`lumberjack` or custom)

**Test Checklist:**

* [ ] Logs include resource usage and policy decision
* [ ] Replay session reproduces execution
* [ ] Log rotation works



## Phase 5: Role-Based Access & Authentication

**Features:**

* API key auth
* Roles: `admin`, `agent`, `viewer`
* Admin: modify policy, view logs
* Agent: execute commands
* Viewer: read-only

**Go Notes:**

* Use middleware in Gin/Chi/Fiber to check API key & role
* Store keys in config or database (e.g., SQLite for hackathon)

**Test Checklist:**

* [ ] Role enforcement works correctly
* [ ] API key rotation works
* [ ] Unauthorized access blocked



## Phase 6: Configuration & Deployment

**Features:**

* CLI flags (Cobra)
* Environment variables
* Config files (`~/.config/bastion/config.toml`)
* Docker deployment
* Pip / uvx analog: `go build` or `go install`

**Test Checklist:**

* [ ] CLI overrides config files
* [ ] Env vars override defaults
* [ ] Docker deployment works
* [ ] Custom Dockerfile supported



## Phase 7: Bonus (Optional)

**Features:**

* Replay sessions deterministically
* Capability profiles
* Dry-run simulation
* CPU/memory benchmarks
* WebSocket streaming performance test

**Test Checklist:**

* [ ] Replay sessions reproduce output
* [ ] Profiles enforce limits
* [ ] Dry-run blocks destructive commands
* [ ] Streaming works reliably



## MVP Feature List (Golang Edition)

| Feature Category    | Feature                                                                                                                          |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Core Execution      | REST execution, WebSocket streaming, Docker sandbox, Bare-metal mode, Workspace-scoped file API                                  |
| Governance          | TOML policies, Command allow/deny, Filesystem scoping, Network toggle, CPU/memory limits, Timeout enforcement, Policy simulation |
| Observability       | JSON audit logs, Resource tracking, Execution metadata, Replay-ready sessions                                                    |
| Deployment & Config | Single Docker command, Go build/install, Config hierarchy, Custom Docker image support                                           |
| Security & Access   | API key auth, Role-based access (`admin`, `agent`, `viewer`)                                                                     |
| Optional / Bonus    | Deterministic replay, Capability profiles, Dry-run simulation, Benchmarks, Streaming tests                                       |

---

## Hackathon Submission Checklist 

- Clean architecture diagram
- REST + WebSocket endpoints documented
- Sample policy files included
- Demo workflow: upload file, run command, stream output, check logs
- CLI + Docker instructions
- Clear README + non-goals section
- Prebuilt Docker image
- Automated tests (policy, execution, files, auth)

