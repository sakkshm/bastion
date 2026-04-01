package session

import (
	"maps"
	"sync"
	"time"
)

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) Add(s *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[s.ID] = s
}

func (sm *SessionManager) Get(ID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, ok := sm.sessions[ID]
	return s, ok
}

func (sm *SessionManager) Delete(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}

func (sm *SessionManager) BatchDelete(toDelete []string) {
	sm.mu.Lock()
	for _, id := range toDelete {
		delete(sm.sessions, id)
	}
	sm.mu.Unlock()
}

func (sm *SessionManager) Count() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return len(sm.sessions)
}

// Update lastUsedAt field after an operation
func (sm *SessionManager) Touch(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if s, ok := sm.sessions[id]; ok {
		s.LastUsedAt = time.Now().UTC()
	}
}

func (sm *SessionManager) UpdateStatus(id string, status Status) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if s, ok := sm.sessions[id]; ok {
		s.Status = status
	}
}

func (sm *SessionManager) Snapshot() map[string]*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	cp := make(map[string]*Session, len(sm.sessions))
	maps.Copy(cp, sm.sessions)

	return cp
}
