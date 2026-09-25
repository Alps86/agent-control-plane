package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainkommentar "agentcontrolplane/app/internal/domain/kommentar"
	"agentcontrolplane/app/internal/domain/rechte"
	portkommentar "agentcontrolplane/app/internal/port/kommentar"
)

// CreateComment prüft Bereich, Archiv und Aufgabenreferenz im selben SQL-Schritt.
func (d *Database) CreateComment(ctx context.Context, actor rechte.Actor, comment domainkommentar.Comment) (domainkommentar.Comment, error) {
	if !actor.Valid() || actor.ID != comment.SourceID || string(actor.Kind) != comment.SourceKind {
		return domainkommentar.Comment{}, portkommentar.ErrNotFound
	}

	refType, refID, refOrg, refName, refLink := d.commentReferenceFields(comment.Reference)
	result, err := d.executor(ctx).ExecContext(ctx, commentInsertSQL, comment.ID, comment.Content,
		comment.SourceKind, comment.SourceID, actor.Kind, actor.ID, refType, refID, refOrg, refName, refLink,
		comment.OrganizationID, comment.TaskID, actor.Kind, actor.ID, actor.Kind, actor.ID,
		refType, refID, comment.OrganizationID)
	if err := d.commentWriteResult(result, err); err != nil {
		return domainkommentar.Comment{}, err
	}

	return d.findComment(ctx, comment.OrganizationID, comment.TaskID, comment.ID)
}

func (d *Database) commentReferenceFields(ref *domainkommentar.Reference) (string, string, string, string, string) {
	if ref == nil {
		return "", "", "", "", ""
	}

	return ref.Type, ref.ID, ref.OrganizationID, ref.DisplayName, ref.Link
}

func (d *Database) commentWriteResult(result sql.Result, err error) error {
	if err != nil {
		return fmt.Errorf("Kommentar speichern: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Kommentar speichern: %w", err)
	}

	if count != 1 {
		return portkommentar.ErrNotFound
	}

	return nil
}

// ListComments verdeckt fremde oder nicht lesbare Threads.
func (d *Database) ListComments(ctx context.Context, actor rechte.Actor, orgID, taskID string) ([]domainkommentar.Comment, error) {
	if !actor.Valid() {
		return nil, portkommentar.ErrNotFound
	}

	rows, err := d.db.QueryContext(ctx, commentListSQL, orgID, taskID,
		actor.Kind, actor.ID, actor.Kind, actor.ID)
	if err != nil {
		return nil, fmt.Errorf("Kommentare lesen: %w", err)
	}

	defer rows.Close()
	return d.collectComments(rows)
}

func (d *Database) collectComments(rows *sql.Rows) ([]domainkommentar.Comment, error) {
	comments := []domainkommentar.Comment{}
	found := false
	for rows.Next() {
		found = true
		comment, err := d.scanComment(rows)
		if err != nil {
			return nil, err
		}

		if comment.ID != "" {
			comments = append(comments, comment)
		}
	}

	if !found && rows.Err() == nil {
		return nil, portkommentar.ErrNotFound
	}

	return comments, rows.Err()
}

// UpdateComment erhält Herkunft, Erstellzeit und Referenz unverändert.
func (d *Database) UpdateComment(ctx context.Context, actor rechte.Actor, orgID, taskID, commentID, content string) (domainkommentar.Comment, error) {
	result, err := d.executor(ctx).ExecContext(ctx, commentUpdateSQL, content, orgID, taskID,
		commentID, actor.Kind, actor.ID, actor.Kind, actor.ID, actor.Kind, actor.ID)
	if err := d.commentWriteResult(result, err); err != nil {
		return domainkommentar.Comment{}, err
	}

	return d.findComment(ctx, orgID, taskID, commentID)
}

func (d *Database) findComment(ctx context.Context, orgID, taskID, commentID string) (domainkommentar.Comment, error) {
	row := d.executor(ctx).QueryRowContext(ctx, `SELECT id, organization_id, task_id, content, source_kind,
		source_id, source_name, created_at, updated_at, reference_type, reference_id,
		reference_organization_id, reference_display_name, reference_link FROM task_comments
		WHERE organization_id = ? AND task_id = ? AND id = ?`, orgID, taskID, commentID)
	return d.scanComment(row)
}

func (d *Database) scanComment(row interface{ Scan(...any) error }) (domainkommentar.Comment, error) {
	var comment domainkommentar.Comment
	var ref domainkommentar.Reference
	err := row.Scan(&comment.ID, &comment.OrganizationID, &comment.TaskID, &comment.Content,
		&comment.SourceKind, &comment.SourceID, &comment.SourceName, &comment.CreatedAt, &comment.UpdatedAt,
		&ref.Type, &ref.ID, &ref.OrganizationID, &ref.DisplayName, &ref.Link)
	if errors.Is(err, sql.ErrNoRows) {
		return domainkommentar.Comment{}, portkommentar.ErrNotFound
	}

	if err != nil {
		return domainkommentar.Comment{}, fmt.Errorf("Kommentar lesen: %w", err)
	}

	if ref.Type != "" {
		comment.Reference = &ref
	}

	return comment, nil
}

// AuditDeniedComment protokolliert eine Verweigerung ohne fremde Ressourcendaten.
func (d *Database) AuditDeniedComment(ctx context.Context, actor rechte.Actor, _ string, taskID, action string) error {
	if actor.Kind != rechte.Agent || actor.ID == "" {
		return portkommentar.ErrNotFound
	}

	result, err := d.executor(ctx).ExecContext(ctx, `INSERT INTO task_comment_denials
		(organization_id, agent_id, task_id, action)
		SELECT a.organization_id, a.id,
			CASE WHEN EXISTS (SELECT 1 FROM tasks t WHERE t.id = ?
				AND t.organization_id = a.organization_id) THEN ? ELSE '' END, ?
		FROM agents a WHERE a.id = ?`, taskID, taskID, action, actor.ID)
	if err != nil {
		return fmt.Errorf("Kommentarverweigerung protokollieren: %w", err)
	}

	return d.commentWriteResult(result, nil)
}
