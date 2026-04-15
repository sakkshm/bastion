package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sakkshm/bastion/internal/docker"
	"github.com/sakkshm/bastion/internal/filesystem"
	"github.com/sakkshm/bastion/internal/session"
	"github.com/sakkshm/bastion/internal/websocket"
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

	// make an FSWorkspace
	filesystem, err := filesystem.NewFSWorkspace(*h.Engine.Config, sessionID)
	if err != nil {
		h.Engine.Logger.Error(
			"Failed to create filesystem for container",
			"session_id", sessionID,
			"error", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "Failed to create filesystem for container")
		return
	}

	// create a container
	containerConfig := docker.ContainerConfig{
		Image:          h.Engine.Config.Sandbox.Image,
		Memory:         h.Engine.Config.Sandbox.Memory,
		CPUs:           h.Engine.Config.Sandbox.CPUs,
		PIDs:           h.Engine.Config.Sandbox.PIDs,
		NetworkEnabled: h.Engine.Config.Sandbox.NetworkEnabled,
		LoadEnv:        h.Engine.Config.Sandbox.LoadEnv,
		EnvPath:        h.Engine.Config.Execution.EnvFilePath,
		FileSystem:     *filesystem,
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

	// make a new job handler
	jobHandler := session.NewJobHandler()

	// make a new WSManager
	wsManager := websocket.NewWSManager(sessionID, h.Engine.Sessions.Touch)
	go wsManager.Run()

	sess := session.Session{
		ID:          sessionID,
		ContainerID: containerID,
		CreatedAt:   now,
		LastUsedAt:  now,
		Status:      session.StatusCreated,
		JobHandler:  jobHandler,
		WSManager:   wsManager,
		FileSystem:  filesystem,
	}
	h.Engine.Sessions.Add(&sess)

	// atatch a worker to this session to execute jobs
	h.Engine.Logger.Info("Attaching workers to session", "session_id", sess.ID)
	h.Engine.AttachWorker(&sess)

	// return session identifier to client
	resp := CreateSessionResponse{
		SessionID: sessionID,
		Status:    sess.Status.String(),
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
	if sess.Status != session.StatusRunning {
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
	sess.Status = session.StatusRunning

	resp := StartSessionResponse{
		ID:          sess.ID,
		ContainerID: sess.ContainerID,
		CreatedAt:   sess.CreatedAt,
		LastUsedAt:  sess.LastUsedAt,
		Status:      sess.Status.String(),
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
	if sess.Status != session.StatusStopped {
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
	sess.Status = session.StatusStopped

	resp := StopSessionResponse{
		ID:          sess.ID,
		ContainerID: sess.ContainerID,
		CreatedAt:   sess.CreatedAt,
		LastUsedAt:  sess.LastUsedAt,
		Status:      sess.Status.String(),
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
	if sess.Status != session.StatusDeleted {
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
	if sess.Status != session.StatusDeleted {
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
			sess.Status == session.StatusRunning {

			h.Engine.Sessions.UpdateStatus(sess.ID, containerStatus)
			sess.Status = containerStatus
		}

	}

	resp := GetSessionStatusResponse{
		ID:          sess.ID,
		ContainerID: sess.ContainerID,
		CreatedAt:   sess.CreatedAt,
		LastUsedAt:  sess.LastUsedAt,
		Status:      sess.Status.String(),
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Engine.Logger.Error("failed to encode response", "error", err)
	}
}
