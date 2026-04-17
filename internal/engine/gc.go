package engine

import (
	"context"
	"time"

	"github.com/sakkshm/bastion/internal/session"
)

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

	// delete session metadata from sessionManager and DB
	e.Sessions.BatchDelete(toDelete)

	for _, session := range toDeleteSessions {
		// kill all ws clients
		session.WSManager.Cancel()

		err := e.Docker.DeleteContainer(context.Background(), session.ContainerID)
		if err != nil {
			e.Logger.Error("Unable to delete docker container",
				"session_id", session.ID,
				"container_id", session.ContainerID,
			)
		}

		err = session.FileSystem.DeleteWorkspace()
		if err != nil {
			e.Logger.Error("Unable to delete fs workspace",
				"session_id", session.ID,
			)
		}
	}
}
