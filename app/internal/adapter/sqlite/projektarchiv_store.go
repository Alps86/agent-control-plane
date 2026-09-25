package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainprojektarchiv "agentcontrolplane/app/internal/domain/projektarchiv"
	portprojektarchiv "agentcontrolplane/app/internal/port/projektarchiv"
)

// ArchiveProject wahrt Organisationsbezug und blockiert aktive Läufe atomar.
func (d *Database) ArchiveProject(ctx context.Context, organizationID, projectID string) error {
	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO project_archives (organization_id, project_id)
		SELECT p.organization_id, p.id FROM projects p
		WHERE p.organization_id = ? AND p.id = ?
		AND NOT EXISTS (SELECT 1 FROM lauf_reservations r
			JOIN tasks t ON t.organization_id = r.organization_id AND t.id = r.task_id
			WHERE t.organization_id = p.organization_id AND t.project_id = p.id
			AND r.status = 'Reserviert')
		ON CONFLICT (organization_id, project_id) DO NOTHING`, organizationID, projectID)
	if err != nil {
		return fmt.Errorf("Projekt archivieren: %w", err)
	}

	return d.archiveResult(ctx, organizationID, projectID, result)
}

func (d *Database) archiveResult(ctx context.Context, organizationID, projectID string, result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Archivierung prüfen: %w", err)
	}

	if count == 1 {
		return nil
	}

	return d.archiveConflict(ctx, organizationID, projectID)
}

func (d *Database) archiveConflict(ctx context.Context, organizationID, projectID string) error {
	status, err := d.ProjectStatus(ctx, organizationID, projectID)
	if err != nil {
		return err
	}

	if status == domainprojektarchiv.StatusArchived {
		return portprojektarchiv.ErrAlreadyArchived
	}

	return portprojektarchiv.ErrActiveRun
}

// RestoreProject entfernt nur den Archivmarker und erhält Aufgaben und Verlauf.
func (d *Database) RestoreProject(ctx context.Context, organizationID, projectID string) error {
	result, err := d.executor(ctx).ExecContext(ctx, `DELETE FROM project_archives
		WHERE organization_id = ? AND project_id = ?`, organizationID, projectID)
	if err != nil {
		return fmt.Errorf("Projekt wiederherstellen: %w", err)
	}

	return d.restoreResult(ctx, organizationID, projectID, result)
}

func (d *Database) restoreResult(ctx context.Context, organizationID, projectID string, result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Wiederherstellung prüfen: %w", err)
	}

	if count == 1 {
		return nil
	}

	if _, err := d.ProjectStatus(ctx, organizationID, projectID); err != nil {
		return err
	}

	return portprojektarchiv.ErrNotArchived
}

// ProjectStatus liest den Zustand mit Organisationsfilter direkt aus SQLite.
func (d *Database) ProjectStatus(ctx context.Context, organizationID, projectID string) (domainprojektarchiv.Status, error) {
	var archived bool
	err := d.executor(ctx).QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM project_archives a
		WHERE a.organization_id = p.organization_id AND a.project_id = p.id)
		FROM projects p WHERE p.organization_id = ? AND p.id = ?`, organizationID, projectID).Scan(&archived)
	if errors.Is(err, sql.ErrNoRows) {
		return "", portprojektarchiv.ErrNotFound
	}

	if err != nil {
		return "", fmt.Errorf("Projektstatus lesen: %w", err)
	}

	if archived {
		return domainprojektarchiv.StatusArchived, nil
	}

	return domainprojektarchiv.StatusActive, nil
}

// ListArchivedProjects erhält Aufgaben unabhängig von der Archivansicht.
func (d *Database) ListArchivedProjects(ctx context.Context, organizationID string) ([]domainprojektarchiv.ArchivedProject, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT p.id, p.organization_id, p.name,
		p.description, p.goal_id, a.archived_at FROM project_archives a
		JOIN projects p ON p.organization_id = a.organization_id AND p.id = a.project_id
		WHERE a.organization_id = ? ORDER BY p.name, p.id`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("Projektarchiv lesen: %w", err)
	}

	defer rows.Close()
	return d.collectArchivedProjects(rows)
}

func (d *Database) collectArchivedProjects(rows *sql.Rows) ([]domainprojektarchiv.ArchivedProject, error) {
	projects := []domainprojektarchiv.ArchivedProject{}
	for rows.Next() {
		var project domainprojektarchiv.ArchivedProject
		if err := rows.Scan(&project.ID, &project.OrganizationID, &project.Name,
			&project.Description, &project.GoalID, &project.ArchivedAt); err != nil {
			return nil, fmt.Errorf("Projektarchiv lesen: %w", err)
		}

		projects = append(projects, project)
	}

	return projects, rows.Err()
}
