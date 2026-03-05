package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID          string
	ContainerID string
	CreatedAt   time.Time
	LastUsedAt  time.Time
	Status      string
}

type Status int

const (
	StatusCreated Status = iota
	StatusStarting
	StatusRunning
	StatusBusy
	StatusStopped
	StatusFailed
	StatusDeleted
)

func (s Status) String() string {
	switch s {
	case StatusCreated:
		return "created"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusBusy:
		return "busy"
	case StatusStopped:
		return "stopped"
	case StatusFailed:
		return "failed"
	case StatusDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

func GenerateSessionID() string {
	id := uuid.New()
	return id.String()[:SESSION_ID_LEN]
}
