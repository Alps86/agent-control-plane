package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Open öffnet den lokalen Datenbestand und führt ausstehende Migrationen aus.
func Open(ctx context.Context, path string) (*Database, error) {
	return OpenWithMigrations(ctx, path)
}

// OpenWithMigrations ergänzt fachliche Migrationen vor dem Serverstart.
func OpenWithMigrations(ctx context.Context, path string, additional ...Migration) (*Database, error) {
	database := &Database{}
	absolute, err := database.openFile(path)
	if err != nil {
		return nil, err
	}

	if err = database.connect(ctx, (&url.URL{Scheme: "file", Path: absolute}).String()); err != nil {
		return nil, err
	}

	if err = database.registerAndMigrate(ctx, additional); err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}

func (d *Database) openFile(path string) (string, error) {
	if path == "" {
		return "", errors.New("Datenbankpfad fehlt")
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Datenbankpfad auflösen: %w", err)
	}

	return absolute, d.prepareFile(absolute)
}

func (d *Database) prepareFile(absolute string) error {
	file, err := os.OpenFile(absolute, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("Datenbank öffnen: %w", err)
	}

	if err = file.Chmod(0600); err != nil {
		file.Close()
		return fmt.Errorf("Datenbankrechte setzen: %w", err)
	}

	if err = file.Close(); err != nil {
		return fmt.Errorf("Datenbankdatei schließen: %w", err)
	}

	return nil
}

func (d *Database) connect(ctx context.Context, dsn string) error {
	connection, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("Datenbank öffnen: %w", err)
	}

	connection.SetMaxOpenConns(1)
	d.db = connection
	if err = d.initialize(ctx); err != nil {
		connection.Close()
		return err
	}

	return nil
}

// Close gibt die Datenbankverbindung frei.
func (d *Database) Close() error {
	return d.db.Close()
}
