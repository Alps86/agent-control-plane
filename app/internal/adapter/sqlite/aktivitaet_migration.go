package sqlite

// ActivityMigration speichert fachliche Aufgabenereignisse ab Schemaversion 15.
func ActivityMigration() Migration {
	return NewMigration(15, `CREATE TABLE activity_events (
		sequence INTEGER PRIMARY KEY AUTOINCREMENT,
		id TEXT NOT NULL UNIQUE,
		organization_id TEXT NOT NULL REFERENCES organizations(id),
		project_id TEXT NOT NULL,
		task_id TEXT NOT NULL,
		kind TEXT NOT NULL CHECK (kind IN ('created', 'assigned')),
		actor TEXT NOT NULL CHECK (length(trim(actor)) > 0),
		source TEXT NOT NULL CHECK (source IN ('api', 'browser')),
		occurred_at TEXT NOT NULL,
		object_title TEXT NOT NULL CHECK (length(trim(object_title)) > 0),
		deep_link TEXT NOT NULL CHECK (length(trim(deep_link)) > 0),
		assignee_id TEXT NOT NULL DEFAULT '',
		assignee_name TEXT NOT NULL DEFAULT '',
		CHECK ((kind = 'created' AND assignee_id = '' AND assignee_name = '') OR
			(kind = 'assigned' AND assignee_id != '' AND assignee_name != ''))
	)`, `CREATE INDEX activity_events_org_time_idx
		ON activity_events (organization_id, occurred_at DESC, sequence DESC)`)
}
