package sqlite

// CommentMigration ergänzt organisationsgebundene Threads und Verweigerungsaudit.
func CommentMigration() Migration {
	return NewMigration(17,
		`CREATE UNIQUE INDEX IF NOT EXISTS task_comments_tasks_org_id_unique ON tasks (organization_id, id)`,
		`CREATE TABLE task_comments (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			content TEXT NOT NULL CHECK (length(trim(content)) > 0),
			source_kind TEXT NOT NULL CHECK (source_kind IN ('operator', 'agent')),
			source_id TEXT NOT NULL,
			source_name TEXT NOT NULL CHECK (length(trim(source_name)) > 0),
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
			reference_type TEXT NOT NULL DEFAULT '' CHECK (reference_type IN ('', 'task', 'artifact')),
			reference_id TEXT NOT NULL DEFAULT '',
			reference_organization_id TEXT NOT NULL DEFAULT '',
			reference_display_name TEXT NOT NULL DEFAULT '',
			reference_link TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (organization_id, task_id) REFERENCES tasks (organization_id, id),
			CHECK ((reference_type = '' AND reference_id = '' AND reference_organization_id = '') OR
				(reference_type != '' AND reference_id != '' AND reference_organization_id = organization_id))
		)`,
		`CREATE INDEX task_comments_thread_idx ON task_comments (organization_id, task_id, created_at, id)`,
		`CREATE TABLE task_comment_denials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			organization_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			task_id TEXT NOT NULL,
			action TEXT NOT NULL CHECK (action IN ('create', 'read', 'update')),
			occurred_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		)`,
		`CREATE INDEX task_comment_denials_lookup_idx ON task_comment_denials (organization_id, agent_id, id)`)
}
