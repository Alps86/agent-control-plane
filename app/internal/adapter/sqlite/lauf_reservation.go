package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	domainlauf "agentcontrolplane/app/internal/domain/lauf"
)

// Reserve legt höchstens eine aktive Reservierung je Organisation und Aufgabe an.
func (d *Database) Reserve(ctx context.Context, organizationID, taskID string) (domainlauf.Run, bool, error) {
	var run domainlauf.Run
	var created bool
	err := d.WithinTransaction(ctx, func(txContext context.Context) error {
		var reserveError error
		run, created, reserveError = d.reserveIn(txContext, organizationID, taskID)
		return reserveError
	})
	if err != nil {
		return domainlauf.Run{}, false, err
	}

	return run, created, err
}

func (d *Database) reserveIn(ctx context.Context, organizationID, taskID string) (domainlauf.Run, bool, error) {
	id, err := d.newRunID()
	if err != nil {
		return domainlauf.Run{}, false, err
	}

	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO lauf_reservations
		(id, organization_id, task_id, status) VALUES (?, ?, ?, 'Reserviert')
		ON CONFLICT DO NOTHING`, id, organizationID, taskID)
	if err != nil {
		return domainlauf.Run{}, false, fmt.Errorf("Lauf reservieren: %w", err)
	}

	return d.reservationResult(ctx, organizationID, taskID, result)
}

func (d *Database) reservationResult(ctx context.Context, organizationID, taskID string, result sql.Result) (domainlauf.Run, bool, error) {
	affected, err := result.RowsAffected()
	if err != nil {
		return domainlauf.Run{}, false, err
	}

	run, found, err := d.activeRun(ctx, organizationID, taskID)
	if err != nil {
		return domainlauf.Run{}, false, err
	}

	if !found {
		return domainlauf.Run{}, false, errors.New("aktive Laufreservierung fehlt")
	}

	return run, affected == 1, nil
}

func (d *Database) activeRun(ctx context.Context, organizationID, taskID string) (domainlauf.Run, bool, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `SELECT id, organization_id, task_id, status, failure_reason
		FROM lauf_reservations WHERE organization_id = ? AND task_id = ? AND status = 'Reserviert'`, organizationID, taskID)
	return d.readRun(row)
}

func (d *Database) newRunID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", fmt.Errorf("Laufkennung erzeugen: %w", err)
	}

	return hex.EncodeToString(id[:]), nil
}

func (d *Database) readRun(row *sql.Row) (domainlauf.Run, bool, error) {
	var run domainlauf.Run
	err := row.Scan(&run.ID, &run.OrganizationID, &run.TaskID, &run.Status, &run.FailureReason)
	if errors.Is(err, sql.ErrNoRows) {
		return domainlauf.Run{}, false, nil
	}

	if err != nil {
		return domainlauf.Run{}, false, fmt.Errorf("Lauf lesen: %w", err)
	}

	return run, true, nil
}
