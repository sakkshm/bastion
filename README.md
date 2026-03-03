# Bation

**A self-hosted, safe terminal environment for AI agents and automation tools.**

Run commands, manage files, and execute code in a controlled sandboxed environment **without risking your host machine**, accessible via a simple REST + WebSocket API.

## Why Bation?

AI models can write scripts and automate workflows, but running them directly on your system is risky. Bation solves this by providing:

* **Sandboxed execution** — containers isolate AI processes from your host.
* **Policy enforcement** — define allowed commands, file access, and network usage.
* **Resource controls** — limit CPU, memory, and execution time.
* **Live streaming** — watch stdout/stderr in real-time.
* **Scoped file management** — safely upload, download, and organize files.
* **Audit logging** — keep detailed execution history for debugging and reproducibility.

Bation is self-contained, lightweight, and designed to plug into AI workflows seamlessly.

## How It Runs

## Sandboxed

* Each session runs in its own ephemeral Docker container.
* CPU, memory, and network can be limited per session.
* Comes pre-installed with Python, Node.js, git, build tools, data libraries, ffmpeg, and other common utilities.
* Ideal for running AI agents safely without affecting the host system.

### Bare-Metal

* Runs directly on your machine.
* Full access to system files and installed tools.
* Suitable for personal automation, local development, or giving an AI full project access.
* Explicit warnings are shown before execution for safety.

## Quick Start

### Docker (Recommended)

```bash
docker run -d --name bation \
  -p 8000:8000 \
  -v bation-workspace:/workspace \
  -e BATION_API_KEY=your-key \
  bation/runtime
```

Visit `http://localhost:8000`, your terminal is ready.

> Tip: If you don’t provide an API key, Bation generates one automatically. Retrieve it with `docker logs bation`.

### Bare-Metal

```bash
bation run --host 0.0.0.0 --port 8000 --api-key your-key
```

> Caution: Bare-metal mode runs commands with your user permissions. Use Docker for safer execution.

## Integrating With AI Agents

Bation is designed to be API-first. AI agents can:

* Run commands via `POST /execute`
* Stream output over `/ws/{session_id}`
* Upload/download files in `/workspace`
* Validate commands before execution with the policy simulation endpoint

Integration can be **direct** (AI talks to your Bation instance) or **proxied** (through a central server for multi-user setups).
