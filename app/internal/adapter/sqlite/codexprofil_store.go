package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domain "agentcontrolplane/app/internal/domain/codexprofil"
	portagent "agentcontrolplane/app/internal/port/agent"
)

// Get liest nur ein zur Betreiberin gehörendes CLI-Profil; fehlende Zeilen bedeuten Default false.
func (d *Database) GetCodexProfile(ctx context.Context, organizationID, agentID, operatorID string) (domain.Profile, error) {
	profile := domain.Profile{OrganizationID: organizationID, AgentID: agentID}
	var workspaceEnabled, writeEnabled int
	err := d.executor(ctx).QueryRowContext(ctx, `SELECT COALESCE(p.workspace_enabled, 0), COALESCE(p.write_enabled, 0)
		FROM agents a JOIN organizations o ON o.id = a.organization_id
		LEFT JOIN codex_profiles p ON p.agent_id = a.id
		WHERE a.id = ? AND a.organization_id = ? AND o.operator_id = ? AND a.execution_kind = 'codex_cli'`,
		agentID, organizationID, operatorID).Scan(&workspaceEnabled, &writeEnabled)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Profile{}, portagent.ErrNotFound
	}

	if err != nil {
		return domain.Profile{}, fmt.Errorf("Codex-Profil lesen: %w", err)
	}

	profile.WorkspaceEnabled = workspaceEnabled == 1
	profile.WriteEnabled = writeEnabled == 1
	return profile, nil
}

// SaveCodexProfile schreibt ausschließlich die beiden erlaubten Schalter, scoped auf Agent und Betreiberin.
func (d *Database) SaveCodexProfile(ctx context.Context, organizationID, agentID, operatorID string, input domain.Input) error {
	if err := input.Validate(); err != nil {
		return err
	}

	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO codex_profiles (agent_id, workspace_enabled, write_enabled)
		SELECT a.id, ?, ? FROM agents a JOIN organizations o ON o.id = a.organization_id
		WHERE a.id = ? AND a.organization_id = ? AND o.operator_id = ? AND a.execution_kind = 'codex_cli'
		ON CONFLICT(agent_id) DO UPDATE SET workspace_enabled = excluded.workspace_enabled,
		write_enabled = excluded.write_enabled`, input.WorkspaceEnabled, input.WriteEnabled,
		agentID, organizationID, operatorID)
	if err != nil {
		return fmt.Errorf("Codex-Profil speichern: %w", err)
	}

	return d.codexProfileWriteResult(result)
}

func (d *Database) codexProfileWriteResult(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Codex-Profilzuordnung prüfen: %w", err)
	}

	if count == 0 {
		return portagent.ErrNotFound
	}

	return nil
}
