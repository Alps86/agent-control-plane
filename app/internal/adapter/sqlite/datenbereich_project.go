package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	domaindatenbereich "agentcontrolplane/app/internal/domain/datenbereich"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	portdatenbereich "agentcontrolplane/app/internal/port/datenbereich"
)

// ReadProjectForAgent schützt den echten Projektdatensatz im SQL-Prädikat.
func (d *Database) ReadProjectForAgent(ctx context.Context, organizationID, agentID, projectID string) (domainprojekt.Project, error) {
	row := d.db.QueryRowContext(ctx, `SELECT p.id, p.organization_id, p.name, p.description, p.goal_id
		FROM projects p JOIN agent_project_scopes s ON s.project_id = p.id AND s.organization_id = p.organization_id
		JOIN agents a ON a.id = s.agent_id AND a.organization_id = p.organization_id
		WHERE p.id = ? AND p.organization_id = ? AND a.id = ? AND s.can_read = 1`, projectID, organizationID, agentID)
	return d.scanScopeProject(row)
}

// UpdateProjectDescriptionForAgent schreibt nur das benannte Feld im erlaubten Projekt.
func (d *Database) UpdateProjectDescriptionForAgent(ctx context.Context, organizationID, agentID, projectID, description string) (domainprojekt.Project, error) {
	result, err := d.db.ExecContext(ctx, `UPDATE projects SET description = ? WHERE id = ? AND organization_id = ?
		AND EXISTS (SELECT 1 FROM agent_project_scopes s JOIN agents a ON a.id = s.agent_id
			WHERE s.project_id = projects.id AND s.organization_id = projects.organization_id
			AND a.organization_id = projects.organization_id AND a.id = ? AND s.can_write = 1)`,
		description, projectID, organizationID, agentID)
	if err != nil {
		return domainprojekt.Project{}, fmt.Errorf("Projektbeschreibung schreiben: %w", err)
	}

	return d.writeProjectResult(result, organizationID, projectID, description)
}

func (d *Database) writeProjectResult(result sql.Result, organizationID, projectID, description string) (domainprojekt.Project, error) {
	count, err := result.RowsAffected()
	if err != nil {
		return domainprojekt.Project{}, fmt.Errorf("Projektänderung prüfen: %w", err)
	}

	if count == 0 {
		return domainprojekt.Project{}, portdatenbereich.ErrNotFound
	}

	// Eine reine Schreibfreigabe gibt weder Projektname noch Zielbezug frei.
	return domainprojekt.Project{ID: projectID, OrganizationID: organizationID, Description: description}, nil
}

// AuditDenied bindet das Ereignis an die echte Agentenorganisation, nicht den URL-Pfad.
func (d *Database) AuditDenied(ctx context.Context, _ string, agentID, projectID string, action domaindatenbereich.Action) error {
	_, err := d.db.ExecContext(ctx, `INSERT INTO agent_data_denials(organization_id, agent_id, project_id, action)
		SELECT a.organization_id, a.id, ?, ? FROM agents a WHERE a.id = ?`, projectID, action, agentID)
	if err != nil {
		return fmt.Errorf("Datenzugriff auditieren: %w", err)
	}

	return nil
}
