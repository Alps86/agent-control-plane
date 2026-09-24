package sqlite

import (
	"context"
	"fmt"
	"path/filepath"

	storage "agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/app/persistence"
)

const firstChange = "CREATE TABLE migration_atomicity_probe (id INTEGER PRIMARY KEY)"
const secondChange = "CREATE TABLE migration_atomicity_probe_second (id INTEGER PRIMARY KEY)"

func (s *Suite) migrationBase() error {
	s.dbPath = filepath.Join(s.t.TempDir(), "migration.sqlite")
	store, err := storage.Open(context.Background(), s.dbPath)
	if err != nil {
		return err
	}

	s.store = store
	s.service = persistence.NewService(store)
	version, err := s.service.SchemaVersion(context.Background())
	if err != nil {
		return err
	}

	if version != 1 {
		return fmt.Errorf("initial schema version: got %d, want 1", version)
	}

	return nil
}

func (s *Suite) configureFailingMigration() error {
	migration := storage.NewMigration(2, firstChange, "INSERT INTO missing_migration_target(id) VALUES (1)")
	return s.store.RegisterMigration(migration)
}

func (s *Suite) failingMigration() error {
	s.failure = s.service.Migrate(context.Background())
	return nil
}

func (s *Suite) migrationFailed() error {
	if s.failure == nil {
		return fmt.Errorf("migration unexpectedly succeeded")
	}

	return nil
}

func (s *Suite) versionRemained() error {
	version, err := s.service.SchemaVersion(context.Background())
	if err != nil {
		return err
	}

	if version != 1 {
		return fmt.Errorf("failed migration changed version to %d", version)
	}

	return nil
}

func (s *Suite) correctedMigration() error {
	if err := s.reopenStore(); err != nil {
		return err
	}

	migration := storage.NewMigration(2, firstChange, secondChange)
	if err := s.store.RegisterMigration(migration); err != nil {
		return err
	}

	if err := s.service.Migrate(context.Background()); err != nil {
		return fmt.Errorf("corrected migration could not repeat first change: %w", err)
	}

	s.retried = true
	return nil
}

func (s *Suite) reopenStore() error {
	if err := s.store.Close(); err != nil {
		return err
	}

	store, err := storage.Open(context.Background(), s.dbPath)
	if err != nil {
		return err
	}

	s.store = store
	s.service = persistence.NewService(store)
	return nil
}

func (s *Suite) correctedVersion() error {
	version, err := s.service.SchemaVersion(context.Background())
	if err != nil {
		return err
	}

	if version != 2 {
		return fmt.Errorf("corrected schema version: got %d, want 2", version)
	}

	return nil
}

func (s *Suite) firstChangeRolledBack() error {
	if !s.retried {
		return fmt.Errorf("identical first change did not succeed on retry")
	}

	return nil
}

func (s *Suite) allChangesPresent() error {
	first := "CREATE INDEX migration_probe_first_idx ON migration_atomicity_probe(id)"
	second := "CREATE INDEX migration_probe_second_idx ON migration_atomicity_probe_second(id)"
	if err := s.store.RegisterMigration(storage.NewMigration(3, first, second)); err != nil {
		return err
	}

	if err := s.service.Migrate(context.Background()); err != nil {
		return fmt.Errorf("corrected migration objects not jointly usable: %w", err)
	}

	return nil
}
