package engine

import (
	"database/sql"
	"time"

	"github.com/sakkshm/bastion/internal/filesystem"
	"github.com/sakkshm/bastion/internal/session"
	"github.com/sakkshm/bastion/internal/websocket"
)

func (e *Engine) ReconcileAllSessions() error {
	rows, err := e.Sessions.DBConn.Database.Query(session.GetAllSessionsData)
	if err != nil {
		return err
	}
	defer rows.Close()

	var validSessions []session.Session
	var invalidSessionIDs []string

	for rows.Next() {
		s, invalidID, err := e.reconcileRow(rows)
		if err != nil {
			e.Logger.Error("failed to reconcile session", "err", err)
			continue
		}

		if invalidID != "" {
			invalidSessionIDs = append(invalidSessionIDs, invalidID)
			continue
		}

		if s != nil {
			validSessions = append(validSessions, *s)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if err := e.activateSessions(validSessions); err != nil {
		return err
	}

	if len(invalidSessionIDs) > 0 {
		e.Logger.Info("Cleaning up invalid sessions", "count", len(invalidSessionIDs))
		for _, id := range invalidSessionIDs {
			e.Sessions.DeleteSessionData(id)
		}
	}

	return nil
}

func (e *Engine) reconcileRow(rows *sql.Rows) (*session.Session, string, error) {
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
		return nil, "", err
	}

	ok, err := e.Docker.ContainerExists(containerID)
	if err != nil || !ok {
		e.Logger.Warn("dropping session: missing container",
			"session", id,
			"container", containerID,
			"err", err,
		)
		return nil, id, nil
	}

	ok, err = filesystem.SessionFSExist(*e.Config, id)
	if err != nil || !ok {
		e.Logger.Warn("dropping session: missing filesystem",
			"session", id,
			"err", err,
		)
		return nil, id, nil
	}

	s, err := e.rebuildSession(id, containerID, createdAt, lastUsedAt, status)
	if err != nil {
		return nil, id, err
	}

	return s, "", nil
}

func (e *Engine) rebuildSession(
	id string,
	containerID string,
	createdAt int64,
	lastUsedAt int64,
	status string,
) (*session.Session, error) {

	jobHandler := session.NewJobHandler()
	wsManager := websocket.NewWSManager(id, e.Sessions.Touch)

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

	if err := e.Sessions.BatchAdd(sessions); err != nil {
		return err
	}

	for i := range sessions {
		e.Logger.Info("Attaching worker to session", "session", sessions[i].ID)
		e.AttachWorker(&sessions[i])
	}

	return nil
}
