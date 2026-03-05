package docker

import (
	"github.com/docker/docker/client"
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
}
