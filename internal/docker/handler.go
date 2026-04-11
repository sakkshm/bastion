package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sakkshm/bastion/internal/session"
)

func NewDockerClient() (*DockerClient, error) {
	apiClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("Error in initializing Docker Client: %w", err)
	}

	return &DockerClient{
		APIClient: apiClient,
	}, nil
}

func (d *DockerClient) CloseClient() error {
	return d.APIClient.Close()
}

func (d *DockerClient) ContainerExists(containerID string) (bool, error) {
	if containerID == "" {
		return false, errors.New("containerID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := d.APIClient.ContainerInspect(ctx, containerID)
	if err == nil {
		return true, nil
	}

	if errdefs.IsNotFound(err) {
		return false, nil
	}

	return false, err
}

func (d *DockerClient) PrefetchImage(imageName string, logger *slog.Logger) error {
	// Pull Image at startup
	reader, err := d.APIClient.ImagePull(
		context.Background(),
		imageName,
		image.PullOptions{},
	)
	if err != nil {
		return err
	}
	defer reader.Close()

	decoder := json.NewDecoder(reader)

	for {
		var msg jsonmessage.JSONMessage

		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// handle docker errors embedded in stream
		if msg.Error != nil {
			return fmt.Errorf("docker pull error: %s", msg.Error.Message)
		}

		logger.Info(
			"docker pull",
			"id", msg.ID,
			"status", msg.Status,
			"progress", msg.ProgressMessage,
		)
	}

	return nil
}

func (d *DockerClient) CreateSandboxContainer(ctx context.Context, cfg ContainerConfig, sessionID string) (string, error) {
	// Container config
	config := &container.Config{
		Image:        cfg.Image,
		Tty:          true, // Allocate Terminal
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		StdinOnce:    false,                         // keep stdin open for terminal
		Cmd:          []string{"sleep", "infinity"}, // Deafault Shell
		WorkingDir:   "/workspace",
		User:         "1000:1000",
		Labels: map[string]string{
			"session_id": sessionID,
		},
	}

	// Host config
	memory_mbs := int64(cfg.Memory * 1024 * 1024)
	cpu_cores := int64(cfg.CPUs * 1_000_000_000)
	pid_limits := int64(cfg.PIDs)
	init_allowed := true

	// check if fs exists
	if !cfg.FileSystem.FSExists() {
		return "", fmt.Errorf("workspace missing")
	}

	hostConfig := &container.HostConfig{
		// AutoRemove:     true,                       // automatically remove when stopped
		ReadonlyRootfs: true,                          // cannot change anything in root fs
		SecurityOpt:    []string{"no-new-privileges"}, // deny privilege escalation
		CapDrop:        []string{"ALL"},               // drop all linux capabilitites
		Resources: container.Resources{
			Memory:    memory_mbs,
			NanoCPUs:  cpu_cores,
			PidsLimit: &pid_limits,
		},
		NetworkMode: "none", // no network
		Init:        &init_allowed,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: cfg.FileSystem.Mount,
				Target: "/workspace",
				BindOptions: &mount.BindOptions{
					Propagation: mount.PropagationRPrivate,
				},
			},
			{
				Type:   mount.TypeTmpfs,
				Target: "/tmp",
				TmpfsOptions: &mount.TmpfsOptions{
					SizeBytes: 64 * 1024 * 1024,
					Mode:      0700,
				},
			},
		},
	}

	// Create Container
	resp, err := d.APIClient.ContainerCreate(
		ctx,
		config,
		hostConfig,
		&network.NetworkingConfig{},
		nil,
		fmt.Sprintf("session-%s", sessionID),
	)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *DockerClient) StartContainer(ctx context.Context, containerID string) error {
	return d.APIClient.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (d *DockerClient) StopContainer(ctx context.Context, containerID string) error {

	var STOP_TIMEOUT int = 0

	return d.APIClient.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &STOP_TIMEOUT,
	})
}

func (d *DockerClient) DeleteContainer(ctx context.Context, containerID string) error {
	return d.APIClient.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}

func (d *DockerClient) GetContainerStatus(ctx context.Context, containerID string) (session.Status, error) {

	info, err := d.APIClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return session.StatusFailed, err
	}

	if info.State == nil {
		return session.StatusFailed, nil
	}

	switch {
	case info.State.Running:
		return session.StatusRunning, nil

	case info.State.Restarting:
		return session.StatusStarting, nil

	case info.State.Paused:
		return session.StatusBusy, nil

	case info.State.Dead:
		return session.StatusFailed, nil

	default:
		return session.StatusStopped, nil
	}
}

func (d *DockerClient) SessionRunJob(ctx context.Context, containerID string, cmd []string, jobID string) (string, string, int, error) {

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		Privileged:   false,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  false,
		Tty:          false,
	}

	resp, err := d.APIClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", "", -1, err
	}

	attach, err := d.APIClient.ContainerExecAttach(ctx, resp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", "", -1, err
	}
	defer attach.Close()

	var stdout, stderr bytes.Buffer
	limited := io.LimitReader(attach.Reader, 10<<20) // 10MB

	done := make(chan error, 1)

	go func() {
		_, err := stdcopy.StdCopy(&stdout, &stderr, limited)
		done <- err
	}()

	select {

	case err := <-done:
		if err != nil {
			return "", "", -1, err
		}

		inspect, err := d.APIClient.ContainerExecInspect(ctx, resp.ID)
		if err != nil {
			return stdout.String(), stderr.String(), -1, err
		}

		if inspect.Running {
			return stdout.String(), stderr.String(), -1, errors.New("exec stream closed early")
		}

	case <-ctx.Done():
		// just stop reading and return
		attach.Close()
		<-done

		return stdout.String(), stderr.String(), -1, ctx.Err()
	}

	inspect, err := d.APIClient.ContainerExecInspect(ctx, resp.ID)
	if err != nil {
		return stdout.String(), stderr.String(), -1, err
	}

	return stdout.String(), stderr.String(), inspect.ExitCode, nil
}

func (d *DockerClient) StartTerminalSession(ctx context.Context, containerID string) (types.HijackedResponse, error) {

	// start a shell process
	execResp, err := d.APIClient.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          []string{"/bin/sh", "-i"},
		Privileged:   false,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	})
	if err != nil {
		return types.HijackedResponse{}, err
	}

	// attach to container with TTY
	resp, err := d.APIClient.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return types.HijackedResponse{}, err
	}

	return resp, err
}
