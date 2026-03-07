package session

import (
	"sync"
)

type JobHandler struct {
	Jobs  map[string]*ExecJob
	Queue chan *ExecJob
	mu    sync.RWMutex
}

const QUEUE_BUFFER_LEN = 64

func NewJobHandler() *JobHandler {
	return &JobHandler{
		Jobs:  make(map[string]*ExecJob),
		Queue: make(chan *ExecJob, QUEUE_BUFFER_LEN),
	}
}

func (h *JobHandler) Add(e *ExecJob) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Jobs[e.JobID] = e
}

func (h *JobHandler) Get(ID string) (*ExecJob, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	e, ok := h.Jobs[ID]
	return e, ok
}

func (h *JobHandler) Delete(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.Jobs, id)
}

func (h *JobHandler) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Jobs)
}
