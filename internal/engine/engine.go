package engine

import (
	"log/slog"

	"github.com/sakkshm/bastion/internal/config"
	"github.com/sakkshm/bastion/internal/docker"
	"github.com/sakkshm/bastion/internal/session"
)

type Engine struct {
	Sessions *session.SessionManager
	Docker   *docker.DockerClient
	Logger   *slog.Logger
	Config   *config.Config
}

func NewEngine(cfg *config.Config, logger *slog.Logger) (*Engine, error) {
	logger.Info("Initializing Docker Client")
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	// Prefetch container image at startup
	logger.Info("Pre-fetching container image")
	err = dockerClient.PrefetchImage(cfg.Sandbox.Image, logger)
	if err != nil {
		return nil, err
	}

	return &Engine{
		Sessions: session.NewSessionManager(),
		Docker:   dockerClient,
		Logger:   logger,
		Config:   cfg,
	}, nil
}
