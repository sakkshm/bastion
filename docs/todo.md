# TODO: 

## CRITICAL

* ~~WS stream dies unexpectedly, its not working at all, it only disconnects when error in terminal~~
* ~~Make API Token validation for WS~~
* Ensure session state stays synced with actual container state
* Ensure valid session status transitions only (no illegal state changes)
* Make session + container actions atomic and consistent
* Improve resource cleanup (no leaks, no zombie processes/containers)
* Container lifecycle cleanup after TTL / deletion (GC-like behavior)
* Improve logging consistency across system
* Standardize error handling across all modules
* Log execution metadata for all actions (command, latency, exit code, container/session IDs)
* Write tests (unit + integration + container lifecycle tests)
* Refactor handlers to only contain request logic (move business logic into services/utils layer)
* Improve abstraction consistency across system

## EXECUTION ENGINE

* Add support for terminal resize
* Improve ExecOptions support and flexibility (TTY, env, working dir, user, streams, privileged mode)
* Explore host config improvements (pidmode, ipcmode, utsmode, cgroup mode)
* Improve containerconfig / hostconfig / networkconfig structure
* Add support for Docker-in-Docker
* Add event-driven updates via Docker events

## CONTAINER LIFECYCLE & PERFORMANCE

* Add container pooling per session (warm pool model)
* Treat containers as compute pools, not persistent state
* Optimize container startup via pre-warmed Node/Python images
* Add prebuilt Node.js and Python sandbox images
* Support custom user-provided Docker images
* Support dynamic image selection at session creation

## SNAPSHOTS & STATE SYSTEM

* Add snapshot system (save / restore / resume workflows instantly)
* Add session forking support
* Move toward volume-based state instead of container persistence
* Add environment snapshots for reproducible workflows

## NETWORKING & ISOLATION

* Add network allowlist / denylist per session
* Add custom DNS configuration support
* Add internet access logging
* Improve network isolation controls

## DEV EXPERIENCE & PLATFORM FEATURES

* Add Git operations (clone, pull minimum)
* Make dependency installation easy (pip / npm inside sandbox)
* Add full web terminal (low-latency browser terminal)
* Add SDKs (Python, TypeScript)
* Add Dockerfile for project
* Add admin dashboard endpoint for system monitoring
* Add latency benchmarking system
* Add documentation/comments for system flows
* Write technical architecture write-up (Docker + isolation design)

## INFRASTRUCTURE & DEPLOYMENT

* Add Infrastructure-as-Code (IaC) for cloud deployment
* Improve system for self-hosting on cloud platforms
* Design for Kubernetes compatibility and elastic scaling

## AI / WORKLOAD USE CASE LAYER

* AI code execution / code interpreter support
* Data analysis & visualization sandbox
* Coding agents with persistent execution state
* Reinforcement learning environments (large-scale parallel sandboxes)
* Computer-use agents (desktop-like sandbox automation)
* Vibe coding runtime (run full apps in sandbox)
* Large dataset research workflows in isolated environments

## SECURITY & SYSTEMS RESEARCH

* runc internals
* containerd architecture
* gVisor syscall interception
* Firecracker microVMs
* Linux namespaces (PID, IPC, UTS, network)
* cgroups v2 resource control
* seccomp BPF filtering
* LXC components
* eBPF-based sandbox monitoring and enforcement system
* Kubernetes-style orchestration patterns
