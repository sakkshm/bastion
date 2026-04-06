package docker

import (
	"github.com/docker/docker/client"
	"github.com/sakkshm/bastion/internal/filesystem"
)

type DockerClient struct {
	APIClient *client.Client
}

type ContainerConfig struct {
	Image          string
	Memory         int
	CPUs           float32
	PIDs           int
	NetworkEnabled bool
	FileSystem     filesystem.FSWorkspace
}
