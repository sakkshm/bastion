package engine

import (
	"context"
	"errors"
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

func (e *Engine) Close() error {
	return e.Docker.CloseClient()
}

func (e *Engine) AttachWorker(sess *session.Session) {

	// for v0.1, just one worker to prevent race conditions and stuff
	// TODO: in the future add multiple workers
	workerCount := 1

	for range workerCount {
		go func() {
			defer func() {
				// recover goroutine if worker panics
				if r := recover(); r != nil {
					e.Logger.Error("worker panic", "error", r)
				}
			}()

			for job := range sess.JobHandler.Queue {
				job.Status = session.JobRunning

				output, errout, exitCode, err := e.Docker.SessionRunJob(
					job.Context,
					sess.ContainerID,
					job.Cmd,
					job.JobID,
				)

				// cancel context after completion
				job.Cancel()

				job.Output.ConsoleOutput = output
				job.Output.ErrOut = errout

				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						job.Status = session.JobTimedout
					} else if errors.Is(err, context.Canceled) {
						job.Status = session.JobCanceled
					} else {
						job.Status = session.JobFailed
					}

					job.Output.ErrOut = err.Error()
					continue
				}
				if exitCode != 0 {
					job.Status = session.JobFailed
				} else {
					job.Status = session.JobCompleted
				}
			}
		}()
	}
}
