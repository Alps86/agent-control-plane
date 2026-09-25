package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainort "agentcontrolplane/app/internal/domain/projektort"
	portort "agentcontrolplane/app/internal/port/projektort"
)

// Get liest nur die Bindung des Projekts in der angegebenen Organisation.
func (d *Database) Get(ctx context.Context, organizationID, projectID string) (domainort.Binding, error) {
	var binding domainort.Binding
	err := d.executor(ctx).QueryRowContext(ctx, `SELECT l.organization_id, l.project_id, l.kind
		FROM project_locations l JOIN projects p ON p.id = l.project_id
		WHERE l.organization_id = ? AND l.project_id = ? AND p.organization_id = ?`,
		organizationID, projectID, organizationID).Scan(&binding.OrganizationID, &binding.ProjectID, &binding.Kind)
	if errors.Is(err, sql.ErrNoRows) {
		return domainort.Binding{}, portort.ErrMissing
	}

	if err != nil {
		return domainort.Binding{}, fmt.Errorf("Projektort lesen: %w", err)
	}

	return binding, nil
}

// Save bindet nur ein existierendes Projekt derselben Organisation.
func (d *Database) Save(ctx context.Context, binding domainort.Binding) error {
	if err := binding.Validate(); err != nil {
		return err
	}

	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO project_locations (project_id, organization_id, kind)
		SELECT id, organization_id, ? FROM projects WHERE id = ? AND organization_id = ?
		ON CONFLICT(project_id) DO UPDATE SET kind = excluded.kind
		WHERE project_locations.organization_id = excluded.organization_id`,
		binding.Kind, binding.ProjectID, binding.OrganizationID)
	if err != nil {
		return fmt.Errorf("Projektort speichern: %w", err)
	}

	return d.projectLocationWriteResult(result)
}

func (d *Database) projectLocationWriteResult(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Projektort speichern: %w", err)
	}

	if count == 0 {
		return portort.ErrNotFound
	}

	return nil
}
