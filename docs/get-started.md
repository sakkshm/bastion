## Getting Started with Bastion

This guide walks you through:

1. Building or downloading Bastion
2. Creating API keys
3. Configuring the system
4. Running the server
5. Verifying it works

## 1. Install / Build

### Option A: Use Prebuilt Binary

Download the latest release and place it in your project:

```bash
./bastion
```

### Option B: Build from Source

```bash
make build
```

This generates:

```
bin/bastion
```

 

## 2. Create API Key

Before running the server, generate an API key:

```bash
bin/bastion key create --name=Admin --scope=admin
```

### Other Key Commands

```bash
bin/bastion key list
bin/bastion key revoke --name=Admin
```

You’ll use this key for all API requests:

```
Authorization: Bearer <API_KEY>
```

## 3. Configuration

Create a file named:

```
config.toml
```

### Full Example

```toml
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

 

## 4. Config Explained

### `[server]`

Controls how Bastion exposes its API.

| Field  | Description                                         |
| ------ | --------------------------------------------------- |
| `host` | Address to bind server (`0.0.0.0` = all interfaces) |
| `port` | Port where API will be available                    |

 

### `[execution]`

Controls session lifecycle and concurrency.

| Field                          | Description                                |
| ------------------------------ | ------------------------------------------ |
| `mode`                         | Execution mode (`sandbox`)                 |
| `max_concurrent_sessions`      | Max active sessions allowed                |
| `session_ttl_minutes`          | Auto-expiry for inactive sessions          |
| `session_cleanup_interval_sec` | Cleanup frequency                          |
| `working_directory_base`       | Base directory for session files           |
| `env_path`                     | Path to `.env` file injected into sessions |


### `[sandbox]`

Defines container runtime behavior.

| Field             | Description                               |
| ----------------- | ----------------------------------------- |
| `enabled`         | Enable sandboxed containers               |
| `image`           | Docker image used for sessions            |
| `load_env`        | Load environment variables into container |
| `network_enabled` | Allow outbound network                    |
| `memory_mbs`      | Memory limit per session                  |
| `cpus`            | CPU allocation                            |
| `pids`            | Max processes                             |
| `job_ttl`         | Max execution time (seconds)              |


### `[filesystem]`

Controls file operations.

| Field                 | Description             |
| --------------------- | ----------------------- |
| `max_upload_size_mbs` | Max allowed upload size |



### `[logging]`

Controls logs and observability.

| Field    | Description                                  |
| -------- | -------------------------------------------- |
| `level`  | Log level (`debug`, `info`, `warn`, `error`) |
| `format` | `text` or `json`                             |
| `file`   | Log output file path                         |


## 5. Run Bastion

Start the server:

```bash
bin/bastion run --config="./config.toml"
```


## 6. Verify It’s Running

Bastion will expose the API at:

```
http://localhost:8080
```

Test health:

```bash
curl http://localhost:8080/health
```

Expected:

```
200 OK
```


## 7. First Session (Quick Test)

### Create Session

```bash
curl -X POST http://localhost:8080/session/create \
  -H "Authorization: Bearer <API_KEY>"
```


### Start Session

```bash
curl -X POST http://localhost:8080/session/<id>/start \
  -H "Authorization: Bearer <API_KEY>"
```

### Get Session Status

```bash
curl -X GET http://localhost:8080/session/<id>/status \
  -H "Authorization: Bearer <API_KEY>"
```


### Run Command

```bash
curl -X POST http://localhost:8080/session/<id>/exec \
  -H "Authorization: Bearer <API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"cmd": ["echo", "hello"]}'
```

> [`See full API docs here.`](./docs/get-started.md)

## 8. Notes

- Sessions must be **started** before executing commands
- All operations are **scoped per session**
- Jobs are **asynchronous**
- Files persist within session directory
- Deleting a session is **irreversible**
