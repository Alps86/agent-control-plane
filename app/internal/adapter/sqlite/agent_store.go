package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	domainagent "agentcontrolplane/app/internal/domain/agent"
	portagent "agentcontrolplane/app/internal/port/agent"
	sqlite3 "modernc.org/sqlite"
)

// CreateAgent bindet den Agenten atomar an eine zugeordnete Organisation.
func (d *Database) CreateAgent(ctx context.Context, operatorID string, agent domainagent.Agent) error {
	return d.WithinTransaction(ctx, func(txCtx context.Context) error {
		return d.insertAgent(txCtx, operatorID, agent)
	})
}

func (d *Database) insertAgent(ctx context.Context, operatorID string, agent domainagent.Agent) error {
	capabilities, err := json.Marshal(agent.Capabilities)
	if err != nil {
		return fmt.Errorf("Fachfähigkeiten kodieren: %w", err)
	}

	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO agents
		(id, organization_id, name, name_key, role, instructions, execution_kind, template_id, capabilities)
		SELECT ?, id, ?, ?, ?, ?, ?, ?, ? FROM organizations
		WHERE id = ? AND operator_id = ?`, agent.ID, agent.Name, strings.ToLower(agent.Name),
		agent.Role, agent.Instructions, agent.ExecutionKind, agent.TemplateID,
		string(capabilities), agent.OrganizationID, operatorID)
	return d.agentWriteResult(result, err)
}

func (d *Database) agentWriteResult(result sql.Result, err error) error {
	var sqliteErr *sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code() == 2067 {
		return portagent.ErrNameConflict
	}

	if err != nil {
		return fmt.Errorf("Agent speichern: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Agentenzuordnung prüfen: %w", err)
	}

	if count == 0 {
		return portagent.ErrNotFound
	}

	return nil
}

// ListAgents liest ausschließlich Agenten der zugeordneten Organisation.
func (d *Database) ListAgents(ctx context.Context, organizationID, operatorID string) ([]domainagent.Agent, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT a.id, a.organization_id, a.name, a.role,
		a.instructions, a.execution_kind, a.template_id, a.capabilities FROM agents a
		JOIN organizations o ON o.id = a.organization_id
		WHERE a.organization_id = ? AND o.operator_id = ? ORDER BY a.name_key, a.id`, organizationID, operatorID)
	if err != nil {
		return nil, fmt.Errorf("Agenten lesen: %w", err)
	}

	defer rows.Close()
	return d.collectAgents(rows)
}

func (d *Database) collectAgents(rows *sql.Rows) ([]domainagent.Agent, error) {
	agents := []domainagent.Agent{}
	for rows.Next() {
		agent, err := d.scanAgent(rows)
		if err != nil {
			return nil, err
		}

		agents = append(agents, agent)
	}

	return agents, rows.Err()
}

// FindAgent verdeckt fremde Organisations- und Agentenkennungen gleichermaßen.
func (d *Database) FindAgent(ctx context.Context, organizationID, agentID, operatorID string) (domainagent.Agent, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `SELECT a.id, a.organization_id, a.name, a.role,
		a.instructions, a.execution_kind, a.template_id, a.capabilities FROM agents a
		JOIN organizations o ON o.id = a.organization_id
		WHERE a.organization_id = ? AND a.id = ? AND o.operator_id = ?`, organizationID, agentID, operatorID)
	return d.scanAgent(row)
}

func (d *Database) scanAgent(row interface{ Scan(...any) error }) (domainagent.Agent, error) {
	var agent domainagent.Agent
	var capabilities string
	err := row.Scan(&agent.ID, &agent.OrganizationID, &agent.Name, &agent.Role,
		&agent.Instructions, &agent.ExecutionKind, &agent.TemplateID, &capabilities)
	if errors.Is(err, sql.ErrNoRows) {
		return domainagent.Agent{}, portagent.ErrNotFound
	}

	if err != nil {
		return domainagent.Agent{}, fmt.Errorf("Agent lesen: %w", err)
	}

	return d.decodeAgent(agent, capabilities)
}

func (d *Database) decodeAgent(agent domainagent.Agent, capabilities string) (domainagent.Agent, error) {
	if err := json.Unmarshal([]byte(capabilities), &agent.Capabilities); err != nil {
		return domainagent.Agent{}, fmt.Errorf("Fachfähigkeiten lesen: %w", err)
	}

	return agent, nil
}
