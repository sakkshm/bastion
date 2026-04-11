package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseConn struct {
	Database *sql.DB
}

func NewDBConn() (*DatabaseConn, error) {
	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return &DatabaseConn{
		Database: db,
	}, nil
}

func (db *DatabaseConn) Close() error {
	return db.Database.Close()
}
