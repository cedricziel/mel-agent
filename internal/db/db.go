package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"strconv"
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
	maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	connMaxIdleTime := getEnvDuration("DB_CONN_MAX_IDLE_TIME", 2*time.Minute)

	DB.SetMaxOpenConns(maxOpenConns)
	DB.SetMaxIdleConns(maxIdleConns)
	DB.SetConnMaxLifetime(connMaxLifetime)
	DB.SetConnMaxIdleTime(connMaxIdleTime)

	if err := DB.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	log.Printf("Database connected with pool: max_open=%d, max_idle=%d, max_lifetime=%v",
		maxOpenConns, maxIdleConns, connMaxLifetime)

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

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvDuration gets a duration environment variable with a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: Invalid duration value for %s: %s, using default: %v", key, value, defaultValue)
	}
	return defaultValue
}
