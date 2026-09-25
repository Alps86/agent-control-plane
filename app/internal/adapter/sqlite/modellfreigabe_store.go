package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	portmodell "agentcontrolplane/app/internal/port/modellfreigabe"
	"agentcontrolplane/app/internal/port/modellschluessel"
)

func NewModellfreigabeStore(database *Database) *ModellfreigabeStore {
	return &ModellfreigabeStore{database: database}
}

func (s *ModellfreigabeStore) GrantOrganization(ctx context.Context, operatorID, organizationID string, binding modellschluessel.Binding) error {
	result, err := s.database.executor(ctx).ExecContext(ctx, `INSERT INTO openrouter_organization_grants(organization_id, reference, generation)
		SELECT id, ?, ? FROM organizations WHERE id = ? AND operator_id = ?
		ON CONFLICT(organization_id) DO UPDATE SET reference = excluded.reference, generation = excluded.generation`,
		binding.Reference, binding.Generation, organizationID, operatorID)
	return s.organizationWriteResult(ctx, result, err, operatorID, organizationID, binding)
}

func (s *ModellfreigabeStore) RevokeOrganization(ctx context.Context, operatorID, organizationID string) error {
	return s.database.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.revokeOrganization(txCtx, operatorID, organizationID)
	})
}

func (s *ModellfreigabeStore) revokeOrganization(ctx context.Context, operatorID, organizationID string) error {
	_, err := s.database.executor(ctx).ExecContext(ctx, `DELETE FROM openrouter_agent_grants
		WHERE organization_id = ? AND organization_id IN
		(SELECT id FROM organizations WHERE operator_id = ?)`, organizationID, operatorID)
	if err != nil {
		return fmt.Errorf("Agentenfreigaben widerrufen: %w", err)
	}
	_, err = s.database.executor(ctx).ExecContext(ctx, `DELETE FROM openrouter_organization_grants
		WHERE organization_id = ? AND organization_id IN
		(SELECT id FROM organizations WHERE operator_id = ?)`, organizationID, operatorID)
	return err
}

func (s *ModellfreigabeStore) OrganizationGranted(ctx context.Context, operatorID, organizationID string, binding modellschluessel.Binding) (bool, error) {
	return s.exists(ctx, `SELECT EXISTS(SELECT 1 FROM openrouter_organization_grants g
		JOIN organizations o ON o.id = g.organization_id
		WHERE g.organization_id = ? AND o.operator_id = ? AND g.reference = ? AND g.generation = ?)`,
		organizationID, operatorID, binding.Reference, binding.Generation)
}

func (s *ModellfreigabeStore) GrantAgent(ctx context.Context, operatorID, organizationID, agentID string, binding modellschluessel.Binding) error {
	result, err := s.database.executor(ctx).ExecContext(ctx, `INSERT INTO openrouter_agent_grants(agent_id, organization_id, reference, generation)
		SELECT a.id, a.organization_id, ?, ? FROM agents a
		JOIN organizations o ON o.id = a.organization_id
		JOIN openrouter_organization_grants g ON g.organization_id = a.organization_id
		WHERE a.id = ? AND a.organization_id = ? AND o.operator_id = ? AND a.execution_kind = 'eino'
		AND g.reference = ? AND g.generation = ?
		ON CONFLICT(agent_id) DO UPDATE SET reference = excluded.reference, generation = excluded.generation`,
		binding.Reference, binding.Generation, agentID, organizationID, operatorID, binding.Reference, binding.Generation)
	if err != nil {
		return fmt.Errorf("Agentenfreigabe speichern: %w", err)
	}
	return s.agentWriteResult(ctx, result, operatorID, organizationID, agentID, binding)
}

func (s *ModellfreigabeStore) agentWriteResult(ctx context.Context, result sql.Result, operatorID, organizationID, agentID string, binding modellschluessel.Binding) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	granted, err := s.AgentGranted(ctx, operatorID, organizationID, agentID, binding)
	if err != nil {
		return err
	}
	if granted {
		return nil
	}
	return portmodell.ErrOrganizationGrantMissing
}

func (s *ModellfreigabeStore) RevokeAgent(ctx context.Context, operatorID, organizationID, agentID string) error {
	_, err := s.database.executor(ctx).ExecContext(ctx, `DELETE FROM openrouter_agent_grants WHERE agent_id = ?
		AND organization_id = ? AND organization_id IN
		(SELECT id FROM organizations WHERE operator_id = ?)`, agentID, organizationID, operatorID)
	return err
}

func (s *ModellfreigabeStore) AgentGranted(ctx context.Context, operatorID, organizationID, agentID string, binding modellschluessel.Binding) (bool, error) {
	return s.exists(ctx, `SELECT EXISTS(SELECT 1 FROM openrouter_agent_grants g
		JOIN organizations o ON o.id = g.organization_id
		WHERE g.agent_id = ? AND g.organization_id = ? AND o.operator_id = ?
		AND g.reference = ? AND g.generation = ?)`, agentID, organizationID, operatorID, binding.Reference, binding.Generation)
}

func (s *ModellfreigabeStore) exists(ctx context.Context, query string, args ...any) (bool, error) {
	var found bool
	err := s.database.executor(ctx).QueryRowContext(ctx, query, args...).Scan(&found)
	if err != nil {
		return false, fmt.Errorf("Modellfreigabe lesen: %w", err)
	}
	return found, nil
}

func (s *ModellfreigabeStore) organizationWriteResult(ctx context.Context, result sql.Result, err error, operatorID, organizationID string, binding modellschluessel.Binding) error {
	if err != nil {
		return fmt.Errorf("Modellfreigabe speichern: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	granted, err := s.OrganizationGranted(ctx, operatorID, organizationID, binding)
	if err != nil {
		return err
	}
	if granted {
		return nil
	}
	return portmodell.ErrOrganizationGrantMissing
}
