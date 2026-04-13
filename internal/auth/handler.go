package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	CreateAPIKeysTable = `
		CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			public_id TEXT NOT NULL UNIQUE,
			secret_hash TEXT NOT NULL,
			name TEXT,
			scope TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			revoked BOOLEAN DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_public_id ON api_keys(public_id);
	`

	StoreAPIKeyQuery = `
		INSERT INTO api_keys (public_id, secret_hash, name, scope)
		VALUES (?, ?, ?, ?)
	`

	RevokeAPIKeyQuery = `
		UPDATE api_keys SET revoked = 1 WHERE public_id = ?;
	`

	ListAPIKeysQuery = `
		SELECT public_id, name, scope, created_at, revoked
		FROM api_keys
		ORDER BY created_at DESC;
	`

	ValidateAPIKeyQuery = `
		SELECT secret_hash, revoked, scope
		FROM api_keys
		WHERE public_id = ?
	`
)

type Scope int

const (
	ScopeAdmin  Scope = 1
	ScopeAgent  Scope = 2
	ScopeViewer Scope = 3
)

func (s Scope) String() string {
	switch s {
	case ScopeAdmin:
		return "admin"
	case ScopeAgent:
		return "agent"
	case ScopeViewer:
		return "viewer"
	default:
		return "unknown"
	}
}

func ParseScope(s string) (Scope, error) {
	switch strings.ToLower(s) {
	case "admin":
		return ScopeAdmin, nil
	case "agent":
		return ScopeAgent, nil
	case "viewer":
		return ScopeViewer, nil
	default:
		return 0, fmt.Errorf("invalid scope: %s", s)
	}
}

func newDBConn() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	_, err = db.Exec(CreateAPIKeysTable)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func CreateAPIKey(name string, scope string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	parsedScope, err := ParseScope(scope)
	if err != nil {
		return errors.New("invalid scope")
	} 

	publicID, secret, fullKey, err := GenerateAPIKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	hashedSecret := HashSecret(secret)

	db, err := newDBConn()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		StoreAPIKeyQuery,
		publicID,
		hashedSecret,
		name,
		parsedScope.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	fmt.Println("API Key Created")
	fmt.Println("----------------------------")
	fmt.Printf("Scope : %s\n", parsedScope.String())
	fmt.Printf("Key   : %s\n", fullKey)
	fmt.Println("----------------------------")
	fmt.Println("Save this key. It won't be shown again.")

	return nil
}

func ListAPIKeys() error {
	db, err := newDBConn()
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(ListAPIKeysQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("API Keys")
	fmt.Println("----------------------------------------------------------------------------------")
	fmt.Printf("%-12s %-15s %-10s %-24s %-8s\n", "PUBLIC_ID", "NAME", "SCOPE", "CREATED", "REVOKED")
	fmt.Println("----------------------------------------------------------------------------------")

	for rows.Next() {
		var id, name, scope, created string
		var revoked bool

		if err := rows.Scan(&id, &name, &scope, &created, &revoked); err != nil {
			return err
		}

		fmt.Printf("%-12s %-15s %-10s %-24s %-8v\n",
			id, name, scope, created, revoked)
	}

	fmt.Println("----------------------------------------------------------------------------------")

	return nil
}

func RevokeAPIKey(publicID string) error {
	if publicID == "" {
		return errors.New("public_id cannot be empty")
	}

	db, err := newDBConn()
	if err != nil {
		return err
	}
	defer db.Close()

	res, err := db.Exec(RevokeAPIKeyQuery, publicID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("no API key found with id: %s", publicID)
	}

	fmt.Println("API Key Revoked")
	fmt.Println("----------------------------")
	fmt.Printf("ID : %s\n", publicID)
	fmt.Println("----------------------------")

	return nil
}

func ValidateAPIKeyWithScope(fullKey string) (Scope, bool, error) {
	parts := strings.Split(fullKey, "_")
	if len(parts) != 3 {
		return 0, false, errors.New("invalid API key format")
	}

	publicID := parts[1]
	secret := parts[2]

	db, err := newDBConn()
	if err != nil {
		return 0, false, err
	}
	defer db.Close()

	var storedHash string
	var revoked bool
	var scopeStr string

	err = db.QueryRow(ValidateAPIKeyQuery, publicID).Scan(&storedHash, &revoked, &scopeStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}

	if revoked {
		return 0, false, nil
	}

	hash := HashSecret(secret)

	if hash != storedHash {
		return 0, false, nil
	}

	scope, err := ParseScope(scopeStr)
	if err != nil {
		return 0, false, err
	}

	return scope, true, nil
}