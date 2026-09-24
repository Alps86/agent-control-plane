package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

var initialMigration = Migration{
	version: 1, statements: []string{
		`CREATE TABLE schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	},
}

// NewMigration erzeugt einen versionierten Schemaschritt aus SQL-Anweisungen.
func NewMigration(version int, statements ...string) Migration {
	return Migration{version: version, statements: append([]string(nil), statements...)}
}

// RegisterMigration ergänzt die nächste Version der Migrationskette.
func (d *Database) RegisterMigration(step Migration) error {
	if step.version != len(d.migrations)+1 || len(step.statements) == 0 {
		return fmt.Errorf("ungültige Migration %d", step.version)
	}

	d.migrations = append(d.migrations, step)
	return nil
}

func (d *Database) registerAndMigrate(ctx context.Context, additional []Migration) error {
	if err := d.RegisterMigration(initialMigration); err != nil {
		return err
	}

	for _, step := range additional {
		if err := d.RegisterMigration(step); err != nil {
			return err
		}
	}

	return d.Migrate(ctx)
}

func (d *Database) initialize(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("Datenbank verbinden: %w", err)
	}

	if _, err := d.db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		return fmt.Errorf("Datenbank konfigurieren: %w", err)
	}

	return nil
}

// Migrate führt die registrierte Kette in einer SQLite-Transaktion aus.
func (d *Database) Migrate(ctx context.Context) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Migration starten: %w", err)
	}

	defer tx.Rollback()
	return d.migrateIn(ctx, tx)
}

func (d *Database) migrateIn(ctx context.Context, tx *sql.Tx) error {
	version, err := d.versionIn(ctx, tx)
	if err != nil {
		return err
	}

	if version > len(d.migrations) {
		return fmt.Errorf("Datenbankschema %d ist neuer als die Anwendung", version)
	}

	if err = d.applyPending(ctx, tx, version); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Database) applyPending(ctx context.Context, tx *sql.Tx, version int) error {
	for _, step := range d.migrations {
		if step.version <= version {
			continue
		}

		if err := d.apply(ctx, tx, step); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) apply(ctx context.Context, tx *sql.Tx, step Migration) error {
	for _, statement := range step.statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("Migration %d: %w", step.version, err)
		}
	}

	_, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES (?)", step.version)
	if err != nil {
		return fmt.Errorf("Migration %d speichern: %w", step.version, err)
	}

	return nil
}

// SchemaVersion liefert die höchste vollständig angewendete Migration.
func (d *Database) SchemaVersion(ctx context.Context) (int, error) {
	return d.versionIn(ctx, d.executor(ctx))
}

func (d *Database) versionIn(ctx context.Context, query versionQuery) (int, error) {
	var exists int
	err := query.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("Migrationsstand lesen: %w", err)
	}

	if exists == 0 {
		return 0, nil
	}

	var version int
	err = query.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("Migrationsstand lesen: %w", err)
	}

	return version, nil
}
