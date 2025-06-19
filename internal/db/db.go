package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"time"

	_ "github.com/lib/pq" // Postgres driver

	"github.com/cedricziel/mel-agent/migrations"
)

var DB *sql.DB

// Connect opens the database and applies migrations.
func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/agentsaas?sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}

	// Configure connection pool for horizontal scaling
	// These settings are optimized for multiple API server instances
	DB.SetMaxOpenConns(25)                 // Maximum number of open connections to the database
	DB.SetMaxIdleConns(10)                 // Maximum number of connections in the idle connection pool
	DB.SetConnMaxLifetime(5 * time.Minute) // Maximum amount of time a connection may be reused
	DB.SetConnMaxIdleTime(2 * time.Minute) // Maximum amount of time a connection may be idle

	if err := DB.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	log.Printf("Database connected with pool: max_open=%d, max_idle=%d, max_lifetime=%v",
		25, 10, 5*time.Minute)

	if err := applyMigrations(); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
}

// applyMigrations reads migration files embedded at build time and applies any not yet run.
func applyMigrations() error {
	if _, err := DB.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version TEXT PRIMARY KEY,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
        )`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// read applied versions into a set
	rows, err := DB.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	defer rows.Close()
	applied := map[string]struct{}{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = struct{}{}
	}

	// collect migration files from embed FS
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, e := range entries {
		name := e.Name()
		if _, ok := applied[name]; ok {
			continue // already applied
		}
		// read migration SQL from embedded migrations FS
		sqlBytes, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := DB.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec %s: %w", name, err)
		}
		if _, err := DB.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`, name, time.Now()); err != nil {
			return err
		}
		log.Printf("migrated %s", name)
	}
	return nil
}

// Tx runs fn inside a SQL transaction.
func Tx(fn func(*sql.Tx) error) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
