package session

import (
	"database/sql"
	"time"
)

const (
	CreateSessionsDataTable = `
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			container_id TEXT,
			created_at INTEGER,
			last_used_at INTEGER,
			status TEXT,
			fs_mount TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_sessions_container_id ON sessions(container_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
		CREATE INDEX IF NOT EXISTS idx_sessions_last_used_at ON sessions(last_used_at);
	`

	AddSessionData = `
		INSERT INTO sessions(
			id,
			container_id,
			created_at,
			last_used_at,
			status,
			fs_mount
		) VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			last_used_at=excluded.last_used_at,
			status=excluded.status;
	`

	DeleteSessionData = `
		DELETE FROM sessions WHERE id = ?
	`

	TouchSessionData = `
		UPDATE sessions SET last_used_at = ? WHERE id = ? 
	`

	UpdateSessionStatusData = `
		UPDATE sessions SET status = ? WHERE id = ? 
	`

	GetAllSessionsData = `
		SELECT id, container_id, created_at, last_used_at, status, fs_mount
		FROM sessions
	`
)

func (sm *SessionManager) withTx(fn func(tx *sql.Tx) error) error {
	tx, err := sm.DBConn.Database.Begin()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (sm *SessionManager) CreateSessionsDataTable() error {
	return sm.withTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(CreateSessionsDataTable)
		return err
	})
}

func (sm *SessionManager) AddSessionData(s *Session) error {
	return sm.withTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			AddSessionData,
			s.ID,
			s.ContainerID,
			s.CreatedAt.Unix(),
			s.LastUsedAt.Unix(),
			s.Status.String(),
			s.FileSystem.Mount,
		)
		return err
	})
}

func (sm *SessionManager) DeleteSessionData(id string) error {
	return sm.withTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			DeleteSessionData,
			id,
		)
		return err
	})
}

func (sm *SessionManager) TouchSessionData(id string, now time.Time) error {
	return sm.withTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			TouchSessionData,
			now.Unix(),
			id,
		)
		return err
	})
}

func (sm *SessionManager) UpdateSessionStatusData(id string, status Status) error {
	return sm.withTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			UpdateSessionStatusData,
			status.String(),
			id,
		)
		return err
	})
}

