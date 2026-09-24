package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	domaindatenbereich "agentcontrolplane/app/internal/domain/datenbereich"
)

// ListDenials zeigt nur Auditereignisse des zugeordneten Agenten.
func (d *Database) ListDenials(ctx context.Context, organizationID, agentID, operatorID string) ([]domaindatenbereich.Denial, error) {
	if err := d.scopeTarget(ctx, organizationID, agentID, operatorID); err != nil {
		return nil, err
	}

	rows, err := d.db.QueryContext(ctx, `SELECT e.id, e.organization_id, e.agent_id,
		COALESCE(p.id, ''), e.action, e.occurred_at FROM agent_data_denials e
		LEFT JOIN projects p ON p.id = e.project_id AND p.organization_id = e.organization_id
		WHERE e.organization_id = ? AND e.agent_id = ? ORDER BY e.id`, organizationID, agentID)
	if err != nil {
		return nil, fmt.Errorf("Verweigerungen lesen: %w", err)
	}

	defer rows.Close()
	return d.collectDenials(rows)
}

func (d *Database) collectDenials(rows *sql.Rows) ([]domaindatenbereich.Denial, error) {
	denials := []domaindatenbereich.Denial{}
	for rows.Next() {
		var denial domaindatenbereich.Denial
		if err := rows.Scan(&denial.ID, &denial.OrganizationID, &denial.AgentID,
			&denial.ProjectID, &denial.Action, &denial.OccurredAt); err != nil {
			return nil, fmt.Errorf("Verweigerung lesen: %w", err)
		}

		denials = append(denials, denial)
	}

	return denials, rows.Err()
}
