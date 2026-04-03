package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sakkshm/bastion/internal/docker"
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

	// make a new job handler
	jobHandler := session.NewJobHandler()

	// make a new WSManager
	wsManager := websocket.NewWSManager(sessionID)
	go wsManager.Run()

	sess := session.Session{
		ID:          sessionID,
		ContainerID: containerID,
		CreatedAt:   now,
		LastUsedAt:  now,
		Status:      session.StatusCreated,
		JobHandler:  jobHandler,
		WSManager:   wsManager,
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

func (h *Handler) JobExecuteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// make sure session is running
	if sess.Status != session.StatusRunning {
		writeJSONError(w, http.StatusForbidden, "container not started")
		return
	}

	// extract req body
	var req JobExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	if req.Cmd == nil {
		writeJSONError(w, http.StatusBadRequest, "command required")
		return
	}

	// generate job and add to queue
	jobID := session.GenerateJobID()

	// add context for timeout and cancelation
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(h.Engine.Config.Sandbox.JobTTL)*time.Second,
	)

	job := &session.ExecJob{
		JobID:     jobID,
		Cmd:       req.Cmd,
		Status:    session.JobQueued,
		Output:    session.ExecJobOutput{},
		Context:   ctx,
		Cancel:    cancel,
		CreatedAt: time.Now().UTC(),
	}

	sess.JobHandler.Add(job)

	// enqueue async job
	sess.JobHandler.Queue <- job

	// touch session
	h.Engine.Sessions.Touch(sess.ID)

	_ = json.NewEncoder(w).Encode(JobExecResponse{
		JobID:  job.JobID,
		Status: session.JobQueued.String(),
	})
}

func (h *Handler) GetJobStatusHandler(w http.ResponseWriter, r *http.Request) {

	jobID := chi.URLParam(r, "job_id")
	if jobID == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	job, ok := sess.JobHandler.Get(jobID)
	if !ok {
		writeJSONError(w, http.StatusForbidden, "job does not exist")
		return
	}

	// touch session
	h.Engine.Sessions.Touch(sess.ID)

	resp := JobStatusResponse{
		JobID:     job.JobID,
		Cmd:       job.Cmd,
		Status:    job.Status.String(),
		CreatedAt: job.CreatedAt,
	}

	if job.Status == session.JobCompleted || job.Status == session.JobFailed {
		resp.Output = &JobOutputResponse{
			ConsoleOutput: job.Output.ConsoleOutput,
			ErrOut:        job.Output.ErrOut,
			StatusCode:    job.Output.StatusCode,
		}
	}

	_ = json.NewEncoder(w).Encode(resp)

}

func (h *Handler) TerminalHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	h.Engine.Sessions.Touch(sess.ID)

	// start terminal session if not already started
	if !sess.WSManager.TerminalSession.IsConnected {
		ctx, cancel := context.WithCancel(context.Background())

		resp, err := h.Engine.Docker.StartTerminalSession(ctx, sess.ContainerID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "unable to start terminal session")
			cancel()
			return
		}

		termSess := websocket.TerminalSession{
			TerminalResp: resp,
			IsConnected:  true,
			Input:        make(chan websocket.WSTermInputMsg),
			Output:       make(chan websocket.WSTermOutputMsg),
			Ctx:          ctx,
			Cancel:       cancel,
		}

		err = h.Engine.Sessions.AddTerminalSession(sess.ID, &termSess)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "unable to start terminal session")
			cancel()
			return
		}

		newSess, ok := h.Engine.Sessions.Get(sess.ID)
		if !ok {
			writeJSONError(w, http.StatusInternalServerError, "unable to find session")
			cancel()
			return
		}

		// start pumps
		newSess.WSManager.TerminalSession.Start()
		go newSess.WSManager.TermToWSPump()
	}

	// register a client
	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.Engine.Logger.Error(
			"failed to upgrade websocket conn",
			"session_id", sess.ID,
			"error", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "failed to upgrade websocket conn")
		return
	}

	client := websocket.NewClient(conn, sess.ID)

	sess.WSManager.Register <- client

	go client.WritePump()
	go client.ReadPump(sess.WSManager)
}
