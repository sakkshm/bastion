package engine

import (
	"context"
	"errors"
	"log/slog"
	"time"

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

	e := Engine{
		Sessions: session.NewSessionManager(),
		Docker:   dockerClient,
		Logger:   logger,
		Config:   cfg,
	}

	e.StartSessionGarbageCollector(
		time.Duration(cfg.Execution.SessionCleanupIntervalSec)*time.Second,
		time.Duration(cfg.Execution.SessionTTLMinutes)*time.Minute,
	)

	return &e, nil
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

func (e *Engine) StartSessionGarbageCollector(interval time.Duration, ttl time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			e.cleanupSessions(ttl)
		}
	}()
}

func (e *Engine) cleanupSessions(ttl time.Duration) {
	snapshot := e.Sessions.Snapshot()
	var toDelete []string
	var toDeleteContainer []string

	for id, session := range snapshot {
		if session.IsExpired(ttl) {
			toDelete = append(toDelete, id)
			toDeleteContainer = append(toDeleteContainer, session.ContainerID)
		}
	}

	// delete from SessionManager
	e.Sessions.BatchDelete(toDelete)

	// delete conatiners from Docker daemon
	for _, containerID := range toDeleteContainer {
		_ = e.Docker.DeleteContainer(context.Background(), containerID)
	}

}
