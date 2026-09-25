package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainziel "agentcontrolplane/app/internal/domain/ziel"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	portziel "agentcontrolplane/app/internal/port/ziel"
)

// MoveGoal ändert die einzige Elternziel-Kante nach atomarer Organisations- und Kreisprüfung.
func (d *Database) MoveGoal(ctx context.Context, organizationID, goalID, parentGoalID string) (domainziel.Goal, error) {
	var moved domainziel.Goal
	err := d.WithinTransaction(ctx, func(txCtx context.Context) error {
		goal, err := d.moveGoalTx(txCtx, organizationID, goalID, parentGoalID)
		moved = goal
		return err
	})
	if err != nil {
		return domainziel.Goal{}, err
	}

	return moved, nil
}

func (d *Database) moveGoalTx(ctx context.Context, organizationID, goalID, parentGoalID string) (domainziel.Goal, error) {
	goal, err := d.goalForMove(ctx, organizationID, goalID)
	if err != nil {
		return domainziel.Goal{}, err
	}

	if err := d.checkMoveParent(ctx, organizationID, goalID, parentGoalID); err != nil {
		return domainziel.Goal{}, err
	}

	return d.updateGoalParent(ctx, goal, parentGoalID)
}

func (d *Database) goalForMove(ctx context.Context, organizationID, goalID string) (domainziel.Goal, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `SELECT id, organization_id, name, parent_goal_id, status
		FROM goals WHERE organization_id = ? AND id = ?`, organizationID, goalID)
	goal, err := d.scanGoal(row)
	if errors.Is(err, portorganisation.ErrNotFound) {
		return domainziel.Goal{}, portziel.ErrGoalNotFound
	}

	if err != nil {
		return domainziel.Goal{}, err
	}

	return goal, nil
}

func (d *Database) checkMoveParent(ctx context.Context, organizationID, goalID, parentGoalID string) error {
	if parentGoalID == "" {
		return nil
	}

	if _, err := d.goalForMove(ctx, organizationID, parentGoalID); err != nil {
		if errors.Is(err, portziel.ErrGoalNotFound) {
			return portziel.ErrInvalidParent
		}

		return err
	}

	return d.checkMoveCycle(ctx, organizationID, goalID, parentGoalID)
}

func (d *Database) checkMoveCycle(ctx context.Context, organizationID, goalID, parentGoalID string) error {
	var found int
	err := d.executor(ctx).QueryRowContext(ctx, `WITH RECURSIVE ancestors(id, parent_goal_id) AS (
		SELECT id, parent_goal_id FROM goals WHERE id = ? AND organization_id = ?
		UNION SELECT goals.id, goals.parent_goal_id FROM goals JOIN ancestors
		ON goals.id = ancestors.parent_goal_id WHERE goals.organization_id = ?)
		SELECT 1 FROM ancestors WHERE id = ? LIMIT 1`, parentGoalID, organizationID, organizationID, goalID).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Zielkante prüfen: %w", err)
	}

	return portziel.ErrGoalCycle
}

func (d *Database) updateGoalParent(ctx context.Context, goal domainziel.Goal, parentGoalID string) (domainziel.Goal, error) {
	var parent any
	if parentGoalID != "" {
		parent = parentGoalID
	}

	row := d.executor(ctx).QueryRowContext(ctx, `UPDATE goals SET parent_goal_id = ?
		WHERE organization_id = ? AND id = ? RETURNING id, organization_id, name, parent_goal_id, status`,
		parent, goal.OrganizationID, goal.ID)
	updated, err := d.scanGoal(row)
	if err != nil {
		return domainziel.Goal{}, fmt.Errorf("Zielkante speichern: %w", err)
	}

	return updated, nil
}
