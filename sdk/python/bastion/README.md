# Bastion Python SDK

## Overview

The Bastion Python SDK provides a high-level interface for working with isolated execution environments (“sessions”). It enables you to run commands, manage files, and interact with a live terminal inside a sandboxed runtime.

Each workflow is scoped to a **session**, which acts as an isolated container.

> [Github Repo](https://github.com/sakkshm/bastion)

## Core Concepts

* **Session**:
  An isolated runtime environment where all operations are executed.

* **Job**:
  A command executed inside a session. Jobs are asynchronous by default.

* **Files**:
  Files stored within a session’s filesystem.

* **Terminal**:
  A real-time interactive shell connected via WebSocket.



## Installation

```bash
pip install bastion-py-sdk
```



## Quickstart

```python
from bastion import Bastion

bastion = Bastion(
    base_url="bastion_url",
    api_key="your_api_key",
)

# Create and start a session
session = bastion.sessions.create()
session_id = session["session_id"]
bastion.sessions.start(session_id)

# Run a command
result = bastion.jobs.run_and_wait(session_id, ["echo", "hello"])
print(result["output"]["console_output"])

bastion.close()
```



## Sessions

Sessions represent isolated environments. All operations like jobs, files and terminal are scoped to a session.

### Create a Session

```python
session = bastion.sessions.create()
session_id = session["session_id"]
```

### Start a Session

```python
bastion.sessions.start(session_id)
```

### Get Session Status

```python
status = bastion.sessions.status(session_id)
print(status["status"])
```

Possible states:

* `created`
* `running`
* `stopped`
* `deleted`

### Stop a Session

```python
bastion.sessions.stop(session_id)
```

### Delete a Session

```python
bastion.sessions.delete(session_id)
```



## Jobs (Command Execution)

Jobs are used to execute commands inside a session.

### Run and Wait (Synchronous)

```python
result = bastion.jobs.run_and_wait(session_id, ["ls", "-la"])
print(result["output"]["console_output"])
```

### Run Asynchronously

```python
job = bastion.jobs.run(session_id, ["echo", "hello"])
job_id = job["job_id"]
```

### Get Job Status

```python
status = bastion.jobs.get(session_id, job_id)
print(status["status"])
```

### Wait for Completion

```python
result = bastion.jobs.wait(session_id, job_id, timeout=10)
```

* Raises `TimeoutError` if the job exceeds the timeout
* Raises `JobFailedError` if the job fails

### Stream Job Updates

```python
def on_update(update):
    print(update["status"])

bastion.jobs.watch(session_id, job_id, on_update)
```



## Job Result Format

```python
{
    "job_id": str,
    "status": "queued" | "running" | "completed" | "failed",
    "output": {
        "console_output": str,
        "errout": str,
        "status_code": int
    } | None
}
```



## Files

File operations are scoped to a session.

### Upload a File

```python
with open("data.txt", "rb") as f:
    bastion.files.upload(session_id, f, "/data.txt")
```

### List Files

```python
files = bastion.files.list(session_id, "/")

for f in files["files"]:
    print(f["name"], f["size"])
```

### Delete a File

```python
bastion.files.delete(session_id, "/data.txt")
```



## Terminal

The terminal provides real-time interaction with a session via WebSocket.

### Connect

```python
def on_message(msg):
    print(msg)

bastion.terminal.connect(
    session_id=session_id,
    on_message=on_message,
)
```

### Send Input

```python
bastion.terminal.send_input("ls -la\n")
```

### Execute Command

```python
bastion.terminal.exec("echo hello")
```

### Close Connection

```python
bastion.terminal.close()
```



## Terminal Message Format

Messages received via `on_message`:

```python
{
    "type": "init" | "term_output",
    "payload": dict
}
```



## Error Handling

All exceptions inherit from `BastionError`.

### Common Exceptions

* `SessionError`
* `SessionStateError`
* `JobError`
* `JobFailedError`
* `FileUploadError`
* `FileListError`
* `FileDeleteError`
* `TerminalConnectionError`
* `TerminalSendError`

### Example

```python
from bastion.exceptions import SessionError

try:
    bastion.sessions.start("invalid-id")
except SessionError as e:
    print(e)
```



## Resource Management

Close the client when finished:

```python
bastion.close()
```

Or use a context manager:

```python
from bastion import Bastion

with Bastion(base_url="http://localhost:8080") as bastion:
    session_id = bastion.sessions.create()["session_id"]
```



## Constraints

* Sessions must be in `running` state before executing jobs or accessing files
* Terminal requires an active session
* Jobs are asynchronous unless explicitly awaited
* File operations are isolated per session
* Deleting a session is irreversible



## Complete Example

```python
from bastion import Bastion

bastion = Bastion(
    base_url="bastion_url",
    api_key="your_api_key",
)

session_id = bastion.sessions.create()["session_id"]
bastion.sessions.start(session_id)

bastion.files.upload(session_id, b'print("Hello from Bastion")', "/main.py")

result = bastion.jobs.run_and_wait(session_id, ["python3", "/main.py"])
print(result["output"]["console_output"])

bastion.sessions.delete(session_id)
bastion.close()
```



## Summary

The Bastion SDK provides:

* Isolated execution via sessions
* Async and synchronous command execution
* File system access per session
* Real-time terminal interaction

It is designed to offer a simple, consistent interface for sandboxed computation workflows.
