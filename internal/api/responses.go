package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type CreateSessionResponse struct {
	SessionID string `json:"session_id"`
}

type GetSessionStatusResponse struct {
	ID          string    `json:"id"`
	ContainerID string    `json:"container_id"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	Status      string    `json:"status"`
}

type StartSessionResponse = GetSessionStatusResponse

type APIError struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(APIError{
		Error: msg,
	})
}
