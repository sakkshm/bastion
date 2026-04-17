package session

import (
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/sakkshm/bastion/internal/database"
	"github.com/sakkshm/bastion/internal/websocket"
)

type SessionManager struct {
	sessions map[string]*Session
	DBConn   *database.DatabaseConn
	mu       sync.RWMutex
}

func NewSessionManager(conn *database.DatabaseConn) (*SessionManager, error) {

	// make sessionManager
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		DBConn:   conn,
	}
	// make db sessions table (if not exist)
	err := sm.CreateSessionsDataTable()
	if err != nil {
		return nil, err
	}

	// TODO: reconcile the state

	return sm, err
}

func (sm *SessionManager) Add(s *Session) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	err := sm.AddSessionData(s)
	if err != nil {
		return err
	}

	sm.sessions[s.ID] = s
	return nil
}

func (sm *SessionManager) BatchAdd(sessions []Session) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// TODO: unoptimised, fix this, make a batch add db handler
	for _, s := range sessions {
		err := sm.AddSessionData(&s)
		if err != nil {
			return err
		}

		sm.sessions[s.ID] = &s
	}
	return nil
}

func (sm *SessionManager) Get(ID string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, ok := sm.sessions[ID]
	return s, ok
}

func (sm *SessionManager) Delete(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	err := sm.DeleteSessionData(id)
	if err != nil {
		return err
	}

	delete(sm.sessions, id)
	return nil
}

func (sm *SessionManager) BatchDelete(toDelete []string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, id := range toDelete {
		// TODO: unoptimised, fix this, make a batch delete db handler
		_ = sm.DeleteSessionData(id)
		delete(sm.sessions, id)
	}
}

func (sm *SessionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// Update lastUsedAt field after an operation
func (sm *SessionManager) Touch(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now().UTC()

	err := sm.TouchSessionData(id, now)
	if err != nil {
		return err
	}

	if s, ok := sm.sessions[id]; ok {
		s.LastUsedAt = now
	}
	return nil
}

func (sm *SessionManager) UpdateStatus(id string, status Status) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	err := sm.UpdateSessionStatusData(id, status)
	if err != nil {
		return err
	}

	if s, ok := sm.sessions[id]; ok {
		s.Status = status
	}
	return nil
}

func (sm *SessionManager) Snapshot() map[string]*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	cp := make(map[string]*Session, len(sm.sessions))
	maps.Copy(cp, sm.sessions)

	return cp
}

func (sm *SessionManager) AddTerminalSession(id string, term *websocket.TerminalSession) error {

	sess, ok := sm.Get(id)
	if !ok {
		return fmt.Errorf("session not found")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess.WSManager.TerminalSession = *term
	return nil
}
