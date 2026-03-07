package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type CreateSessionResponse struct {
	SessionID string    `json:"session_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type DeleteSessionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
}

type GetSessionStatusResponse struct {
	ID          string    `json:"id"`
	ContainerID string    `json:"container_id"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	Status      string    `json:"status"`
}

type StartSessionResponse = GetSessionStatusResponse
type StopSessionResponse = GetSessionStatusResponse

type JobExecResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type JobStatusResponse struct {
	JobID     string    `json:"job_id"`
	Cmd       []string  `json:"cmd"`
	Status    string    `json:"status"`
	Output    string    `json:"output"`
	ErrOut    string    `json:"errout"`
	CreatedAt time.Time `json:"created_at"`
}

type APIError struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(APIError{
		Error: msg,
	})
}
