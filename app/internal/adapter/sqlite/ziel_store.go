package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainziel "agentcontrolplane/app/internal/domain/ziel"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
)

// CreateGoal speichert nur ein Stammziel einer vorhandenen Organisation.
func (d *Database) CreateGoal(ctx context.Context, goal domainziel.Goal) error {
	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO goals
		(id, organization_id, name, parent_goal_id)
		SELECT ?, id, ?, NULL FROM organizations WHERE id = ?`,
		goal.ID, goal.Name, goal.OrganizationID)
	if err != nil {
		return fmt.Errorf("Ziel speichern: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Ziel speichern: %w", err)
	}

	if count == 0 {
		return portorganisation.ErrNotFound
	}

	return nil
}

// ListGoals liest nur Ziele der angegebenen Organisation.
func (d *Database) ListGoals(ctx context.Context, organizationID string) ([]domainziel.Goal, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT id, organization_id, name, parent_goal_id
		FROM goals WHERE organization_id = ? ORDER BY name, id`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("Ziele lesen: %w", err)
	}

	defer rows.Close()
	return d.collectGoals(rows)
}

func (d *Database) collectGoals(rows *sql.Rows) ([]domainziel.Goal, error) {
	goals := []domainziel.Goal{}
	for rows.Next() {
		goal, err := d.scanGoal(rows)
		if err != nil {
			return nil, err
		}

		goals = append(goals, goal)
	}

	return goals, rows.Err()
}

func (d *Database) scanGoal(row interface{ Scan(...any) error }) (domainziel.Goal, error) {
	var goal domainziel.Goal
	var parent sql.NullString
	err := row.Scan(&goal.ID, &goal.OrganizationID, &goal.Name, &parent)
	if errors.Is(err, sql.ErrNoRows) {
		return domainziel.Goal{}, portorganisation.ErrNotFound
	}

	if err != nil {
		return domainziel.Goal{}, fmt.Errorf("Ziel lesen: %w", err)
	}

	if parent.Valid {
		goal.ParentGoalID = &parent.String
	}

	return goal, nil
}
