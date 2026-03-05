package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sakkshm/bastion/internal/docker"
	"github.com/sakkshm/bastion/internal/session"
)

func (h *Handler) CreateNewSession(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if h.Engine.Sessions.Count() >= h.Config.Execution.MaxConcurrent {
		h.Logger.Error("Maximum sessions reached", "error", "max_sessions_reached")
		writeJSONError(w, http.StatusTooManyRequests, "Maximum sessions reached")
		return
	}

	// generate session ID
	sessionID := session.GenerateSessionID()

	// create a container
	containerConfig := docker.ContainerConfig{
		Image:          h.Config.Sandbox.Image,
		Memory:         h.Config.Sandbox.Memory,
		CPUs:           h.Config.Sandbox.CPUs,
		PIDs:           h.Config.Sandbox.PIDs,
		NetworkEnabled: h.Config.Sandbox.NetworkEnabled,
	}

	containerID, err := h.Engine.Docker.CreateSandboxContainer(
		r.Context(),
		containerConfig,
		sessionID,
	)

	if err != nil {
		h.Logger.Error(
			"Failed to create sandbox container",
			"session_id", sessionID,
			"error", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "Failed to create sandbox container")
		return
	}

	// create a session entry
	now := time.Now().UTC()

	sess := session.Session{
		ID:          sessionID,
		ContainerID: containerID,
		CreatedAt:   now,
		LastUsedAt:  now,
		Status:      session.StatusCreated.String(),
	}
	h.Engine.Sessions.Add(&sess)

	// return session identifier to client
	resp := CreateSessionResponse{
		SessionID: sessionID,
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// Update session data
	h.Engine.Sessions.Touch(sess.ID)

	if sess.Status != session.StatusRunning.String() {
		err := h.Engine.Docker.StartContainer(r.Context(), sess.ContainerID)
		if err != nil {
			h.Logger.Error(
				"Failed to start sandbox container",
				"session_id", sess.ID,
				"error", err,
			)
			writeJSONError(w, http.StatusInternalServerError, "Failed to start sandbox container")
			return
		}
	}

	h.Engine.Sessions.UpdateStatus(sess.ID, session.StatusRunning)
	sess.Status = session.StatusRunning.String()

	resp := StartSessionResponse{
		ID:          sess.ID,
		ContainerID: sess.ContainerID,
		CreatedAt:   sess.CreatedAt,
		LastUsedAt:  sess.LastUsedAt,
		Status:      sess.Status,
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Error("failed to encode response", "error", err)
	}

}

func (h *Handler) GetSessionStatusHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// Update session data
	h.Engine.Sessions.Touch(sess.ID)

	resp := GetSessionStatusResponse{
		ID:          sess.ID,
		ContainerID: sess.ContainerID,
		CreatedAt:   sess.CreatedAt,
		LastUsedAt:  sess.LastUsedAt,
		Status:      sess.Status,
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Error("failed to encode response", "error", err)
	}
}
