package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/sakkshm/bastion/internal/session"
)

var STOP_TIMEOUT int = 3

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
		Image: cfg.Image,
		Tty:   true,                // Allocate Terminal
		Cmd:   []string{"/bin/sh"}, // Deafault Shell
	}

	// Host config
	memory_mbs := int64(cfg.Memory * 1024 * 1024)
	cpu_cores := cfg.CPUs * 1_000_000_000
	pid_limits := int64(cfg.PIDs)

	hostConfig := &container.HostConfig{
		// AutoRemove:     true,                          // automatically remove when stopped
		ReadonlyRootfs: true,                          // cannot change anything in root fs
		SecurityOpt:    []string{"no-new-privileges"}, // deny privilege escalation
		CapDrop:        []string{"ALL"},               // drop linux
		Resources: container.Resources{
			Memory:    memory_mbs,
			NanoCPUs:  int64(cpu_cores),
			PidsLimit: &pid_limits,
		},
		NetworkMode: "none", // no network
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeTmpfs,
				Target: "/sandbox", // ephemeral storage
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
