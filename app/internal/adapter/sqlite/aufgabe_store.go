package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	domainaufgabe "agentcontrolplane/app/internal/domain/aufgabe"
	portaufgabe "agentcontrolplane/app/internal/port/aufgabe"
)

// CreateTask schreibt nur bei passendem Projekt und aktivem Agenten derselben Organisation.
func (d *Database) CreateTask(ctx context.Context, task domainaufgabe.Task) error {
	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO tasks
		(id, organization_id, project_id, title, description, priority, assignee_id, status)
		SELECT ?, p.organization_id, p.id, ?, ?, ?, a.id, ? FROM projects p
		JOIN agents a ON a.organization_id = p.organization_id AND a.id = ? AND a.status = 'active'
		JOIN organizations o ON o.id = p.organization_id
		WHERE p.organization_id = ? AND p.id = ?`, task.ID, task.Title, task.Description,
		task.Priority, task.Status, task.AssigneeID, task.OrganizationID, task.ProjectID)
	return d.taskWriteResult(result, err)
}

func (d *Database) taskWriteResult(result sql.Result, err error) error {
	if err != nil && strings.Contains(err.Error(), "project_archived") {
		return portaufgabe.ErrProjectArchived
	}

	if err != nil {
		return fmt.Errorf("Aufgabe speichern: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Aufgabe speichern: %w", err)
	}

	if count != 1 {
		return portaufgabe.ErrInvalidReference
	}

	return nil
}

// ListTasks liest ausschließlich die Aufgaben des gewählten Projekts.
func (d *Database) ListTasks(ctx context.Context, organizationID, projectID string) ([]domainaufgabe.Task, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT id, organization_id, project_id, title,
		description, priority, assignee_id, status FROM tasks
		WHERE organization_id = ? AND project_id = ? ORDER BY title, id`, organizationID, projectID)
	if err != nil {
		return nil, fmt.Errorf("Aufgaben lesen: %w", err)
	}

	defer rows.Close()
	return d.collectTasks(rows)
}

func (d *Database) collectTasks(rows *sql.Rows) ([]domainaufgabe.Task, error) {
	tasks := []domainaufgabe.Task{}
	for rows.Next() {
		task, err := d.scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// FindTask verdeckt fremde Aufgabenkennungen durch die Organisationsbedingung.
func (d *Database) FindTask(ctx context.Context, organizationID, taskID string) (domainaufgabe.Task, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `SELECT id, organization_id, project_id, title,
		description, priority, assignee_id, status FROM tasks
		WHERE organization_id = ? AND id = ?`, organizationID, taskID)
	return d.scanTask(row)
}

func (d *Database) scanTask(row interface{ Scan(...any) error }) (domainaufgabe.Task, error) {
	var task domainaufgabe.Task
	err := row.Scan(&task.ID, &task.OrganizationID, &task.ProjectID, &task.Title,
		&task.Description, &task.Priority, &task.AssigneeID, &task.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return domainaufgabe.Task{}, portaufgabe.ErrNotFound
	}

	if err != nil {
		return domainaufgabe.Task{}, fmt.Errorf("Aufgabe lesen: %w", err)
	}

	return task, nil
}
