package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domain "agentcontrolplane/app/internal/domain/modellwahl"
	port "agentcontrolplane/app/internal/port/modellwahl"
)

// Get reads a route only through the agent's organization and operator owner.
func (d *Database) Get(ctx context.Context, organizationID, agentID, operatorID string) (domain.Selection, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `SELECT m.provider_id, m.connection_id, m.model_id
		FROM agent_model_selections m JOIN agents a ON a.id=m.agent_id
		JOIN organizations o ON o.id=a.organization_id
		WHERE a.id=? AND a.organization_id=? AND o.operator_id=?`, agentID, organizationID, operatorID)
	var selection domain.Selection
	err := row.Scan(&selection.Provider, &selection.ConnectionReference, &selection.Model)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Selection{}, port.ErrNotFound
	}

	if err != nil {
		return domain.Selection{}, fmt.Errorf("Modellwahl lesen: %w", err)
	}

	return selection, nil
}

// Replace changes only the authorized Eino agent's route in one statement.
func (d *Database) Replace(ctx context.Context, organizationID, agentID, operatorID string, selection domain.Selection) error {
	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO agent_model_selections
		(agent_id,provider_id,connection_id,model_id) SELECT a.id,?,?,? FROM agents a
		JOIN organizations o ON o.id=a.organization_id WHERE a.id=? AND a.organization_id=?
		AND a.execution_kind='eino' AND o.operator_id=?
		ON CONFLICT(agent_id) DO UPDATE SET provider_id=excluded.provider_id,
		connection_id=excluded.connection_id,model_id=excluded.model_id,updated_at=CURRENT_TIMESTAMP`,
		selection.Provider, selection.ConnectionReference, selection.Model, agentID, organizationID, operatorID)
	if err != nil {
		return fmt.Errorf("Modellwahl speichern: %w", err)
	}

	return d.choiceResult(result)
}

func (d *Database) choiceResult(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Modellwahlzuordnung prüfen: %w", err)
	}

	if count == 0 {
		return port.ErrNotFound
	}

	return nil
}
