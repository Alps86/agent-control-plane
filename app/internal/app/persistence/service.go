package persistence

import (
	"context"

	"agentcontrolplane/app/internal/port/persistence"
)

// NewService bindet die Speichergrenze an einen Anwendungsfall.
func NewService(store persistence.Store) *Service {
	return &Service{store: store}
}

// Migrate wendet die registrierte Migrationskette atomar an.
func (s *Service) Migrate(ctx context.Context) error {
	return s.store.Migrate(ctx)
}

// SchemaVersion liefert die dauerhaft gespeicherte Schemaversion.
func (s *Service) SchemaVersion(ctx context.Context) (int, error) {
	return s.store.SchemaVersion(ctx)
}
