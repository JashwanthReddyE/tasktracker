package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// dsn builds a SQLite connection string. The pragmas must travel in the DSN
// rather than being issued after Open: foreign_keys and busy_timeout are
// per-connection, and database/sql opens connections lazily throughout the
// process lifetime, so anything set once on the first connection would silently
// not apply to the rest of the pool.
func dsn(path string) string {
	// SQLite wants a URI here. Forward slashes are valid on Windows, and spaces
	// are the one character realistically found in a Windows profile path that
	// would otherwise terminate URI parsing.
	p := strings.ReplaceAll(filepath.ToSlash(path), " ", "%20")
	return "file:" + p +
		"?_pragma=journal_mode(WAL)" +
		"&_pragma=foreign_keys(ON)" +
		"&_pragma=busy_timeout(5000)" +
		"&_pragma=synchronous(NORMAL)"
}

func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// WAL permits one writer alongside concurrent readers. busy_timeout above
	// makes contending writers wait rather than fail, so a modest pool is safe.
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// migrate applies column additions that CREATE TABLE IF NOT EXISTS cannot make
// to a database that predates them. SQLite has no ADD COLUMN IF NOT EXISTS, so
// each statement runs unconditionally and a "duplicate column" error — meaning
// it was already applied — is treated as success.
func migrate(db *sql.DB) error {
	stmts := []string{
		`ALTER TABLE users ADD COLUMN prefs TEXT NOT NULL DEFAULT '{}'`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil &&
			!strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

// tx runs fn inside a transaction, rolling back on error or panic. Multi-table
// writes (a task plus its people plus its events) must be atomic or a crash
// mid-write leaves a task with half its assignees.
func tx(db *sql.DB, fn func(*sql.Tx) error) error {
	t, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			t.Rollback()
			panic(p)
		}
	}()
	if err := fn(t); err != nil {
		t.Rollback()
		return err
	}
	return t.Commit()
}
