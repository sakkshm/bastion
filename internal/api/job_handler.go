package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sakkshm/bastion/internal/session"
	"github.com/sakkshm/bastion/internal/websocket"
)

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
			Input:        make(chan websocket.WSTermInputMsg, 256),
			Output:       make(chan websocket.WSTermOutputMsg, 256),
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
