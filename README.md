# Bastion

**A self-hosted, sandboxed execution environment for AI agents and automation tools to run untrusted code.**

Run commands, manage files, and execute code in a controlled environment **without risking your host machine**, via a simple REST + WebSocket API. 

> [`Get started  here.`](./docs/get-started.md)
 

## Why Bastion?

Running AI-generated code locally is risky. Bastion provides:

* **Isolated execution** : containerized environments protect your host
* **Policy + resource control** : restrict commands, filesystem, network, CPU, memory, and time
* **Persistent workspaces** : long-lived environments instead of ephemeral jobs
* **Concurrent execution** : multiple commands run in parallel
* **Realtime interaction** : streaming output + interactive terminal (WebSocket)
* **Safe file access** : scoped upload, download, and management
* **Auditability** : structured logs with execution metadata

Lightweight, self-hosted, and built for AI-native workflows. 

**Bastion** is not just a container runner, it’s a stateful execution layer for AI agents, with built-in safety, observability, and control.

 

## How It Works

### Sandboxed Runtime

* Each **session runs in an isolated Docker container**, providing strong separation from the host system 
* **Custom Docker images supported**, allowing you to tailor environments (Python, Node, ML stacks, etc.) for specific workflows 
* **Persistent + controlled runtime** : bind-mounted storage ensures files survive restarts, while CPU, memory, execution time, and network access are strictly governed per session 
* **Policy-driven secure execution** : enforce resource and filesystem restrictions, enabling safe, concurrent execution of untrusted or AI-generated code without risking the host system 
* Designed for **safe, concurrent execution of untrusted or AI-generated code** without risking your machine

Designed for **safe, concurrent execution of untrusted or AI-generated code**.

 

## Core Concepts

* **Session** : Persistent, isolated container environment with its own filesystem (`created → running → stopped → deleted`)
* **Jobs (Execution)** : Commands run via Docker `exec`; stateless, concurrent, async, with optional interactive TTY
* **Terminal** : WebSocket-based interactive shell with real-time bidirectional I/O
* **Filesystem** : Session-scoped, persistent storage with strict isolation (no path traversal)
 

## Observability & Security

* **Structured JSON logs** : command, output, duration, resource usage, policy decisions
* **API key authentication** + role-based access (`admin`, `agent`, `viewer`)
* **Container + filesystem isolation** with network restrictions

 

## Persistence & Recovery

* Session metadata stored in SQLite
* Automatic reconciliation with Docker on restart
* Handles orphan containers and state recovery

 

## Use Cases

* AI agents code execution
* Secure code interpreters
* Dev sandboxes & automation pipelines
* Research & data workflows
* Multi-agent systems


## Roadmap

* Policy engine (TOML)
* RBAC & multi-tenancy
* Distributed + Kubernetes-native runtime
* Session Replays & Agent tracing
* Snapshotting & prebuilds
* Container pooling
* Advanced observability (OpenTelemetry)
