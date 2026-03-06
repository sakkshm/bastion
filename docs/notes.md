.
├── cmd
│   └── bastion
│       └── main.go
├── internal
│   ├── docker
│   │   ├── docker.go
│   │   └── handler.go
├── Makefile

--- 

# Notes

## Main Process:
- Can set config from flags
- First load the config

## logger

- Uses slog
    - Opens a file to store logs (specified in config)
    - Uses Multiwriter to write to Console + File
    - Format is json or text

- SimpleHandler is a custom slog.Handler that writes log messages to a io.Writer in a text format.

## engine

- It is a central store for Sessions map, Docker client, Logger, Config. Used to inject dependecies to a method.
- During init:
    - It creates Docker Client (from ENV, does API version negotiation).
    - Prefetches the container image at startup. (Streams logs to logger)
    - Returns engine with all the deps.

## sessions
- Basically a way to manage state for all conatiners.
- A session can have the following states:
	- StatusCreated
	- StatusStarting
	- StatusRunning
	- StatusBusy
	- StatusStopped
	- StatusFailed
	- StatusDeleted
- SessionsManager is a map to manage all the relations.


## api
- A handler exists to basically inject the Engine into the api context.
- A simple middleware to extract ids from URLs and add session data to r.Context
- Requests/Responses contain models for req/res
- Routes contain route patterns for endpoints
- There are the following handlers:
    - CreateNewSession: Checks if max concurrent sessions are reached, generates a session ID, creates a ContainerConfig, calls Docker.CreateSandboxContainer, adds the session to memory, and returns CreateSessionResponse with SessionID, Status, and CreatedAt.
    - StartSessionHandler: Retrieves session from context, touches it, starts the container if not already running, updates session status to running, and returns StartSessionResponse with session details.
    - StopSessionHandler: Retrieves session from context, touches it, stops the container if not already stopped, updates session status to stopped, and returns StopSessionResponse with session details.
    - DeleteSessionHandler: Retrieves session from context, deletes the container if not already deleted, updates session status to deleted, and returns DeleteSessionResponse with SessionID and Status.
    - GetSessionStatusHandler: Retrieves session from context, touches it, inspects the container if not deleted, syncs Docker container state with session state if needed, and returns GetSessionStatusResponse with session details.

## docker

- A client exists to wrap the Docker Go SDK (client.Client) and provide utility methods for sandboxed container management.
- It supports creating, starting, stopping, deleting containers, prefetching images, and checking container status.
- There are the following methods/handlers:

  - NewDockerClient: Initializes a Docker API client from environment and returns a DockerClient.
  - CloseClient: Closes the underlying Docker API client.
  - PrefetchImage: Pulls a Docker image, logs progress and errors, ensures the image is ready before container creation.
  - CreateSandboxContainer: Builds container and host configs (memory, CPU, PID limits, security, network isolation, ephemeral storage), creates a sandbox container for a session, and returns the container ID.
  - StartContainer: Starts a container given its ID.
  - StopContainer: Stops a container gracefully with a timeout.
  - DeleteContainer: Removes a container forcefully.
  - GetContainerStatus: Inspects a container and returns a corresponding session.Status based on its Docker state (running, restarting, paused, stopped, dead).
