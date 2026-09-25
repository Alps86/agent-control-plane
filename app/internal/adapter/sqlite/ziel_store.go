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

// CreateChildGoal speichert ein Teilziel nur bei einem Elternziel derselben Organisation.
func (d *Database) CreateChildGoal(ctx context.Context, goal domainziel.Goal) error {
	if goal.ParentGoalID == nil {
		return portziel.ErrInvalidParent
	}

	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO goals
		(id, organization_id, name, parent_goal_id, status)
		SELECT ?, parent.organization_id, ?, parent.id, ? FROM goals parent
		WHERE parent.id = ? AND parent.organization_id = ?`,
		goal.ID, goal.Name, goal.Status, *goal.ParentGoalID, goal.OrganizationID)
	if err != nil {
		return fmt.Errorf("Teilziel speichern: %w", err)
	}

	return d.childWriteResult(result)
}

func (d *Database) childWriteResult(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Teilziel speichern: %w", err)
	}

	if count == 0 {
		return portziel.ErrInvalidParent
	}

	return nil
}

// ListGoals liest nur Ziele der angegebenen Organisation.
func (d *Database) ListGoals(ctx context.Context, organizationID string) ([]domainziel.Goal, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT id, organization_id, name, parent_goal_id, status
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
	err := row.Scan(&goal.ID, &goal.OrganizationID, &goal.Name, &parent, &goal.Status)
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

// SetGoalStatus ändert ausschließlich den ausdrücklich adressierten Zielstatus.
func (d *Database) SetGoalStatus(ctx context.Context, organizationID, goalID, status string) (domainziel.Goal, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `UPDATE goals SET status = ?
		WHERE organization_id = ? AND id = ?
		RETURNING id, organization_id, name, parent_goal_id, status`, status, organizationID, goalID)
	goal, err := d.scanGoal(row)
	if errors.Is(err, portorganisation.ErrNotFound) {
		return domainziel.Goal{}, portziel.ErrGoalNotFound
	}

	return goal, err
}

// ListGoalProjectLinks liest nur Projekte und verknüpfte Ziele derselben Organisation.
func (d *Database) ListGoalProjectLinks(ctx context.Context, organizationID string) ([]domainziel.ProjectLink, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT projects.id, projects.name, goals.id
		FROM projects JOIN goals ON goals.id = projects.goal_id
		AND goals.organization_id = projects.organization_id
		WHERE projects.organization_id = ? ORDER BY projects.name, projects.id`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("Projektzielpfade lesen: %w", err)
	}

	defer rows.Close()
	return d.collectGoalProjectLinks(rows)
}

func (d *Database) collectGoalProjectLinks(rows *sql.Rows) ([]domainziel.ProjectLink, error) {
	links := []domainziel.ProjectLink{}
	for rows.Next() {
		var link domainziel.ProjectLink
		if err := rows.Scan(&link.ProjectID, &link.ProjectName, &link.GoalID); err != nil {
			return nil, fmt.Errorf("Projektzielpfad lesen: %w", err)
		}

		links = append(links, link)
	}

	return links, rows.Err()
}
