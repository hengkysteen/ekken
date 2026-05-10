package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

var (
	migrationMu sync.Mutex
	migrations  []string
)

// RegisterMigration allows modules to register their own table schemas.
// This should be called from a module's init() function.
func RegisterMigration(query string) {
	migrationMu.Lock()
	defer migrationMu.Unlock()
	migrations = append(migrations, query)
}

func Open(dataDir string) (*DB, error) {
	dbDir := filepath.Join(dataDir, "db")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "ekken.db")
	conn, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	conn.SetMaxOpenConns(1) // SQLite best practice

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying sql.DB connection for modular repositories.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

func (db *DB) migrate() error {
	migrationMu.Lock()
	defer migrationMu.Unlock()

	// 1. Run core system migrations
	coreQueries := []string{
		`CREATE TABLE IF NOT EXISTS app_secrets (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	}

	for _, q := range coreQueries {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("exec core migration: %w", err)
		}
	}

	// 2. Run registered module migrations
	for _, q := range migrations {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("exec module migration: %w", err)
		}
	}

	return nil
}

func (db *DB) addColumnIfNotExists(table, column, definition string) error {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil // column already exists
		}
	}

	_, err = db.conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition))
	return err
}

// Helper for modules to add columns to their tables if needed in future updates.
func (db *DB) AddColumnIfNotExists(table, column, definition string) error {
	return db.addColumnIfNotExists(table, column, definition)
}
