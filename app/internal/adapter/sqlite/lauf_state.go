package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainlauf "agentcontrolplane/app/internal/domain/lauf"
)

// Lookup sucht einen Lauf ausschließlich in seiner Organisation und Aufgabe.
func (d *Database) Lookup(ctx context.Context, organizationID, taskID, runID string) (domainlauf.Run, bool, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `SELECT id, organization_id, task_id, status, failure_reason
		FROM lauf_reservations WHERE organization_id = ? AND task_id = ? AND id = ?`, organizationID, taskID, runID)
	return d.readRun(row)
}

// List liefert die dauerhaft gespeicherten Läufe einer Aufgabe.
func (d *Database) List(ctx context.Context, organizationID, taskID string) ([]domainlauf.Run, error) {
	query := d.executor(ctx).(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	})
	rows, err := query.QueryContext(ctx, `SELECT id, organization_id, task_id, status, failure_reason
		FROM lauf_reservations WHERE organization_id = ? AND task_id = ? ORDER BY rowid`, organizationID, taskID)
	if err != nil {
		return nil, fmt.Errorf("Läufe lesen: %w", err)
	}

	defer rows.Close()
	return d.collectRuns(rows)
}

func (d *Database) collectRuns(rows *sql.Rows) ([]domainlauf.Run, error) {
	runs := []domainlauf.Run{}
	for rows.Next() {
		var run domainlauf.Run
		if err := rows.Scan(&run.ID, &run.OrganizationID, &run.TaskID, &run.Status, &run.FailureReason); err != nil {
			return nil, fmt.Errorf("Lauf lesen: %w", err)
		}

		runs = append(runs, run)
	}

	return runs, rows.Err()
}

// FailPreparation hält einen Vorbereitungsfehler vor jedem Adapterlauf fest.
func (d *Database) FailPreparation(ctx context.Context, organizationID, taskID, runID, reason string) (domainlauf.Run, error) {
	var changed domainlauf.Run
	err := d.WithinTransaction(ctx, func(txContext context.Context) error {
		var transitionError error
		changed, transitionError = d.failPreparationIn(txContext, organizationID, taskID, runID, reason)
		return transitionError
	})
	if err != nil {
		return domainlauf.Run{}, err
	}

	return changed, err
}

func (d *Database) failPreparationIn(ctx context.Context, organizationID, taskID, runID, reason string) (domainlauf.Run, error) {
	run, found, err := d.Lookup(ctx, organizationID, taskID, runID)
	if err != nil {
		return domainlauf.Run{}, err
	}

	if !found {
		return domainlauf.Run{}, errors.New("Lauf nicht gefunden")
	}

	changed, err := run.FailPreparation(reason)
	if err != nil {
		return domainlauf.Run{}, err
	}

	return d.persistPreparationFailure(ctx, changed)
}

func (d *Database) persistPreparationFailure(ctx context.Context, run domainlauf.Run) (domainlauf.Run, error) {
	result, err := d.executor(ctx).ExecContext(ctx, `UPDATE lauf_reservations
		SET status = ?, failure_reason = ?
		WHERE id = ? AND organization_id = ? AND task_id = ? AND status = 'Reserviert'`,
		run.Status, run.FailureReason, run.ID, run.OrganizationID, run.TaskID)
	if err != nil {
		return domainlauf.Run{}, fmt.Errorf("Vorbereitungsfehler speichern: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return domainlauf.Run{}, err
	}

	if affected != 1 {
		return domainlauf.Run{}, domainlauf.ErrInvalidTransition
	}

	return run, nil
}
