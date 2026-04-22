# Bastion API Reference

Base URL:

```text
http://localhost:8080
```

 

## Authentication

All endpoints (except `/health`) require:

```http
Authorization: Bearer <API_KEY>
```

 

## API Overview

Bastion exposes a REST + WebSocket API organized into:

* **Sessions** → lifecycle of environments
* **Jobs** → command execution
* **Terminal** → real-time shell
* **Files** → filesystem operations
* **System** → health checks

 

# 1. System

## Health Check

```http
GET /health
```

No authentication required.

 

# 2. Sessions

Sessions represent isolated execution environments.

 

## Create Session

```http
POST /session/create
```

**Response (201)**

```json
{
  "session_id": "abc123",
  "status": "created",
  "created_at": "2026-04-22T12:00:00Z"
}
```

**Errors**

* `429` → Max sessions reached
* `500` → Internal error

 

## Start Session

```http
POST /session/{id}/start
```

**Response (200)**

```json
{
  "id": "abc123",
  "container_id": "docker_id",
  "status": "running",
  "created_at": "...",
  "last_used_at": "..."
}
```

 

## Stop Session

```http
POST /session/{id}/stop
```

 

## Get Session Status

```http
GET /session/{id}/status
```

 

## Delete Session

```http
DELETE /session/{id}
```

**Response**

```json
{
  "session_id": "abc123",
  "status": "deleted"
}
```

 

## Session States

```text
created → running → stopped → deleted
                 ↘ busy
```

 

# 3. Jobs (Command Execution)

Each command runs as an **asynchronous job**.

 

## Execute Command

```http
POST /session/{id}/exec
```

**Request**

```json
{
  "cmd": ["ls", "-la"]
}
```

**Response**

```json
{
  "job_id": "job_123",
  "status": "queued"
}
```

**Errors**

* `400` → Invalid request
* `403` → Session not running

 

## Get Job Status

```http
GET /session/{id}/job/{job_id}
```

**Response**

```json
{
  "job_id": "job_123",
  "cmd": ["ls", "-la"],
  "status": "completed",
  "created_at": "...",
  "output": {
    "console_output": "file.txt",
    "errout": "",
    "status_code": 0
  }
}
```

 

## Job States

```text
queued → running → completed | failed
```

 

Here’s your section rewritten in the **same format**, but with correct `client_id` flow included:

---

# 4. Terminal (WebSocket)

Interactive shell via WebSocket.

---

## Connect

```http
GET /session/{id}/terminal
```

Upgrades to WebSocket (`101 Switching Protocols`).


## Message Format

### Init (Server → Client)

First message received after connection:

```json
{
  "type": "init",
  "client_id": "78d2c05d",
  "payload": {
    "msg": "connected"
  }
}
```

> Save `client_id` — it must be sent in all future messages.

### Send Input

```json
{
  "type": "term_input",
  "client_id": "78d2c05d",
  "session_id": "5e7fa513",
  "payload": {
    "input": "ls -la\n"
  }
}
```

### Receive Output

```json
{
  "type": "terminal_output",
  "client_id": "78d2c05d",
  "session_id": "5e7fa513",
  "payload": {
    "output": "file.txt"
  }
}
```


## Notes

* Requires session in `running` state
* First message is always `init` (contains `client_id`)
* `client_id` must be included in all subsequent messages
* Fully interactive (TTY)
* Bidirectional communication

 

# 5. File Operations

All file operations are scoped per session.

 

## Upload File

```http
POST /session/{id}/upload
Content-Type: multipart/form-data
```

**Fields**

* `file` → binary file
* `metadata.path` → destination path

 

## Download File

```http
GET /session/{id}/download?path=/file.txt
```

Returns raw file stream.

 

## Delete File

```http
DELETE /session/{id}/delete
```

**Request**

```json
{
  "path": "/file.txt"
}
```

 

## List Files

```http
GET /session/{id}/list?path=/&page=1&limit=10
```

**Response**

```json
{
  "page": 1,
  "limit": 10,
  "total": 20,
  "total_pages": 2,
  "files": [
    {
      "name": "file.txt",
      "is_dir": false,
      "size": 123,
      "mode": "-rw-r--r--",
      "mod_time": "..."
    }
  ]
}
```

 

# 6. Data Models

 

## SessionStatus

```text
created | running | stopped | deleted | busy
```

 

## JobStatus

```text
queued | running | completed | failed
```

 

## Job Output

```json
{
  "console_output": "string",
  "errout": "string",
  "status_code": 0
}
```

 

## Error Format

```json
{
  "error": "message"
}
```

 

# 7. Typical Workflow

### 1. Create session

```http
POST /session/create
```

 

### 2. Start session

```http
POST /session/{id}/start
```

 

### 3. Execute command

```http
POST /session/{id}/exec
```

 

### 4. Poll result

```http
GET /session/{id}/job/{job_id}
```

 

### 5. Cleanup

```http
POST   /session/{id}/stop
DELETE /session/{id}
```

 

# 8. Python SDK (Recommended)

Instead of raw API calls:

```bash
pip install bastion-py-sdk
```

### Example

```python
from bastion import Bastion

with Bastion(base_url="http://localhost:8080") as b:
    session_id = b.sessions.create()["session_id"]
    b.sessions.start(session_id)

    result = b.jobs.run_and_wait(session_id, ["echo", "hello"])
    print(result["output"]["console_output"])
```

> [`See full SDK docs here.`](./py-sdk-docs.md)
 

# 9. Important Constraints

* Session must be **running** before executing jobs
* Jobs are **asynchronous by default**
* File system is **isolated per session**
* Terminal requires active session
* Deleting a session is **irreversible**

