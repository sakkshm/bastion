package engine

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/sakkshm/bastion/internal/config"
	"github.com/sakkshm/bastion/internal/database"
	"github.com/sakkshm/bastion/internal/docker"
	"github.com/sakkshm/bastion/internal/filesystem"
	"github.com/sakkshm/bastion/internal/session"
	"github.com/sakkshm/bastion/internal/websocket"
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

	// Prefetch container image at startup
	logger.Info("Pre-fetching container image")
	err = dockerClient.PrefetchImage(cfg.Sandbox.Image, logger)
	if err != nil {
		return nil, err
	}

	// initializing DB
	logger.Info("Initializing DB Connection")
	dbConn, err := database.NewDBConn()
	if err != nil {
		return nil, err
	}

	// make session manager
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
	e.ReconcileAllSessions()

	e.StartSessionGarbageCollector(
		time.Duration(cfg.Execution.SessionCleanupIntervalSec)*time.Second,
		time.Duration(cfg.Execution.SessionTTLMinutes)*time.Minute,
	)

	return &e, nil
}

func (e *Engine) Close() error {
	err := e.Docker.CloseClient()
	if err != nil {
		return err
	}

	err = e.Database.Close()
	if err != nil {
		return err
	}

	return nil
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
	var toDeleteSessions []*session.Session

	for id, session := range snapshot {
		if session.IsExpired(ttl) {
			toDelete = append(toDelete, id)
			toDeleteSessions = append(toDeleteSessions, session)
		}
	}

	// delete from SessionManager
	e.Sessions.BatchDelete(toDelete)

	// delete containers related data
	for _, session := range toDeleteSessions {

		// delete docker container
		err := e.Docker.DeleteContainer(context.Background(), session.ContainerID)
		if err != nil {
			e.Logger.Error(
				"Unable to delete docker conatiner for a session",
				"session_id", session.ID,
				"conatiner_id", session.ContainerID,
			)
		}

		// disconnect all clients
		session.WSManager.Cancel()

		// delete fs workspace
		err = session.FileSystem.DeleteWorkspace()
		if err != nil {
			e.Logger.Error(
				"Unable to delete fs workspace for a session",
				"session_id", session.ID,
			)
		}
	}

}

func (e *Engine) ReconcileAllSessions() error {
	rows, err := e.Sessions.DBConn.Database.Query(session.GetAllSessionsData)
	if err != nil {
		return err
	}
	defer rows.Close()

	var validSessions []session.Session

	for rows.Next() {
		s, err := e.reconcileRow(rows)
		if err != nil {
			e.Logger.Error("failed to reconcile session", "err", err)
			continue
		}
		if s != nil {
			validSessions = append(validSessions, *s)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return e.activateSessions(validSessions)
}

func (e *Engine) reconcileRow(rows *sql.Rows) (*session.Session, error) {
	var id string
	var containerID string
	var createdAt int64
	var lastUsedAt int64
	var status string
	var fsMount string

	err := rows.Scan(
		&id,
		&containerID,
		&createdAt,
		&lastUsedAt,
		&status,
		&fsMount,
	)
	if err != nil {
		return nil, err
	}

	// validate container
	ok, err := e.Docker.ContainerExists(containerID)
	if err != nil || !ok {
		e.Logger.Warn("dropping session: missing container",
			"session", id,
			"container", containerID,
			"err", err,
		)
		return nil, nil
	}

	// validate filesystem
	ok, err = filesystem.SessionFSExist(*e.Config, id)
	if err != nil || !ok {
		e.Logger.Warn("dropping session: missing filesystem",
			"session", id,
			"err", err,
		)
		return nil, nil
	}

	// rebuild session
	return e.rebuildSession(id, containerID, createdAt, lastUsedAt, status)
}

func (e *Engine) rebuildSession(
	id string,
	containerID string,
	createdAt int64,
	lastUsedAt int64,
	status string,
) (*session.Session, error) {

	jobHandler := session.NewJobHandler()
	wsManager := websocket.NewWSManager(id)

	go wsManager.Run()

	fs, err := filesystem.NewFSWorkspace(*e.Config, id)
	if err != nil {
		return nil, err
	}

	return &session.Session{
		ID:          id,
		ContainerID: containerID,
		CreatedAt:   time.Unix(createdAt, 0).UTC(),
		LastUsedAt:  time.Unix(lastUsedAt, 0).UTC(),
		Status:      session.ParseStatus(status),
		JobHandler:  jobHandler,
		WSManager:   wsManager,
		FileSystem:  fs,
	}, nil
}

func (e *Engine) activateSessions(sessions []session.Session) error {
	e.Logger.Info("Activating sessions", "count", len(sessions))

	// batch insert into memory manager
	if err := e.Sessions.BatchAdd(sessions); err != nil {
		return err
	}

	for i := range sessions {
		e.Logger.Info("Attaching worker to session", "session", sessions[i].ID)
		e.AttachWorker(&sessions[i])
	}

	return nil
}
