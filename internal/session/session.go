package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sakkshm/bastion/internal/websocket"
)

type Session struct {
	ID          string
	ContainerID string
	CreatedAt   time.Time
	LastUsedAt  time.Time
	Status      Status
	JobHandler  *JobHandler
	WSManager   *websocket.WSManager
}

func (s Session) IsExpired(ttl time.Duration) bool {
	return time.Since(s.LastUsedAt) > ttl || s.Status == StatusDeleted
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

type ExecJob struct {
	JobID     string
	Cmd       []string
	Status    JobStatus
	Output    ExecJobOutput
	Context   context.Context
	Cancel    context.CancelFunc
	CreatedAt time.Time
}

type ExecJobOutput struct {
	ConsoleOutput string
	ErrOut        string
	StatusCode    int
}

type JobStatus int

const (
	JobQueued JobStatus = iota
	JobRunning
	JobCompleted
	JobFailed
	JobTimedout
	JobCanceled
)

func (s JobStatus) String() string {
	switch s {
	case JobQueued:
		return "queued"
	case JobRunning:
		return "running"
	case JobCompleted:
		return "completed"
	case JobFailed:
		return "failed"
	case JobTimedout:
		return "timeout"
	case JobCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

const SESSION_ID_LEN = 8
const JOB_ID_LEN = 6

func GenerateSessionID() string {
	id := uuid.New()
	return id.String()[:SESSION_ID_LEN]
}

func GenerateJobID() string {
	id := uuid.New()
	return id.String()[:JOB_ID_LEN]
}
