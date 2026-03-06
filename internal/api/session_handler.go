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

	// check if max concurrent sessions reached
	if h.Engine.Sessions.Count() >= h.Engine.Config.Execution.MaxConcurrent {
		h.Engine.Logger.Error("Maximum sessions reached", "error", "max_sessions_reached")
		writeJSONError(w, http.StatusTooManyRequests, "Maximum sessions reached")
		return
	}

	// generate session ID
	sessionID := session.GenerateSessionID()

	// create a container
	containerConfig := docker.ContainerConfig{
		Image:          h.Engine.Config.Sandbox.Image,
		Memory:         h.Engine.Config.Sandbox.Memory,
		CPUs:           h.Engine.Config.Sandbox.CPUs,
		PIDs:           h.Engine.Config.Sandbox.PIDs,
		NetworkEnabled: h.Engine.Config.Sandbox.NetworkEnabled,
	}

	containerID, err := h.Engine.Docker.CreateSandboxContainer(
		r.Context(),
		containerConfig,
		sessionID,
	)

	if err != nil {
		h.Engine.Logger.Error(
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
		Status:    sess.Status,
		CreatedAt: sess.CreatedAt,
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Engine.Logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) StartSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// Update session data
	h.Engine.Sessions.Touch(sess.ID)

	// start container if not already running
	if sess.Status != session.StatusRunning.String() {
		err := h.Engine.Docker.StartContainer(r.Context(), sess.ContainerID)
		if err != nil {
			h.Engine.Logger.Error(
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
		h.Engine.Logger.Error("failed to encode response", "error", err)
	}

}

func (h *Handler) StopSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// Update session data
	h.Engine.Sessions.Touch(sess.ID)

	// stop if not already stopped
	if sess.Status != session.StatusStopped.String() {
		err := h.Engine.Docker.StopContainer(r.Context(), sess.ContainerID)
		if err != nil {
			h.Engine.Logger.Error(
				"Failed to stop sandbox container",
				"session_id", sess.ID,
				"error", err,
			)
			writeJSONError(w, http.StatusInternalServerError, "Failed to stop sandbox container")
			return
		}
	}

	h.Engine.Sessions.UpdateStatus(sess.ID, session.StatusStopped)
	sess.Status = session.StatusStopped.String()

	resp := StopSessionResponse{
		ID:          sess.ID,
		ContainerID: sess.ContainerID,
		CreatedAt:   sess.CreatedAt,
		LastUsedAt:  sess.LastUsedAt,
		Status:      sess.Status,
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Engine.Logger.Error("failed to encode response", "error", err)
	}

}

func (h *Handler) DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// delete container if not already deleted
	// does not delete entry from SessionManager - Will be handles by GC
	if sess.Status != session.StatusDeleted.String() {
		err := h.Engine.Docker.DeleteContainer(r.Context(), sess.ContainerID)
		if err != nil {
			h.Engine.Logger.Error(
				"Failed to delete sandbox container",
				"session_id", sess.ID,
				"error", err,
			)
			writeJSONError(w, http.StatusInternalServerError, "Failed to delete sandbox container")
			return
		}

		h.Engine.Sessions.UpdateStatus(sess.ID, session.StatusDeleted)
	}

	w.WriteHeader(http.StatusOK)

	resp := DeleteSessionResponse{
		SessionID: sess.ID,
		Status:    session.StatusDeleted.String(),
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Engine.Logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) GetSessionStatusHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	h.Engine.Sessions.Touch(sess.ID)

	var (
		containerStatus session.Status
		err             error
	)

	// sync status if not deleted
	if sess.Status != session.StatusDeleted.String() {
		containerStatus, err = h.Engine.Docker.GetContainerStatus(
			r.Context(),
			sess.ContainerID,
		)

		if err != nil {
			h.Engine.Logger.Error(
				"failed to inspect container",
				"session_id", sess.ID,
				"error", err,
			)
			writeJSONError(w, http.StatusInternalServerError, "failed to inspect container")
			return
		}

		// sync docker state with session state
		// only update the session state if the session thinks it's running but docker says it isn't
		if containerStatus != session.StatusRunning &&
			containerStatus != session.StatusBusy &&
			sess.Status == session.StatusRunning.String() {

			h.Engine.Sessions.UpdateStatus(sess.ID, containerStatus)
			sess.Status = containerStatus.String()
		}

	}

	resp := GetSessionStatusResponse{
		ID:          sess.ID,
		ContainerID: sess.ContainerID,
		CreatedAt:   sess.CreatedAt,
		LastUsedAt:  sess.LastUsedAt,
		Status:      sess.Status,
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Engine.Logger.Error("failed to encode response", "error", err)
	}
}
