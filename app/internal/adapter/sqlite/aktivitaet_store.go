package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	domainaktivitaet "agentcontrolplane/app/internal/domain/aktivitaet"
	portaktivitaet "agentcontrolplane/app/internal/port/aktivitaet"
)

const activityTimeFormat = "2006-01-02T15:04:05.000000000Z"

// Record schreibt ein Ereignis nur für eine Aufgabe derselben Organisation.
func (d *Database) Record(ctx context.Context, event domainaktivitaet.Event) error {
	if !event.Valid() {
		return portaktivitaet.ErrInvalidEvent
	}

	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO activity_events
		(id, organization_id, project_id, task_id, kind, actor, source, occurred_at,
		 object_title, deep_link, assignee_id, assignee_name)
		SELECT ?, t.organization_id, t.project_id, t.id, ?, ?, ?, ?, ?, ?, ?, ?
		FROM tasks t WHERE t.organization_id = ? AND t.project_id = ? AND t.id = ?`,
		event.ID, event.Kind, event.Actor, event.Source, event.OccurredAt.UTC().Format(activityTimeFormat),
		event.ObjectTitle, event.DeepLink, event.AssigneeID, event.AssigneeName,
		event.OrganizationID, event.ProjectID, event.TaskID)
	return d.activityResult(result, err)
}

func (d *Database) activityResult(result sql.Result, err error) error {
	if err != nil {
		return fmt.Errorf("Aktivität speichern: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Aktivität speichern: %w", err)
	}

	if count != 1 {
		return portaktivitaet.ErrInvalidReference
	}

	return nil
}

// ListEvents liest den Verlauf einer Organisation in stabiler Zeitreihenfolge.
func (d *Database) ListEvents(ctx context.Context, organizationID string) ([]domainaktivitaet.Event, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT id, organization_id, project_id, task_id,
		kind, actor, source, occurred_at, object_title, deep_link, assignee_id, assignee_name
		FROM activity_events WHERE organization_id = ?
		ORDER BY occurred_at DESC, sequence DESC`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("Aktivität lesen: %w", err)
	}

	defer rows.Close()
	return d.collectEvents(rows)
}

func (d *Database) collectEvents(rows *sql.Rows) ([]domainaktivitaet.Event, error) {
	events := []domainaktivitaet.Event{}
	for rows.Next() {
		event, err := d.scanEvent(rows)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

func (d *Database) scanEvent(rows *sql.Rows) (domainaktivitaet.Event, error) {
	var event domainaktivitaet.Event
	var occurredAt string
	err := rows.Scan(&event.ID, &event.OrganizationID, &event.ProjectID, &event.TaskID,
		&event.Kind, &event.Actor, &event.Source, &occurredAt, &event.ObjectTitle,
		&event.DeepLink, &event.AssigneeID, &event.AssigneeName)
	if err != nil {
		return event, fmt.Errorf("Aktivität lesen: %w", err)
	}

	event.OccurredAt, err = time.Parse(activityTimeFormat, occurredAt)
	if err != nil {
		return event, fmt.Errorf("Aktivitätszeit lesen: %w", err)
	}

	return event, nil
}
