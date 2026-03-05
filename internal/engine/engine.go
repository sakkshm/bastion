package engine

import (
	"github.com/sakkshm/bastion/internal/docker"
	"github.com/sakkshm/bastion/internal/session"
)

type Engine struct {
	Sessions *session.SessionManager
	Docker   *docker.DockerClient
}

func NewEngine() (*Engine, error) {
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	return &Engine{
		Sessions: session.NewSessionManager(),
		Docker:   dockerClient,
	}, nil
}
