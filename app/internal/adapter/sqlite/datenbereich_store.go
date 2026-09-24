package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	domaindatenbereich "agentcontrolplane/app/internal/domain/datenbereich"
	domainprojekt "agentcontrolplane/app/internal/domain/projekt"
	portagent "agentcontrolplane/app/internal/port/agent"
	portdatenbereich "agentcontrolplane/app/internal/port/datenbereich"
)

// ReplaceScopes prüft alle Projektkennungen, bevor es den Umfang atomar ersetzt.
func (d *Database) ReplaceScopes(ctx context.Context, organizationID, agentID, operatorID string, scopes []domaindatenbereich.Scope) error {
	return d.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := d.scopeTarget(txCtx, organizationID, agentID, operatorID); err != nil {
			return err
		}

		if err := d.validateScopeProjects(txCtx, organizationID, scopes); err != nil {
			return err
		}

		return d.replaceScopeRows(txCtx, organizationID, agentID, scopes)
	})
}

func (d *Database) scopeTarget(ctx context.Context, organizationID, agentID, operatorID string) error {
	var present int
	err := d.executor(ctx).QueryRowContext(ctx, `SELECT 1 FROM agents a JOIN organizations o ON o.id = a.organization_id
		WHERE a.id = ? AND a.organization_id = ? AND o.operator_id = ?`, agentID, organizationID, operatorID).Scan(&present)
	if errors.Is(err, sql.ErrNoRows) {
		return portdatenbereich.ErrNotFound
	}

	return err
}

func (d *Database) validateScopeProjects(ctx context.Context, organizationID string, scopes []domaindatenbereich.Scope) error {
	for _, scope := range scopes {
		var present int
		err := d.executor(ctx).QueryRowContext(ctx, `SELECT 1 FROM projects WHERE id = ? AND organization_id = ?`,
			scope.ProjectID, organizationID).Scan(&present)
		if errors.Is(err, sql.ErrNoRows) {
			return portdatenbereich.ErrNotFound
		}

		if err != nil {
			return fmt.Errorf("Projektumfang prüfen: %w", err)
		}
	}

	return nil
}

func (d *Database) replaceScopeRows(ctx context.Context, organizationID, agentID string, scopes []domaindatenbereich.Scope) error {
	_, err := d.executor(ctx).ExecContext(ctx, `DELETE FROM agent_project_scopes WHERE organization_id = ? AND agent_id = ?`, organizationID, agentID)
	if err != nil {
		return fmt.Errorf("Datenbereich ersetzen: %w", err)
	}

	for _, scope := range scopes {
		if !scope.CanRead && !scope.CanWrite {
			continue
		}

		_, err = d.executor(ctx).ExecContext(ctx, `INSERT INTO agent_project_scopes
			(agent_id, organization_id, project_id, can_read, can_write) VALUES (?, ?, ?, ?, ?)`,
			agentID, organizationID, scope.ProjectID, scope.CanRead, scope.CanWrite)
		if err != nil {
			return fmt.Errorf("Datenbereich speichern: %w", err)
		}
	}

	return nil
}

// ListScopes liest nur den Umfang eines Agenten der Betreiberorganisation.
func (d *Database) ListScopes(ctx context.Context, organizationID, agentID, operatorID string) ([]domaindatenbereich.Scope, error) {
	if err := d.scopeTarget(ctx, organizationID, agentID, operatorID); err != nil {
		return nil, err
	}

	rows, err := d.db.QueryContext(ctx, `SELECT agent_id, organization_id, project_id, can_read, can_write
		FROM agent_project_scopes WHERE organization_id = ? AND agent_id = ? ORDER BY project_id`, organizationID, agentID)
	if err != nil {
		return nil, fmt.Errorf("Datenbereich lesen: %w", err)
	}

	defer rows.Close()
	return d.collectScopes(rows)
}

func (d *Database) collectScopes(rows *sql.Rows) ([]domaindatenbereich.Scope, error) {
	scopes := []domaindatenbereich.Scope{}
	for rows.Next() {
		var scope domaindatenbereich.Scope
		if err := rows.Scan(&scope.AgentID, &scope.OrganizationID, &scope.ProjectID, &scope.CanRead, &scope.CanWrite); err != nil {
			return nil, fmt.Errorf("Datenbereich lesen: %w", err)
		}

		scopes = append(scopes, scope)
	}

	return scopes, rows.Err()
}

// FindAgentForOperator liefert ein Profil nur unter seiner Betreiberorganisation.
func (d *Database) FindAgentForOperator(ctx context.Context, organizationID, agentID, operatorID string) (domainagent.Agent, error) {
	agent, err := d.FindAgent(ctx, organizationID, agentID, operatorID)
	if errors.Is(err, portagent.ErrNotFound) {
		return domainagent.Agent{}, portdatenbereich.ErrNotFound
	}

	return agent, err
}

// ListProjectsForOperator zeigt nur Projekte der Agentenorganisation.
func (d *Database) ListProjectsForOperator(ctx context.Context, organizationID, agentID, operatorID string) ([]domainprojekt.Project, error) {
	if err := d.scopeTarget(ctx, organizationID, agentID, operatorID); err != nil {
		return nil, err
	}

	rows, err := d.db.QueryContext(ctx, `SELECT id, organization_id, name, description, goal_id
		FROM projects WHERE organization_id = ? ORDER BY name, id`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("Projekte lesen: %w", err)
	}

	defer rows.Close()
	return d.collectScopeProjects(rows)
}

func (d *Database) collectScopeProjects(rows *sql.Rows) ([]domainprojekt.Project, error) {
	projects := []domainprojekt.Project{}
	for rows.Next() {
		project, err := d.scanScopeProject(rows)
		if err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (d *Database) scanScopeProject(row interface{ Scan(...any) error }) (domainprojekt.Project, error) {
	var project domainprojekt.Project
	err := row.Scan(&project.ID, &project.OrganizationID, &project.Name, &project.Description, &project.GoalID)
	if errors.Is(err, sql.ErrNoRows) {
		return domainprojekt.Project{}, portdatenbereich.ErrNotFound
	}

	if err != nil {
		return domainprojekt.Project{}, fmt.Errorf("Projekt lesen: %w", err)
	}

	return project, nil
}
