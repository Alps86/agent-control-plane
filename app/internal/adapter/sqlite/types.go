package sqlite

import (
	"context"
	"database/sql"
)

// Database verwaltet eine lokale SQLite-Datei und ihre Transaktionen.
type Database struct {
	db         *sql.DB
	migrations []Migration
}

// OrganizationStore bindet Organisationsoperationen ohne Methodenkollision an die Datenbank.
type OrganizationStore struct {
	database *Database
}

// Migration beschreibt einen geordneten, atomar anzuwendenden Schemaschritt.
type Migration struct {
	version    int
	statements []string
}

type transactionKey struct{}

type versionQuery interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}
