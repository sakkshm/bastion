package engine

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/sakkshm/bastion/internal/config"
	"github.com/sakkshm/bastion/internal/database"
	"github.com/sakkshm/bastion/internal/docker"
	"github.com/sakkshm/bastion/internal/session"
)

type Engine struct {
	Sessions *session.SessionManager
	Docker   *docker.DockerClient
	Database *database.DatabaseConn
	Logger   *slog.Logger
	Config   *config.Config
}

func NewEngine(cfg *config.Config, logger *slog.Logger) (*Engine, error) {
	logger.Info("Initializing Docker Client")
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}

	logger.Info("Pre-fetching container image")
	err = dockerClient.PrefetchImage(cfg.Sandbox.Image, logger)
	if err != nil {
		return nil, err
	}

	logger.Info("Initializing DB Connection")
	dbConn, err := database.NewDBConn()
	if err != nil {
		return nil, err
	}

	sm, err := session.NewSessionManager(dbConn)
	if err != nil {
		return nil, err
	}

	e := Engine{
		Sessions: sm,
		Docker:   dockerClient,
		Database: dbConn,
		Logger:   logger,
		Config:   cfg,
	}

	// reconcile state
	_ = e.ReconcileAllSessions()

	e.StartSessionGarbageCollector(
		time.Duration(cfg.Execution.SessionCleanupIntervalSec)*time.Second,
		time.Duration(cfg.Execution.SessionTTLMinutes)*time.Minute,
	)

	return &e, nil
}

func (e *Engine) Close() error {
	if err := e.Docker.CloseClient(); err != nil {
		return err
	}
	if err := e.Database.Close(); err != nil {
		return err
	}

	return nil
}

func (e *Engine) AttachWorker(sess *session.Session) {
	workerCount := 1

	for i := 0; i < workerCount; i++ {
		go func() {
			defer func() {
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
