package persistence

import "context"

// Store beschreibt die technische Speichergrenze ohne Datenbanktypen.
type Store interface {
	SchemaVersion(context.Context) (int, error)
	Migrate(context.Context) error
	WithinTransaction(context.Context, func(context.Context) error) error
	Close() error
}
