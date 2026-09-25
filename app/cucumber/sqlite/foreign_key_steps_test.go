package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	storage "agentcontrolplane/app/internal/adapter/sqlite"
	"agentcontrolplane/app/internal/app/persistence"
	"github.com/cucumber/godog"
)

const createParent = "CREATE TABLE fk_parent (id INTEGER PRIMARY KEY)"
const createChild = "CREATE TABLE fk_child (id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES fk_parent(id))"
const insertOrphan = "INSERT INTO fk_child (id, parent_id) VALUES (1, 999)"

func (s *Suite) initializeForeignKeyScenario(sc *godog.ScenarioContext) {
	sc.Step(`^ein lokaler SQLite-Bestand mit verknüpften Tabellen ist angelegt$`, s.foreignKeyBase)
	sc.Step(`^eine Migration einen Kinddatensatz ohne Elternsatz schreiben will$`, s.foreignKeyMigration)
	sc.Step(`^wird der Fremdschlüsselverstoß abgewiesen und die Schemaversion bleibt unverändert$`, s.foreignKeyRejected)
	sc.Step(`^ich denselben Bestand erneut öffne und den Fremdschlüsselverstoß wiederhole$`, s.reopenForeignKeyBase)
	sc.Step(`^ein vorhandener SQLite-Bestand enthält eine ungültige Fremdschlüsselreferenz$`, s.legacyOrphan)
	sc.Step(`^ich den Bestand über die öffentliche SQLite-Grenze öffne$`, s.openLegacyOrphan)
	sc.Step(`^wird der beschädigte Bestand mit einem verständlichen Fremdschlüsselfehler abgewiesen$`, s.legacyOrphanRejected)
}

func (s *Suite) foreignKeyBase() error {
	s.dbPath = filepath.Join(s.t.TempDir(), "fk.sqlite")
	store, err := storage.OpenWithMigrations(context.Background(), s.dbPath,
		storage.NewMigration(2, createParent, createChild))
	if err != nil {
		return err
	}
	s.store = store
	s.service = persistence.NewService(store)
	return nil
}

func (s *Suite) foreignKeyMigration() error {
	if err := s.store.RegisterMigration(storage.NewMigration(3, insertOrphan)); err != nil {
		return err
	}
	s.failure = s.service.Migrate(context.Background())
	return nil
}

func (s *Suite) foreignKeyRejected() error {
	if s.failure == nil || !strings.Contains(s.failure.Error(), "FOREIGN KEY constraint failed") {
		return fmt.Errorf("Fremdschlüsselverstoß nicht abgewiesen: %v", s.failure)
	}
	version, err := s.service.SchemaVersion(context.Background())
	if err != nil {
		return err
	}
	if version != 2 {
		return fmt.Errorf("Schemaversion nach abgewiesener Migration = %d, erwartet 2", version)
	}
	return nil
}

func (s *Suite) reopenForeignKeyBase() error {
	if err := s.store.Close(); err != nil {
		return err
	}
	store, err := storage.OpenWithMigrations(context.Background(), s.dbPath,
		storage.NewMigration(2, createParent, createChild))
	if err != nil {
		return err
	}
	s.store = store
	s.service = persistence.NewService(store)
	return s.foreignKeyMigration()
}

func (s *Suite) legacyOrphan() error {
	s.dbPath = filepath.Join(s.t.TempDir(), "legacy-orphan.sqlite")
	legacy, err := sql.Open("sqlite", s.dbPath)
	if err != nil {
		return err
	}
	defer legacy.Close()
	legacy.SetMaxOpenConns(1)
	for _, statement := range []string{
		"PRAGMA foreign_keys = OFF",
		createParent,
		createChild,
		insertOrphan,
	} {
		if _, err := legacy.ExecContext(context.Background(), statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *Suite) openLegacyOrphan() error {
	store, err := storage.Open(context.Background(), s.dbPath)
	if store != nil {
		store.Close()
	}
	s.failure = err
	return nil
}

func (s *Suite) legacyOrphanRejected() error {
	if s.failure == nil || !strings.Contains(s.failure.Error(), "verletzte Fremdschlüssel") {
		return fmt.Errorf("ungültiger Altbestand nicht klar abgewiesen: %v", s.failure)
	}
	return nil
}
