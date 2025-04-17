package db

import (
    "database/sql"
    "embed"
    "fmt"
    "io/fs"
    "log"
    "os"
    "sort"
    "strings"
    "time"

    _ "github.com/lib/pq" // Postgres driver
)

//go:embed ../../migrations/*.sql
var migrationsFS embed.FS

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
    if err := DB.Ping(); err != nil {
        log.Fatalf("db ping: %v", err)
    }

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

    // collect migration files
    entries, err := fs.ReadDir(migrationsFS, "../../migrations")
    if err != nil {
        return err
    }
    sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

    for _, e := range entries {
        name := e.Name()
        if _, ok := applied[name]; ok {
            continue // already applied
        }
        sqlBytes, err := migrationsFS.ReadFile("../../migrations/" + name)
        if err != nil {
            return err
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
