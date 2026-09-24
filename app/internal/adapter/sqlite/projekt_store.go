package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	portorganisation "agentcontrolplane/app/internal/port/organisation"
	portprojekt "agentcontrolplane/app/internal/port/projekt"
)

// CreateProject speichert ein Projekt nur mit einem Ziel derselben Organisation.
func (d *Database) CreateProject(ctx context.Context, project domainprojekt.Project) error {
	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO projects
		(id, organization_id, name, description, goal_id)
		SELECT ?, organizations.id, ?, ?, goals.id FROM organizations
		JOIN goals ON goals.organization_id = organizations.id
		WHERE organizations.id = ? AND goals.id = ?`,
		project.ID, project.Name, project.Description, project.OrganizationID, project.GoalID)
	return d.projectWriteResult(result, err)
}

func (d *Database) projectWriteResult(result sql.Result, err error) error {
	if err != nil {
		return fmt.Errorf("Projekt speichern: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Projekt speichern: %w", err)
	}

	if count == 0 {
		return portprojekt.ErrInvalidGoal
	}

	return nil
}

// ListProjects liest Projekte der angegebenen Organisation.
func (d *Database) ListProjects(ctx context.Context, organizationID string) ([]domainprojekt.Project, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT id, organization_id, name, description, goal_id
		FROM projects WHERE organization_id = ? ORDER BY name, id`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("Projekte lesen: %w", err)
	}

	defer rows.Close()
	return d.collectProjects(rows)
}

func (d *Database) collectProjects(rows *sql.Rows) ([]domainprojekt.Project, error) {
	projects := []domainprojekt.Project{}
	for rows.Next() {
		project, err := d.scanProject(rows)
		if err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// FindProject liest ein Projekt nur unter seiner Organisationskennung.
func (d *Database) FindProject(ctx context.Context, organizationID, projectID string) (domainprojekt.Project, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `SELECT id, organization_id, name, description, goal_id
		FROM projects WHERE organization_id = ? AND id = ?`, organizationID, projectID)
	return d.scanProject(row)
}

func (d *Database) scanProject(row interface{ Scan(...any) error }) (domainprojekt.Project, error) {
	var project domainprojekt.Project
	err := row.Scan(&project.ID, &project.OrganizationID, &project.Name, &project.Description, &project.GoalID)
	if errors.Is(err, sql.ErrNoRows) {
		return domainprojekt.Project{}, portorganisation.ErrNotFound
	}

	if err != nil {
		return domainprojekt.Project{}, fmt.Errorf("Projekt lesen: %w", err)
	}

	return project, nil
}
