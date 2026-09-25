package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	domainberichtsweg "agentcontrolplane/app/internal/domain/berichtsweg"
	portberichtsweg "agentcontrolplane/app/internal/port/berichtsweg"
)

// ListLines liest alle Agenten und ihre Berichtslinie nur für den Betreiber.
func (d *Database) ListLines(ctx context.Context, organizationID, operatorID string) ([]domainberichtsweg.Line, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT a.id, a.name, a.execution_kind,
		COALESCE(r.parent_agent_id, '') FROM agents a
		JOIN organizations o ON o.id = a.organization_id
		LEFT JOIN agent_reporting_lines r ON r.organization_id = a.organization_id AND r.agent_id = a.id
		WHERE a.organization_id = ? AND o.operator_id = ? ORDER BY a.name_key, a.id`, organizationID, operatorID)
	if err != nil {
		return nil, fmt.Errorf("Berichtswege lesen: %w", err)
	}

	defer rows.Close()
	return d.collectLines(rows)
}

func (d *Database) collectLines(rows *sql.Rows) ([]domainberichtsweg.Line, error) {
	lines := []domainberichtsweg.Line{}
	for rows.Next() {
		var line domainberichtsweg.Line
		if err := rows.Scan(&line.AgentID, &line.Name, &line.ExecutionKind, &line.ParentID); err != nil {
			return nil, fmt.Errorf("Berichtsweg lesen: %w", err)
		}

		lines = append(lines, line)
	}

	return lines, rows.Err()
}

// AssignParent setzt oder entfernt die Berichtskante atomar und ohne Fremdbezug.
func (d *Database) AssignParent(ctx context.Context, organizationID, operatorID, agentID, parentID string) error {
	return d.WithinTransaction(ctx, func(txCtx context.Context) error {
		return d.assignParent(txCtx, organizationID, operatorID, agentID, parentID)
	})
}

func (d *Database) assignParent(ctx context.Context, organizationID, operatorID, agentID, parentID string) error {
	if err := d.requireScopedAgent(ctx, organizationID, operatorID, agentID); err != nil {
		return err
	}

	if parentID == "" {
		return d.clearParent(ctx, organizationID, agentID)
	}

	if err := d.requireScopedAgent(ctx, organizationID, operatorID, parentID); err != nil {
		return err
	}

	return d.assignCheckedParent(ctx, organizationID, agentID, parentID)
}

func (d *Database) requireScopedAgent(ctx context.Context, organizationID, operatorID, agentID string) error {
	var exists bool
	err := d.executor(ctx).QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM agents a JOIN organizations o ON o.id = a.organization_id
		WHERE a.organization_id = ? AND o.operator_id = ? AND a.id = ?
	)`, organizationID, operatorID, agentID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("Agentenzuordnung prüfen: %w", err)
	}

	if !exists {
		return portberichtsweg.ErrAgentNotFound
	}

	return nil
}

func (d *Database) clearParent(ctx context.Context, organizationID, agentID string) error {
	_, err := d.executor(ctx).ExecContext(ctx, `DELETE FROM agent_reporting_lines
		WHERE organization_id = ? AND agent_id = ?`, organizationID, agentID)
	if err != nil {
		return fmt.Errorf("Berichtsweg entfernen: %w", err)
	}

	return nil
}

func (d *Database) assignCheckedParent(ctx context.Context, organizationID, agentID, parentID string) error {
	if agentID == parentID {
		return portberichtsweg.ErrSelfParent
	}

	cycle, err := d.parentReachesAgent(ctx, organizationID, parentID, agentID)
	if err != nil {
		return err
	}

	if cycle {
		return portberichtsweg.ErrCycle
	}

	return d.writeParent(ctx, organizationID, agentID, parentID)
}

func (d *Database) parentReachesAgent(ctx context.Context, organizationID, parentID, agentID string) (bool, error) {
	var reaches bool
	err := d.executor(ctx).QueryRowContext(ctx, `WITH RECURSIVE ancestors(id) AS (
		SELECT ? UNION SELECT r.parent_agent_id FROM agent_reporting_lines r
		JOIN ancestors a ON a.id = r.agent_id WHERE r.organization_id = ?
	) SELECT EXISTS (SELECT 1 FROM ancestors WHERE id = ?)`, parentID, organizationID, agentID).Scan(&reaches)
	if err != nil {
		return false, fmt.Errorf("Berichtslinie prüfen: %w", err)
	}

	return reaches, nil
}

func (d *Database) writeParent(ctx context.Context, organizationID, agentID, parentID string) error {
	_, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO agent_reporting_lines
		(organization_id, agent_id, parent_agent_id) VALUES (?, ?, ?)
		ON CONFLICT (organization_id, agent_id) DO UPDATE SET parent_agent_id = excluded.parent_agent_id`,
		organizationID, agentID, parentID)
	if err != nil {
		return fmt.Errorf("Berichtsweg speichern: %w", err)
	}

	return nil
}
