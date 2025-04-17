package db

import (
    "database/sql"
    "log"
    "os"

    _ "github.com/lib/pq" // postgres driver
    "github.com/jmoiron/sqlx"
)

// DB is a global handle accessed throughout the app.
var DB *sqlx.DB

// Connect initialises the global database handle.
// It expects DATABASE_URL env var (e.g. postgres://user:pass@host:5432/db?sslmode=disable).
// Falls back to a sensible local default for docker‑compose dev.
func Connect() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://postgres:postgres@localhost:5432/agentsaas?sslmode=disable"
    }

    var err error
    DB, err = sqlx.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("db: open connection failed: %v", err)
    }

    if err := DB.Ping(); err != nil {
        log.Fatalf("db: ping failed: %v", err)
    }

    // pool config – small by default, tweak later
    DB.SetMaxIdleConns(5)
    DB.SetMaxOpenConns(20)
}

// Tx helper runs fn inside a transaction.
func Tx(fn func(*sqlx.Tx) error) error {
    tx, err := DB.Beginx()
    if err != nil {
        return err
    }
    if err := fn(tx); err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}

// NullString converts sql.NullString to *string.
func NullString(s sql.NullString) *string {
    if !s.Valid {
        return nil
    }
    return &s.String
}
